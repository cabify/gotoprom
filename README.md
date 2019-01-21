# gotoprom
## A Prometheus metrics builder

`gotoprom` offers an easy to use declarative API for building and using Prometheus metrics. 
It doesn't replace the [official Prometheus client](https://github.com/prometheus/client_golang)
but adds a wrapper on top of it.

It tries to solve the ugly initialization code and remove the usage of error-prone `map[string]string` 
as well as groups all the metrics of a given package together.

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