package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

const Scheme = "consul"

type Builder struct {
	client *consulapi.Client
}

func NewBuilder(client *consulapi.Client) *Builder {
	b := &Builder{client: client}
	resolver.Register(b)
	return b
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	serviceName := target.Endpoint()
	ctx, cancel := context.WithCancel(context.Background())

	r := &consulResolver{
		client:      b.client,
		serviceName: serviceName,
		cc:          cc,
		ctx:         ctx,
		cancel:      cancel,
	}

	go r.watch()
	r.resolve()
	return r, nil
}

func (b *Builder) Scheme() string { return Scheme }

type consulResolver struct {
	client      *consulapi.Client
	serviceName string
	cc          resolver.ClientConn
	ctx         context.Context
	cancel      context.CancelFunc
	lastIndex   uint64
}

func (r *consulResolver) watch() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.resolve()
		}
	}
}

func (r *consulResolver) resolve() {
	opts := &consulapi.QueryOptions{
		WaitIndex: r.lastIndex,
		WaitTime:  5 * time.Minute, // Blocking Query long poll
	}

	services, meta, err := r.client.Health().Service(r.serviceName, "", true, opts)
	if err != nil {
		if r.ctx.Err() != nil {
			return
		} // context cancel Exit directly
		logger.Error("consul resolve failed",
			"service", r.serviceName,
			"lastIndex", r.lastIndex,
			"waitTime", opts.WaitTime,
			"error", err)
		r.cc.ReportError(err)
		time.Sleep(3 * time.Second) // retry
		return
	}

	r.lastIndex = meta.LastIndex
	addrs := make([]resolver.Address, 0, len(services))
	for _, svc := range services {
		addr := fmt.Sprintf("%s:%d", svc.Service.Address, svc.Service.Port)
		addrs = append(addrs, resolver.Address{Addr: addr})
	}

	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		logger.Error("update resolver state failed", "error", err)
	}
}

func (r *consulResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *consulResolver) Close() {
	r.cancel() // Notification watch goroutine exit
}
