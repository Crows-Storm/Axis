package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/Crows-Storm/Axis/common/config/logger"
	consulapi "github.com/hashicorp/consul/api"
)

const (
	// HealthTTL is how long Consul waits for the next TTL update before marking
	// the instance critical. A newly registered TTL check starts critical, so
	// Register must pass the check before it returns.
	HealthTTL = 30 * time.Second
	// HeartbeatInterval must stay at or below one third of HealthTTL. Consul
	// flips the check as soon as the TTL elapses, and the ticker can fire late.
	HeartbeatInterval = 10 * time.Second
	deregisterAfter   = 60 * time.Second
)

type ServiceInfo struct {
	Name string
	ID   string
	Host string
	Port int
}

type Registrar struct {
	client    *consulapi.Client
	service   ServiceInfo
	opTimeout time.Duration
}

func (r *Registrar) GetAddress() string {
	return fmt.Sprintf("%s:%d", r.service.Host, r.service.Port)
}

func (r *Registrar) GetName() string {
	return r.service.Name
}

func NewRegistrar(client *consulapi.Client, svc ServiceInfo, opTimeout time.Duration) *Registrar {
	if opTimeout <= 0 {
		opTimeout = 5 * time.Second
	}
	return &Registrar{client: client, service: svc, opTimeout: opTimeout}
}

func (r *Registrar) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, r.opTimeout)
}

func (r *Registrar) HealthCheck(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	opts := (&consulapi.QueryOptions{}).WithContext(ctx)
	return r.client.Agent().UpdateTTLOpts(r.service.ID, "online", consulapi.HealthPassing, opts)
}

func (r *Registrar) Register(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", r.service.Host, r.service.Port)
	registration := &consulapi.AgentServiceRegistration{
		ID:      r.service.ID,
		Name:    r.service.Name,
		Address: r.service.Host,
		Port:    r.service.Port,
		Check: &consulapi.AgentServiceCheck{
			CheckID:                        r.service.ID,
			TLSSkipVerify:                  false,
			TTL:                            HealthTTL.String(),
			DeregisterCriticalServiceAfter: deregisterAfter.String(),
		},
	}

	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	if err := r.client.Agent().ServiceRegisterOpts(registration, consulapi.ServiceRegisterOpts{}.WithContext(ctx)); err != nil {
		return fmt.Errorf("consul register: %w", err)
	}
	if err := r.HealthCheck(ctx); err != nil {
		if derr := r.Deregister(context.Background()); derr != nil {
			return fmt.Errorf("consul initial ttl: %w (deregister: %v)", err, derr)
		}
		return fmt.Errorf("consul initial ttl: %w", err)
	}
	logger.Info("service registered", "name", r.service.Name, "addr", addr)
	return nil
}

func (r *Registrar) Deregister(ctx context.Context) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()
	logger.Info("deregistering service", "id", r.service.ID)
	opts := (&consulapi.QueryOptions{}).WithContext(ctx)
	return r.client.Agent().ServiceDeregisterOpts(r.service.ID, opts)
}
