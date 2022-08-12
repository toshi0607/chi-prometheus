package chiprometheus_test

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	t.Run("must panic without collectors", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("must have panicked")
			}
		}()
		m := chiprometheus.Middleware{}
		m.MustRegisterDefault()
	})

	t.Run("must not panic with collectors", func(t *testing.T) {
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

func testHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	w.WriteHeader(http.StatusOK)
}

func TestMiddleware_Handler(t *testing.T) {
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
	r.Get("/healthz", testHandler)
	r.Get("/users/{firstName}", testHandler)

	paths := [][]string{
		{"healthz"},
		{"users", "bob"},
		{"users", "alice"},
		{"metrics"},
	}
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
	body := rec.Body.String()

	if !strings.Contains(body, chiprometheus.RequestsCollectorName) {
		t.Errorf("body should contain request total entry '%s'", chiprometheus.RequestsCollectorName)
	}
	if !strings.Contains(body, chiprometheus.LatencyCollectorName) {
		t.Errorf("body should contain request duration entry '%s'", chiprometheus.LatencyCollectorName)
	}

	healthzCount := `chi_request_duration_milliseconds_count{code="OK",method="GET",path="/healthz",service="test"} 1`
	bobCount := `chi_request_duration_milliseconds_count{code="OK",method="GET",path="/users/bob",service="test"} 1`
	aliceCount := `chi_request_duration_milliseconds_count{code="OK",method="GET",path="/users/alice",service="test"} 1`
	aggregatedCount := `chi_request_duration_milliseconds_count{code="OK",method="GET",path="/users/{firstName}",service="test"} 2`
	if !strings.Contains(body, healthzCount) {
		t.Errorf("body should contain healthz count summary '%s'", healthzCount)
	}
	if strings.Contains(body, bobCount) {
		t.Errorf("body should NOT contain Bob count summary '%s'", bobCount)
	}
	if strings.Contains(body, aliceCount) {
		t.Errorf("body should NOT contain Alice count summary '%s'", aliceCount)
	}
	if !strings.Contains(body, aggregatedCount) {
		t.Errorf("body should contain first name count summary '%s'", aggregatedCount)
	}
}

func TestMiddleware_HandlerWithCustomRegistry(t *testing.T) {
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
	r.Get("/healthz", testHandler)

	paths := [][]string{
		{"healthz"},
		{"metrics"},
	}
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
	body := rec.Body.String()

	if !strings.Contains(body, chiprometheus.RequestsCollectorName) {
		t.Errorf("body should contain request total entry '%s'", chiprometheus.RequestsCollectorName)
	}
	if !strings.Contains(body, chiprometheus.LatencyCollectorName) {
		t.Errorf("body should contain request duration entry '%s'", chiprometheus.LatencyCollectorName)
	}

	if !strings.Contains(body, "promhttp_metric_handler_requests_total") {
		t.Error("body should contain promhttp_metric_handler_requests_total from ProcessCollector")
	}
	if !strings.Contains(body, "go_goroutines") {
		t.Errorf("body should contain Go runtime metrics from GoCollector")
	}
}
