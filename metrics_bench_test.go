package gotoprom_test

import (
	"testing"

	"github.com/cabify/gotoprom"

	"github.com/cabify/gotoprom/prometheusvanilla"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkVanilla(b *testing.B) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	cvec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "benchmarks",
		Name:      "do_add1",
		Help:      "does an add",
	}, []string{"region"})

	prometheus.MustRegister(cvec)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		cvec.With(map[string]string{
			"region": "madrid",
		}).Add(1)
	}
}

func BenchmarkGotoprom(b *testing.B) {
	type labels struct {
		Region string `label:"region"`
	}
	var metrics struct {
		DoAdd func(labels) prometheus.Counter `name:"do_add2" help:"does an add"`
	}
	initializer := gotoprom.NewInitializer(prometheus.NewRegistry())
	initializer.MustAddBuilder(prometheusvanilla.CounterType, prometheusvanilla.BuildCounter)
	initializer.MustInit(&metrics, "benchmarks")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		metrics.DoAdd(labels{Region: "madrid"}).Add(1)
	}
}
