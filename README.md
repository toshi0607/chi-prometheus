# chi-prometheus

[![GoDoc](https://pkg.go.dev/badge/github.com/toshi0607/chi-prometheus.svg)](https://pkg.go.dev/github.com/toshi0607/chi-prometheus)
[![License](https://img.shields.io/github/license/toshi0607/chi-prometheus.svg?style=flat-square)](https://github.com/toshi0607/chi-prometheus/blob/master/LICENSE)
[![Release](https://img.shields.io/github/v/release/toshi0607/chi-prometheus?include_prereleases&style=flat-square)](https://github.com/toshi0607/chi-prometheus/releases)
![build](https://github.com/toshi0607/chi-prometheus/actions/workflows/test.yml/badge.svg)
![coverage](coverage.svg)

## Description

chi-prometheus is a prometheus collectors for [go-chi/chi](https://github.com/go-chi/chi).

## Example metrics

```
# HELP chi_request_duration_milliseconds Time spent on the request partitioned by status code, method and HTTP path.
# TYPE chi_request_duration_milliseconds histogram
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/healthz",service="test",le="300"} 1
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/healthz",service="test",le="1200"} 1
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/healthz",service="test",le="5000"} 1
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/healthz",service="test",le="+Inf"} 1
chi_request_duration_milliseconds_sum{code="OK",method="GET",path="/healthz",service="test"} 1
chi_request_duration_milliseconds_count{code="OK",method="GET",path="/healthz",service="test"} 1
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/users/{firstName}",service="test",le="300"} 2
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/users/{firstName}",service="test",le="1200"} 2
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/users/{firstName}",service="test",le="5000"} 2
chi_request_duration_milliseconds_bucket{code="OK",method="GET",path="/users/{firstName}",service="test",le="+Inf"} 2
chi_request_duration_milliseconds_sum{code="OK",method="GET",path="/users/{firstName}",service="test"} 14
chi_request_duration_milliseconds_count{code="OK",method="GET",path="/users/{firstName}",service="test"} 2
# HELP chi_requests_total Number of HTTP requests partitioned by status code, method and HTTP path.
# TYPE chi_requests_total counter
chi_requests_total{code="OK",method="GET",path="/healthz",service="test"} 1
chi_requests_total{code="OK",method="GET",path="/users/{firstName}",service="test"} 2
```

## Usage

chi-prometheus is uses as a middleware. It also supports both a default registry and a custom registry. You can see full examples in [middleware_test.go](middleware_test.go)

### Default registry

```go
    r := chi.NewRouter()
    m := chiprometheus.New("test")
    # use DefaultRegisterer
    m.MustRegisterDefault()
    r.Use(m.Handler)
    r.Handle("/metrics", promhttp.Handler())
    r.Get("/healthz", [YOUR HandlerFunc])
```

### Custom registry

```go
    r := chi.NewRouter()
    # use your registry that works well with promauto
    # see also https://github.com/prometheus/client_golang/issues/716#issuecomment-590282553
    reg := prometheus.NewRegistry()
    if err := reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
        t.Error(err)
    }
    if err := reg.Register(collectors.NewGoCollector()); err != nil {
        t.Error(err)
    }
    m := chiprometheus.New("test")
    reg.MustRegister(m.Collectors()...)
    promh := promhttp.InstrumentMetricHandler(
        reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
    )
    r.Use(m.Handler)
    r.Handle("/metrics", promh)
    r.Get("/healthz", [YOUR HandlerFunc])
```

## Install

```console
$ go get github.com/toshi0607/chi-prometheus
```

## Licence

[MIT](LICENSE) file for details.

## Author

* [GitHub](https://github.com/toshi0607)
* [twitter](https://twitter.com/toshi0607)
