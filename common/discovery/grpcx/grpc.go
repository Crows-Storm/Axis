package grpcx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/discovery"
	"github.com/Crows-Storm/Axis/common/discovery/consulx"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var (
	initOnce sync.Once
	initErr  error

	connPool  = make(map[string]*grpc.ClientConn)
	poolMu    sync.RWMutex
	dialGroup singleflight.Group
)

func InitClientInfrastructure() error {
	initOnce.Do(func() {
		cfg := config.Get()
		consulAddr := fmt.Sprintf("%s:%d", cfg.ServiceDiscoveryConfig.Host, cfg.ServiceDiscoveryConfig.Port)

		client, err := consulx.NewClient(&consulx.Config{
			Address: consulAddr,
			Token:   cfg.ServiceDiscoveryConfig.ACTToken,
			Timeout: cfg.ServiceDiscoveryConfig.Timeout,
		})
		if err != nil {
			initErr = err
			return
		}
		discovery.NewBuilder(client)
	})
	return initErr
}

func DialService(serviceName string) (*grpc.ClientConn, error) {
	if err := InitClientInfrastructure(); err != nil {
		return nil, err
	}

	poolMu.RLock()
	if conn, ok := connPool[serviceName]; ok {
		poolMu.RUnlock()
		return conn, nil
	}
	poolMu.RUnlock()

	v, err, _ := dialGroup.Do(serviceName, func() (any, error) {
		poolMu.RLock()
		if conn, ok := connPool[serviceName]; ok {
			poolMu.RUnlock()
			return conn, nil
		}
		poolMu.RUnlock()

		target := fmt.Sprintf("%s:///%s", discovery.Scheme, serviceName)

		serviceConfig := `{
			"loadBalancingConfig": [{"round_robin":{}}],
			"methodConfig": [{
				"name": [{}],
				"waitForReady": true,
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.1s",
					"maxBackoff": "1s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE"]
				},
				"timeout": "10s"
			}]
		}`

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(serviceConfig),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			}),
			grpc.WithChainUnaryInterceptor(contextx.UnaryClientInterceptor()),
		}

		conn, err := grpc.NewClient(target, opts...)
		if err != nil {
			return nil, fmt.Errorf("dial %s: %w", serviceName, err)
		}

		poolMu.Lock()
		connPool[serviceName] = conn
		defer poolMu.Unlock()

		return conn, nil
	})

	if err != nil {
		return nil, err
	}
	return v.(*grpc.ClientConn), nil
}

func CloseAll(ctx context.Context) error {
	poolMu.Lock()
	defer poolMu.Unlock()

	var errs []error
	for name, conn := range connPool {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", name, err))
		}
		delete(connPool, name)
	}

	if len(errs) > 0 {
		return fmt.Errorf("close connections: %v", errs)
	}
	return nil
}
