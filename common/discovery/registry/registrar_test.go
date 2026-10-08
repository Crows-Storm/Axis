package registry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
)

func TestRegisterMarksPassingBeforeReturn(t *testing.T) {
	var mu sync.Mutex
	var registration consulapi.AgentServiceRegistration
	var ttlStatus string
	var deregistered bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/v1/agent/service/register":
			if err := json.Unmarshal(body, &registration); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		case "/v1/agent/check/update/auth-1":
			var update struct {
				Status string
			}
			if err := json.Unmarshal(body, &update); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ttlStatus = update.Status
		case "/v1/agent/service/deregister/auth-1":
			deregistered = true
		default:
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := testConsul(t, srv)
	reg := NewRegistrar(client, ServiceInfo{
		Name: "auth",
		ID:   "auth-1",
		Host: "127.0.0.1",
		Port: 9001,
	}, time.Second)

	require.NoError(t, reg.Register(context.Background()))

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, consulapi.HealthPassing, ttlStatus)
	ttl, err := time.ParseDuration(registration.Check.TTL)
	require.NoError(t, err)
	require.GreaterOrEqual(t, ttl, 3*HeartbeatInterval)
	require.False(t, deregistered)
}

func TestRegisterDeregistersWhenInitialTTLFails(t *testing.T) {
	var mu sync.Mutex
	deregistered := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/v1/agent/service/register":
			w.WriteHeader(http.StatusOK)
		case "/v1/agent/check/update/auth-1":
			http.Error(w, "ttl failed", http.StatusInternalServerError)
		case "/v1/agent/service/deregister/auth-1":
			deregistered = true
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	client := testConsul(t, srv)
	reg := NewRegistrar(client, ServiceInfo{Name: "auth", ID: "auth-1", Host: "127.0.0.1", Port: 9001}, time.Second)

	err := reg.Register(context.Background())
	require.Error(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.True(t, deregistered)
}

func TestHealthCheckStopsWhenConsulHangs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	reg := NewRegistrar(testConsul(t, srv), ServiceInfo{
		Name: "auth",
		ID:   "auth-1",
		Host: "127.0.0.1",
		Port: 9001,
	}, 200*time.Millisecond)

	start := time.Now()
	err := reg.HealthCheck(context.Background())
	require.Error(t, err)
	require.Less(t, time.Since(start), time.Second)
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
