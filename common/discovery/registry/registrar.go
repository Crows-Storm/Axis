package registry

import (
	"context"
	"fmt"

	"github.com/Crows-Storm/Axis/common/config/logger"
	consulapi "github.com/hashicorp/consul/api"
)

type ServiceInfo struct {
	Name string
	ID   string
	Host string
	Port int
}

type Registrar struct {
	client  *consulapi.Client
	service ServiceInfo
}

func (r *Registrar) GetAddress() string {
	return fmt.Sprintf("%s:%d", r.service.Host, r.service.Port)
}

func (r *Registrar) GetName() string {
	return r.service.Name
}

func NewRegistrar(client *consulapi.Client, svc ServiceInfo) *Registrar {
	return &Registrar{client: client, service: svc}
}

func (r *Registrar) HealthCheck() error {
	return r.client.Agent().UpdateTTL(r.service.ID, "online", consulapi.HealthPassing)
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
			TTL:                            "10s",
			DeregisterCriticalServiceAfter: "60s",
		},
	}

	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("consul register: %w", err)
	}
	logger.Info("service registered", "name", r.service.Name, "addr", addr)
	return nil
}

func (r *Registrar) Deregister() error {
	logger.Info("deregistering service", "id", r.service.ID)
	return r.client.Agent().ServiceDeregister(r.service.ID)
}
