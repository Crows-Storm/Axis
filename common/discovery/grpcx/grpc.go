package grpcx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/discovery"
	"github.com/Crows-Storm/Axis/common/discovery/consulx"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
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

// DialService obtains a connection by service name (automatic reuse to prevent concurrent duplicate creation)
func DialService(serviceName string) (*grpc.ClientConn, error) {
	if err := InitClientInfrastructure(); err != nil {
		return nil, err
	}

	// Fast path: Return immediately upon read lock hit.
	poolMu.RLock()
	if conn, ok := connPool[serviceName]; ok {
		poolMu.RUnlock()
		return conn, nil
	}
	poolMu.RUnlock()

	// Slow path: Use singleflight to merge concurrent requests with the same serviceName to avoid creating 100 connections when 100 goroutines dial the same service at the same time.
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
			grpc.WithChainUnaryInterceptor(metadataPropagationInterceptor()),
		}

		conn, err := grpc.NewClient(target, opts...)
		if err != nil {
			return nil, fmt.Errorf("dial %s: %w", serviceName, err)
		}

		// Write to the connection pool (singleflight ensures that only one goroutine executes to this point for the same key)
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

// metadataPropagationInterceptor Automatically transmit the metadata from upstream to downstream.
func metadataPropagationInterceptor() grpc.UnaryClientInterceptor {
	allowedKeys := map[string]bool{
		"trace-id":     true,
		"auth-token":   true,
		"x-request-id": true,
	}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			outMD := make(metadata.MD, len(md))
			for k, vals := range md {
				if allowedKeys[k] {
					outMD[k] = append([]string(nil), vals...) // Deep copy prevents sharing of underlying arrays
				}
			}
			if len(outMD) > 0 {
				ctx = metadata.NewOutgoingContext(ctx, outMD)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
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
