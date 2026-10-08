package server_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/common/discovery/registry"
	"github.com/Crows-Storm/Axis/common/server"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestGRPCListenFailureDoesNotRegister(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = occupied.Close() })
	port := occupied.Addr().(*net.TCPAddr).Port

	var registered atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/agent/service/register" {
			registered.Store(true)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	reg := registry.NewRegistrar(testConsul(t, srv), registry.ServiceInfo{
		Name: "auth",
		ID:   "auth-1",
		Host: "127.0.0.1",
		Port: port,
	}, time.Second)

	errCh := server.RunGRPCServerWithLifecycle(context.Background(), func(*grpc.Server) {}, reg)
	select {
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("expected listen to fail")
	}
	require.False(t, registered.Load())
}

func TestGRPCShutdownDeregistersWithoutDrainingFirst(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())

	deregistered := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/agent/service/deregister/auth-1" {
			select {
			case <-deregistered:
			default:
				close(deregistered)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	reg := registry.NewRegistrar(testConsul(t, srv), registry.ServiceInfo{
		Name: "auth",
		ID:   "auth-1",
		Host: "127.0.0.1",
		Port: port,
	}, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := server.RunGRPCServerWithLifecycle(ctx, func(*grpc.Server) {}, reg)

	start := time.Now()
	cancel()
	select {
	case <-deregistered:
		require.Less(t, time.Since(start), time.Second)
	case <-time.After(2 * time.Second):
		t.Fatal("deregister did not happen before the drain wait")
	}
	select {
	case <-errCh:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not finish")
	}
}

func testConsul(t *testing.T, srv *httptest.Server) *consulapi.Client {
	t.Helper()
	cfg := consulapi.DefaultConfig()
	cfg.Address = srv.Listener.Addr().String()
	cfg.Scheme = "http"
	client, err := consulapi.NewClient(cfg)
	require.NoError(t, err)
	return client
}
