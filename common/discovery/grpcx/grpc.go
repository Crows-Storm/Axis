package grpcx

import (
	"context"
	"fmt"
	"strings"
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

// metadataPropagationInterceptor Automatically transmits metadata from upstream to downstream.
//
// It handles two propagation sources and merges them into the outgoing context:
//  1. Outbound metadata attached by HTTP middleware via WithOutgoingMetadata
//     (bridges HTTP request headers into the gRPC call — the primary path when
//     this service is the HTTP entry point).
//  2. Incoming metadata from an upstream gRPC call (for gRPC→gRPC chains).
//
// Only keys in allowedPropagationKeys are forwarded, and all keys are
// normalized to lowercase (gRPC metadata keys are lowercase on the wire).
func metadataPropagationInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		outMD := make(metadata.MD)

		// Source: metadata bridged from HTTP middleware
		if md, ok := OutgoingMetadataFromContext(ctx); ok {
			appendAllowed(outMD, md)
		}

		// Source: metadata from upstream gRPC call
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			appendAllowed(outMD, md)
		}

		if len(outMD) > 0 {
			ctx = metadata.NewOutgoingContext(ctx, outMD)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// appendAllowed copies only allow-listed keys from src into dst, normalizing keys to lowercase.
func appendAllowed(dst, src metadata.MD) {
	for k, vals := range src {
		lower := strings.ToLower(k)
		if !allowedPropagationKeys[lower] {
			continue
		}
		dst[lower] = append(dst[lower], append([]string(nil), vals...)...)
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
