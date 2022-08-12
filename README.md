# chi-prometheus
![build](https://github.com/toshi0607/chi-prometheus/actions/workflows/test.yml/badge.svg) ![coverage](coverage.svg)

## Description

chi-prometheus is a prometheus collectors for [go-chi/chi](https://github.com/go-chi/chi).

## Usage

chi-prometheus is uses as a middleware. It also supports both a default registry and a custom registry. You can see full examples in [middleware_test.go](middleware_test.go)

### Default registry

```go
    r := chi.NewRouter()
    m := chiprometheus.New("test")
    m.MustRegisterDefault()
    r.Use(m.Handler)
    r.Handle("/metrics", promhttp.Handler())
    r.Get(`/healthz`, [YOUR HandlerFunc])
```

### Custom registry

```go
    r := chi.NewRouter()
    reg := prometheus.NewRegistry()
    if err := reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
        t.Error(err)
    }
    if err := reg.Register(collectors.NewGoCollector()); err != nil {
        t.Error(err)
    }
    m := chiprometheus.New("test")
    for _, c := range m.Collectors() {
        if err := reg.Register(c); err != nil {
            return err
        }
    }
    promh := promhttp.InstrumentMetricHandler(
        reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
    )
    r.Use(m.Handler)
    r.Handle("/metrics", promh)
    r.Get(`/healthz`, [YOUR HandlerFunc])
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
