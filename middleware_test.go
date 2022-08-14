package chiprometheus_test

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	chiprometheus "github.com/toshi0607/chi-prometheus"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const testHost = "http://localhost"

func TestMiddleware_MustRegisterDefault(t *testing.T) {
	t.Run("without collectors", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("must have panicked")
			}
		}()
		m := chiprometheus.Middleware{}
		m.MustRegisterDefault()
	})

	t.Run("with collectors", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("must not have panicked")
			}
		}()
		m := chiprometheus.New("test")
		m.MustRegisterDefault()
		t.Cleanup(func() {
			for _, c := range m.Collectors() {
				prometheus.Unregister(c)
			}
		})
	})
}

func TestMiddleware_Collectors(t *testing.T) {
	m := chiprometheus.New("test")
	want := 2
	got := len(m.Collectors())
	if got != want {
		t.Errorf("number of collectors should be %v, got %v", want, got)
	}
}

func testHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}
}

func makeRequest(t *testing.T, r *chi.Mux, paths [][]string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	for _, p := range paths {
		u, err := url.JoinPath(testHost, p...)
		if err != nil {
			t.Error(err)
		}
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			t.Error(err)
		}
		r.ServeHTTP(rec, req)
	}
	return rec.Body.String()
}

func TestMiddleware_Handler(t *testing.T) {
	tests := map[string]struct {
		body string
		want bool
	}{
		"request header": {chiprometheus.RequestsCollectorName, true},
		"latency header": {chiprometheus.LatencyCollectorName, true},
		"path variable":  {`chi_request_duration_milliseconds_count{code="OK",method="GET",path="/users/{firstName}",service="test"} 2`, true},
		// specific path values should be omitted
		"bob":   {`chi_request_duration_milliseconds_count{code="OK",method="GET",path="/users/bob",service="test"} 1`, false},
		"alice": {`chi_request_duration_milliseconds_count{code="OK",method="GET",path="/users/alice",service="test"} 1`, false},
	}

	r := chi.NewRouter()
	m := chiprometheus.New("test")
	m.MustRegisterDefault()
	t.Cleanup(func() {
		for _, c := range m.Collectors() {
			prometheus.Unregister(c)
		}
	})
	r.Use(m.Handler)
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/healthz", testHandler(t))
	r.Get("/users/{firstName}", testHandler(t))
	paths := [][]string{
		{"healthz"},
		{"users", "bob"},
		{"users", "alice"},
		{"metrics"},
	}
	got := makeRequest(t, r, paths)

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tt.want && !strings.Contains(got, tt.body) {
				t.Fatalf("body should contain %s", tt.body)
			} else if !tt.want && strings.Contains(got, tt.body) {
				t.Fatalf("body should NOT contain %s", tt.body)
			}
		})
	}
}

func TestMiddleware_HandlerWithCustomRegistry(t *testing.T) {
	tests := map[string]struct {
		want string
	}{
		"request":    {chiprometheus.RequestsCollectorName},
		"latency":    {chiprometheus.LatencyCollectorName},
		"process":    {"promhttp_metric_handler_requests_total"},
		"go runtime": {"go_goroutines"},
	}

	r := chi.NewRouter()
	reg := prometheus.NewRegistry()
	if err := reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		t.Error(err)
	}
	if err := reg.Register(collectors.NewGoCollector()); err != nil {
		t.Error(err)
	}
	m := chiprometheus.New("test")
	reg.MustRegister(m.Collectors()...)
	t.Cleanup(func() {
		for _, c := range m.Collectors() {
			prometheus.Unregister(c)
		}
	})
	promh := promhttp.InstrumentMetricHandler(
		reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	)
	r.Use(m.Handler)
	r.Handle("/metrics", promh)
	r.Get("/healthz", testHandler(t))
	paths := [][]string{
		{"healthz"},
		{"metrics"},
	}
	got := makeRequest(t, r, paths)

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(got, tt.want) {
				t.Fatalf("body should contain %s", tt.want)
			}
		})
	}
}

func TestMiddleware_HandlerWithBucketEnv(t *testing.T) {
	key := chiprometheus.EnvChiPrometheusLatencyBuckets

	t.Run("with invalid env", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("must have panicked")
			}
		}()
		if err := os.Setenv(key, "invalid value"); err != nil {
			t.Fatalf("failed to set %s", key)
		}
		t.Cleanup(func() { _ = os.Unsetenv(key) })
		chiprometheus.New("test")
	})

	t.Run("with valid env", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("must not have panicked")
			}
		}()

		tests := map[string]struct {
			body string
			want bool
		}{
			"le 101":  {`path="/healthz",service="test",le="101"`, true},
			"le 201":  {`path="/healthz",service="test",le="201"`, true},
			"le +Inf": {`path="/healthz",service="test",le="+Inf"`, true},
			// default values should be overwritten
			"le 300":  {`path="/healthz",service="test",le="300"`, false},
			"le 1200": {`path="/healthz",service="test",le="1200"`, false},
			"le 5000": {`path="/healthz",service="test",le="1200"`, false},
		}

		if err := os.Setenv(key, "101,201"); err != nil {
			t.Fatalf("failed to set %s", key)
		}
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		r := chi.NewRouter()
		m := chiprometheus.New("test")
		m.MustRegisterDefault()
		t.Cleanup(func() {
			for _, c := range m.Collectors() {
				prometheus.Unregister(c)
			}
		})
		r.Use(m.Handler)
		r.Handle("/metrics", promhttp.Handler())
		r.Get("/healthz", testHandler(t))
		paths := [][]string{
			{"healthz"},
			{"metrics"},
		}
		got := makeRequest(t, r, paths)

		for name, tt := range tests {
			tt := tt
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				if tt.want && !strings.Contains(got, tt.body) {
					t.Fatalf("body should contain %s", tt.body)
				} else if !tt.want && strings.Contains(got, tt.body) {
					t.Fatalf("body should NOT contain %s", tt.body)
				}
			})
		}
	})
}
