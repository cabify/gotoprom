# gotoprom
## A Prometheus metrics builder

[![Travis CI build status](https://travis-ci.com/cabify/gotoprom.svg?branch=master)](https://travis-ci.com/cabify/gotoprom)
[![Coverage Status](https://coveralls.io/repos/github/cabify/gotoprom/badge.svg)](https://coveralls.io/github/cabify/gotoprom)
[![GoDoc](https://godoc.org/github.com/cabify/gotoprom?status.svg)](https://godoc.org/github.com/cabify/gotoprom)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  

`gotoprom` offers an easy to use declarative API with type-safe labels for building and using Prometheus metrics.
It doesn't replace the [official Prometheus client](https://github.com/prometheus/client_golang)
but adds a wrapper on top of it.

`gotoprom` is built for developers who like type safety, navigating the code using IDEs and using a “find usages”
functionality, making refactoring and debugging easier at the cost of performance and writing slightly more verbose code.


## Motivation

Main motivation for this library was to have type-safety on the Prometheus labels, which are
just a `map[string]string` in the original library, and their values can be reported even
without mentioning the label name, just relying on the order they were declared in.

For example, it replaces:
```go
httpReqs := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
    },
    []string{"code", "method"},
)
prometheus.MustRegister(httpReqs)

// ...

httpReqs.WithLabelValues("404", "POST").Add(42)
```

With:
```go
var metrics = struct{
	Reqs func(labels) prometheus.Counter `name:"requests_total" help:"How many HTTP requests processed, partitioned by status code and HTTP method."`
}

type labels struct {
	Code   int    `label:"code"`
	Method string `label:"method"`
}

gotoprom.MustInit(&metrics, "http")

// ...

metrics.Reqs(labels{Code: 404, Method: "POST"}).Inc()
```

This way it's impossible to mess the call by exchanging the order of `"POST"` & `"404"` params.


## Usage

Define your metrics:

```go
var metrics struct {
	SomeCounter                      func() prometheus.Counter   `name:"some_counter" help:"some counter"`
	SomeHistogram                    func() prometheus.Histogram `name:"some_histogram" help:"Some histogram with default buckets"`
	SomeHistogramWithSpecificBuckets func() prometheus.Histogram `name:"some_histogram_with_buckets" help:"Some histogram with custom buckets" buckets:".01,.05,.1"`
	SomeGauge                        func() prometheus.Gauge     `name:"some_gauge" help:"Some gauge"`
	SomeSummaryWithSpecificMaxAge    func() prometheus.Summary   `name:"some_summary_with_specific_max_age" help:"Some summary with custom max age" max_age:"20m"`

	Requests struct {
		Total func(requestLabels) prometheus.Count `name:"total" help:"Total amount of requests served"`
	} `namespace:"requests"`
}

type requestLabels struct {
	Service    string `label:"service"`
	StatusCode int    `label:"status"`
	Success    bool   `label:"success"`
}
```

Initialize them:

```go
func init() {
	gotoprom.MustInit(&metrics, "namespace")
}
```

Measure stuff:

```go
metrics.SomeGauge().Set(100)
metrics.Requests.Total(requestLabels{Service: "google", StatusCode: 404, Success: false}).Inc()
```


## Custom metric types

By default, only some basic metric types are registered when `gotoprom` is intialized:
* `prometheus.Counter`
* `prometheus.Histogram`
* `prometheus.Gauge`
* `prometheus.Summary`

You can extend this by adding more types, for instance, if you want to observe time and want
to avoid repetitive code you can create a `prometheusx.TimeHistogram`:
```go
package prometheusx

import (
	"reflect"
	"time"

	"github.com/cabify/gotoprom"
	"github.com/cabify/gotoprom/prometheusvanilla"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// TimeHistogramType is the reflect.Type of the TimeHistogram interface
	TimeHistogramType = reflect.TypeOf((*TimeHistogram)(nil)).Elem()
)

func init() {
	gotoprom.MustAddBuilder(TimeHistogramType, RegisterTimeHistogram)
}

// RegisterTimeHistogram registers a TimeHistogram after registering the underlying prometheus.Histogram in the prometheus.Registerer provided
// The function it returns returns a TimeHistogram type as an interface{}
func RegisterTimeHistogram(name, help, namespace string, labelNames []string, tag reflect.StructTag) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
	f, collector, err := prometheusvanilla.BuildHistogram(name, help, namespace, labelNames, tag)
	if err != nil {
		return nil, nil, err
	}

	return func(labels prometheus.Labels) interface{} {
		return timeHistogramAdapter{Histogram: f(labels).(prometheus.Histogram)}
	}, collector, nil
}

// TimeHistogram offers the basic prometheus.Histogram functionality
// with additional time-observing functions
type TimeHistogram interface {
	prometheus.Histogram
	// Duration observes the duration in seconds
	Duration(duration time.Duration)
	// Since observes the duration in seconds since the time point provided
	Since(time.Time)
}

type timeHistogramAdapter struct {
	prometheus.Histogram
}

// Duration observes the duration in seconds
func (to timeHistogramAdapter) Duration(duration time.Duration) {
	to.Observe(duration.Seconds())
}

// Since observes the duration in seconds since the time point provided
func (to timeHistogramAdapter) Since(duration time.Time) {
	to.Duration(time.Since(duration))
}
```

So you can later define it as:

```go
var metrics struct {
	DurationSeconds func() prometheusx.TimeHistogram `name:"duration_seconds" help:"Duration in seconds"`
}

func init() {
	gotoprom.MustInit(&metrics, "requests")
}
```

And use it as:

```go
// ...
defer metrics.DurationSeconds().Since(t0)
// ...
```


### Replacing metric builders
If you don't like the default metric builders, you can replace the `DefaultInitializer` with your own one.


## Performance

Obviously, there's a performance cost to perform the type-safety mapping magic to the original
Prometheus client's API.

In general terms, it takes 3x to increment a counter than with vanilla Prometheus, which is
around 600ns (we're talking about a portion of a microsecond, less than a thousandth of a millisecond)

```
$ go test -bench . -benchtime 3s
goos: darwin
goarch: amd64
pkg: github.com/cabify/gotoprom
BenchmarkVanilla-4    	10000000	       387 ns/op
BenchmarkGotoprom-4   	 5000000	      1049 ns/op
PASS
ok  	github.com/cabify/gotoprom	10.611s
```

In terms of memory, there's a also a 33% increase in terms of space, and 3x increase in allocations:

```
$ go test -bench . -benchmem
goos: darwin
goarch: amd64
pkg: github.com/cabify/gotoprom
BenchmarkVanilla-4    	 5000000	       381 ns/op	     336 B/op	       2 allocs/op
BenchmarkGotoprom-4   	 1000000	      1030 ns/op	     432 B/op	       6 allocs/op
PASS
ok  	github.com/cabify/gotoprom	3.369s
```

This costs are probably assumable in most of the applications, especially when measuring
network accesses, etc. which are magnitudes higher.
