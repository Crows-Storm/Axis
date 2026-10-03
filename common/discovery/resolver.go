package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

const Scheme = "consul"

const (
	blockingWait = 5 * time.Minute
	minBackoff   = time.Second
	maxBackoff   = 5 * time.Second
)

type Builder struct {
	client *consulapi.Client
}

func NewBuilder(client *consulapi.Client) *Builder {
	b := &Builder{client: client}
	resolver.Register(b)
	return b
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	serviceName := target.Endpoint()
	if serviceName == "" {
		return nil, fmt.Errorf("consul resolver: empty service name")
	}
	ctx, cancel := context.WithCancel(context.Background())

	r := &consulResolver{
		client:      b.client,
		serviceName: serviceName,
		cc:          cc,
		ctx:         ctx,
		cancel:      cancel,
	}

	go r.watch()
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
	emptyNoted  bool
}

func (r *consulResolver) watch() {
	var backoff time.Duration
	for r.ctx.Err() == nil {
		err := r.resolve()
		if r.ctx.Err() != nil {
			return
		}
		if err == nil {
			backoff = 0
			continue
		}
		if backoff == 0 {
			backoff = minBackoff
		} else if backoff < maxBackoff {
			backoff *= 2
		}
		timer := time.NewTimer(backoff)
		select {
		case <-r.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (r *consulResolver) resolve() error {
	opts := (&consulapi.QueryOptions{
		WaitIndex: r.lastIndex,
		WaitTime:  blockingWait,
	}).WithContext(r.ctx)

	services, meta, err := r.client.Health().Service(r.serviceName, "", true, opts)
	if err != nil {
		if r.ctx.Err() != nil {
			return nil
		}
		logger.WithField("service", r.serviceName).WithError(err).Error("consul resolve failed")
		r.cc.ReportError(err)
		return err
	}
	if meta == nil || meta.LastIndex == 0 {
		err := fmt.Errorf("consul: missing index for %s", r.serviceName)
		logger.WithField("service", r.serviceName).WithError(err).Error("consul resolve failed")
		r.cc.ReportError(err)
		return err
	}
	// Consul resets the index on restart. A lower index must replace the old
	// WaitIndex, or the next blocking query waits until the old value is passed.
	r.lastIndex = meta.LastIndex

	addrs := endpointsFrom(services)
	if len(addrs) == 0 {
		if !r.emptyNoted {
			logger.WithField("service", r.serviceName).Warn("consul resolver has no healthy endpoints")
			r.emptyNoted = true
		}
		// An empty list is a valid "nobody is passing" result. gRPC answers it
		// with ErrBadResolverState so the channel drops stale backends.
		if err := r.cc.UpdateState(resolver.State{}); err != nil && !errors.Is(err, balancer.ErrBadResolverState) {
			logger.WithField("service", r.serviceName).WithError(err).Error("update resolver state failed")
			return err
		}
		return nil
	}

	r.emptyNoted = false
	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		logger.WithField("service", r.serviceName).WithError(err).Error("update resolver state failed")
		return err
	}
	return nil
}

func endpointsFrom(services []*consulapi.ServiceEntry) []resolver.Address {
	addrs := make([]resolver.Address, 0, len(services))
	for _, svc := range services {
		if svc == nil || svc.Service == nil || svc.Service.Port <= 0 {
			continue
		}
		host := svc.Service.Address
		if host == "" && svc.Node != nil {
			host = svc.Node.Address
		}
		if host == "" {
			continue
		}
		addrs = append(addrs, resolver.Address{Addr: fmt.Sprintf("%s:%d", host, svc.Service.Port)})
	}
	return addrs
}

func (r *consulResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *consulResolver) Close() {
	r.cancel()
}
