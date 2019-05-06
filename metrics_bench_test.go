package gotoprom_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cabify/gotoprom"

	"github.com/cabify/gotoprom/prometheusvanilla"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

type DummyRegistry struct{}

func (DummyRegistry) Gather() ([]*io_prometheus_client.MetricFamily, error) { return nil, nil }
func (DummyRegistry) Register(prometheus.Collector) error                   { return nil }
func (DummyRegistry) MustRegister(...prometheus.Collector)                  {}
func (DummyRegistry) Unregister(prometheus.Collector) bool                  { return true }

func BenchmarkDefaultLib(b *testing.B) {
	prometheus.DefaultRegisterer = DummyRegistry{}

	cvec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "benchmarks",
		Name:      "do_add1",
		Help:      "does an add",
	}, []string{"region"})

	prometheus.MustRegister(cvec)

	for n := 0; n < b.N; n++ {
		cvec.With(map[string]string{
			"region": "madrid",
		}).Add(1)
	}
}

func BenchmarkMagicLib(b *testing.B) {
	prometheus.DefaultRegisterer = DummyRegistry{}

	type labels struct {
		Region string `label:"region"`
	}
	var metrics struct {
		DoAdd func(labels) prometheus.Counter `name:"do_add2" help:"does an add"`
	}
	initializer := gotoprom.NewInitializer(DummyRegistry{})
	initializer.MustAddBuilder(prometheusvanilla.CounterType, prometheusvanilla.BuildCounter)
	initializer.MustInit(&metrics, "benchmarks")

	for n := 0; n < b.N; n++ {
		metrics.DoAdd(labels{Region: "madrid"}).Add(1)
	}
}

func BenchmarkSprintf(b *testing.B) {
	err := errors.New("everything is broken")
	for n := 0; n < b.N; n++ {
		_ = fmt.Errorf("something failed: %s", err)
	}
}
