# gotoprom
## A Prometheus metrics builder

`gotoprom` offers an easy to use declarative API for building and using Prometheus metrics.
It doesn't replace the [official Prometheus client](https://github.com/prometheus/client_golang)
but adds a wrapper on top of it.

It tries to solve the ugly initialization code and remove the usage of error-prone `map[string]string`
by replacing them with functions with type-safe label structs as well as groups all the metrics of a
given package together.

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
	Code int `label:"code"`
	Method string `label:"method"`
}

gotoprom.MustInit(&metrics, "http")

// ...

metrics.Reqs(labels{Code: 404, Method: "POST"}).Inc()

```

`gotoprom` is built for developers who like to navigate the code using their IDEs and like to use
the "find usages" functionality, making refactoring and debugging easier at the code of writing
slightly more verbose code.

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

This costs are probably assumable in most of the applications, specially when measuring
network accesses, etc. which are magnitudes higher.
