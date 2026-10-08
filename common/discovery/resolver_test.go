package discovery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

func TestBuildDoesNotOverlapBlockingQueries(t *testing.T) {
	var inflight atomic.Int32
	var maxInflight atomic.Int32
	queryDone := make(chan struct{})
	var once sync.Once

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := inflight.Add(1)
		for {
			cur := maxInflight.Load()
			if n <= cur || maxInflight.CompareAndSwap(cur, n) {
				break
			}
		}
		defer inflight.Add(-1)

		select {
		case <-r.Context().Done():
			once.Do(func() { close(queryDone) })
		case <-time.After(20 * time.Second):
			http.Error(w, "query was not cancelled", http.StatusGatewayTimeout)
		}
	}))
	t.Cleanup(srv.Close)

	client := testConsul(t, srv)
	cc := &recordConn{}
	buildDone := make(chan struct{})
	var built resolver.Resolver
	go func() {
		var err error
		built, err = NewBuilder(client).Build(userTarget(), cc, resolver.BuildOptions{})
		require.NoError(t, err)
		close(buildDone)
	}()

	select {
	case <-buildDone:
	case <-time.After(time.Second):
		t.Fatal("Build blocked on the Consul blocking query")
	}

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		if got := maxInflight.Load(); got > 1 {
			t.Fatalf("overlapping consul queries: %d", got)
		}
		time.Sleep(50 * time.Millisecond)
	}

	built.Close()
	select {
	case <-queryDone:
	case <-time.After(time.Second):
		t.Fatal("Close did not cancel the in-flight Consul query")
	}
}

func TestResolveAdoptsDecreasedConsulIndex(t *testing.T) {
	var mu sync.Mutex
	var indexes []string
	body := []byte(`[{"Node":{"Node":"local","Address":"10.0.0.1"},"Service":{"ID":"user-1","Service":"user","Address":"127.0.0.1","Port":9002},"Checks":[]}]`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := r.URL.Query().Get("index")
		mu.Lock()
		indexes = append(indexes, idx)
		n := len(indexes)
		mu.Unlock()

		switch n {
		case 1:
			w.Header().Set("X-Consul-Index", "100")
			_, _ = w.Write(body)
		case 2:
			w.Header().Set("X-Consul-Index", "5")
			_, _ = w.Write(body)
		default:
			select {
			case <-r.Context().Done():
			case <-time.After(5 * time.Second):
			}
		}
	}))
	t.Cleanup(srv.Close)

	built, err := NewBuilder(testConsul(t, srv)).Build(userTarget(), &recordConn{}, resolver.BuildOptions{})
	require.NoError(t, err)
	t.Cleanup(func() { built.Close() })

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(indexes) >= 3 && indexes[0] == "" && indexes[1] == "100" && indexes[2] == "5"
	}, 3*time.Second, 20*time.Millisecond)
}

func TestResolvePublishesHealthyEndpoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("index") != "" {
			select {
			case <-r.Context().Done():
			case <-time.After(5 * time.Second):
			}
			return
		}
		w.Header().Set("X-Consul-Index", "1")
		body, err := json.Marshal([]map[string]any{
			{
				"Node": map[string]any{"Node": "local", "Address": "10.0.0.1"},
				"Service": map[string]any{
					"ID":      "user-1",
					"Service": "user",
					"Address": "127.0.0.1",
					"Port":    9002,
				},
				"Checks": []any{},
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	cc := &recordConn{}
	built, err := NewBuilder(testConsul(t, srv)).Build(userTarget(), cc, resolver.BuildOptions{})
	require.NoError(t, err)
	t.Cleanup(func() { built.Close() })

	require.Eventually(t, func() bool {
		states := cc.states()
		return len(states) == 1 && len(states[0].Addresses) == 1 && states[0].Addresses[0].Addr == "127.0.0.1:9002"
	}, 2*time.Second, 20*time.Millisecond)
}

type recordConn struct {
	mu      sync.Mutex
	updates []resolver.State
}

func (c *recordConn) UpdateState(state resolver.State) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updates = append(c.updates, state)
	return nil
}

func (c *recordConn) ReportError(error) {}

func (c *recordConn) NewAddress([]resolver.Address) {}

func (c *recordConn) ParseServiceConfig(string) *serviceconfig.ParseResult { return nil }

func (c *recordConn) states() []resolver.State {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]resolver.State, len(c.updates))
	copy(out, c.updates)
	return out
}

func userTarget() resolver.Target {
	return resolver.Target{URL: url.URL{Scheme: Scheme, Path: "/user"}}
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
