package gotoprom

import (
	"reflect"

	"github.com/cabify/gotoprom/prometheusvanilla"
	"github.com/prometheus/client_golang/prometheus"
)

// DefaultInitializer is the instance instance of the Initializer used by default
var DefaultInitializer = NewInitializer(prometheus.DefaultRegisterer)

func init() {
	DefaultInitializer.MustAddBuilder(prometheusvanilla.HistogramType, prometheusvanilla.BuildHistogram)
	DefaultInitializer.MustAddBuilder(prometheusvanilla.CounterType, prometheusvanilla.BuildCounter)
	DefaultInitializer.MustAddBuilder(prometheusvanilla.GaugeType, prometheusvanilla.BuildGauge)
	DefaultInitializer.MustAddBuilder(prometheusvanilla.SummaryType, prometheusvanilla.BuildSummary)
}

// MustAddBuilder will AddBuilder and panic if an error occurs
func MustAddBuilder(typ reflect.Type, registerer Builder) {
	DefaultInitializer.MustAddBuilder(typ, registerer)
}

// AddBuilder adds a new registerer for type typ.
func AddBuilder(typ reflect.Type, registerer Builder) error {
	return DefaultInitializer.AddBuilder(typ, registerer)
}

// MustInit initializes the metrics or panics.
func MustInit(metrics interface{}, namespace string) {
	DefaultInitializer.MustInit(metrics, namespace)
}

// Init initializes the metrics in the given namespace.
func Init(metrics interface{}, namespace string) error {
	return DefaultInitializer.Init(metrics, namespace)
}
