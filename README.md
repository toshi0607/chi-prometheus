# chi-prometheus

[![GoDoc](https://pkg.go.dev/badge/github.com/toshi0607/chi-prometheus.svg)](https://pkg.go.dev/github.com/toshi0607/chi-prometheus)
[![License](https://img.shields.io/github/license/toshi0607/chi-prometheus.svg?style=flat-square)](https://github.com/toshi0607/chi-prometheus/blob/master/LICENSE)
[![Release](https://img.shields.io/github/v/release/toshi0607/chi-prometheus?include_prereleases&style=flat-square)](https://github.com/toshi0607/chi-prometheus/releases)
![build](https://github.com/toshi0607/chi-prometheus/actions/workflows/test.yml/badge.svg)
![coverage](coverage.svg)

## Description

chi-prometheus is a Prometheus collectors for [go-chi/chi](https://github.com/go-chi/chi).

## Example metrics

```
# HELP chi_request_duration_milliseconds Time spent on the request partitioned by status code, method and HTTP path.
# TYPE chi_request_duration_milliseconds histogram
chi_request_duration_milliseconds_bucket{code="200",method="GET",path="/users/{firstName}",service="test",le="300"} 2
chi_request_duration_milliseconds_bucket{code="200",method="GET",path="/users/{firstName}",service="test",le="1200"} 2
chi_request_duration_milliseconds_bucket{code="200",method="GET",path="/users/{firstName}",service="test",le="5000"} 2
chi_request_duration_milliseconds_bucket{code="200",method="GET",path="/users/{firstName}",service="test",le="+Inf"} 2
chi_request_duration_milliseconds_sum{code="200",method="GET",path="/users/{firstName}",service="test"} 4
chi_request_duration_milliseconds_count{code="200",method="GET",path="/users/{firstName}",service="test"} 2
chi_request_duration_milliseconds_bucket{code="404",method="GET",path="/healthz",service="test",le="300"} 0
chi_request_duration_milliseconds_bucket{code="404",method="GET",path="/healthz",service="test",le="1200"} 0
chi_request_duration_milliseconds_bucket{code="404",method="GET",path="/healthz",service="test",le="5000"} 0
chi_request_duration_milliseconds_bucket{code="404",method="GET",path="/healthz",service="test",le="+Inf"} 1
chi_request_duration_milliseconds_sum{code="404",method="GET",path="/healthz",service="test"} 9339
chi_request_duration_milliseconds_count{code="404",method="GET",path="/healthz",service="test"} 1
# HELP chi_requests_total Number of HTTP requests partitioned by status code, method and HTTP path.
# TYPE chi_requests_total counter
chi_requests_total{code="200",method="GET",path="/users/{firstName}",service="test"} 2
chi_requests_total{code="404",method="GET",path="/healthz",service="test"} 1
```

## Requirement

chi-prometheus only works with [go-chi/chi](https://github.com/go-chi/chi). Though you can set the middleware for other routers (multiplexers) including [http.ServeMux](https://pkg.go.dev/net/http#ServeMux), nothing happens.

## Usage

chi-prometheus is used as a middleware. It also supports both a default registry and a custom registry. You can see full examples in [middleware_test.go](middleware_test.go)

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

### Configuration

Latency histogram bucket is configurable with `CHI_PROMETHEUS_LATENCY_BUCKETS`. Default values are `300, 1200, 5000` (milliseconds).

You can override them as follows;

```shell
# comma separated string value
CHI_PROMETHEUS_LATENCY_BUCKETS="100,200,300,400"
```

## Install

```console
$ go get github.com/toshi0607/chi-prometheus
```

## Comparison

There are great prior implementations for similar use cases.

### [zbindenren/negroni-prometheus](https://github.com/zbindenren/negroni-prometheus)

This is a router-agnostic Prometheus middleware implementation for an HTTP server. It is not actively developed and is archived by the owner.

### [766b/chi-prometheus](https://github.com/766b/chi-prometheus)

This is a Prometheus middleware taking advantage of a [chi middleware](https://github.com/go-chi/chi/tree/master/middleware) library for chi. It is not actively developed, and I had to newly support the following features.

- Go Modules
- Custom registry
- New features of [prometheus/client_golang](https://github.com/prometheus/client_golang)
- Deleting the previous default codes (non-aggregation collector for paths, such as `/users/{firstName}`) 

So I decided to implement a new one instead of forking and adding some tweaks.

## Licence

[MIT](LICENSE) file for details.

## Author

* [GitHub](https://github.com/toshi0607)
* [twitter](https://twitter.com/toshi0607)
