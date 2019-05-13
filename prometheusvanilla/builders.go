package prometheusvanilla

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ObserverType is the type of prometheus.Observer interface
	ObserverType = reflect.TypeOf((*prometheus.Observer)(nil)).Elem()
	// CounterType is the type of prometheus.Counter interface
	CounterType = reflect.TypeOf((*prometheus.Counter)(nil)).Elem()
	// GaugeType is the type of prometheus.Gauge interface
	GaugeType = reflect.TypeOf((*prometheus.Gauge)(nil)).Elem()
	// SummaryType is the type of prometheus.Summary interface
	SummaryType = reflect.TypeOf((*prometheus.Summary)(nil)).Elem()
)

// BuildCounter builds a prometheus.Counter in the given prometheus.Registerer
// The function it returns returns a prometheus.Counter type as an interface{}
func BuildCounter(name, help, namespace string, labelNames []string, tag reflect.StructTag) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      name,
			Help:      help,
			Namespace: namespace,
		},
		labelNames,
	)

	return func(labels prometheus.Labels) interface{} {
		return counter.With(labels)
	}, counter, nil
}

// BuildGauge builds a prometheus.Gauge in the given prometheus.Registerer
// The function it returns returns a prometheus.Gauge type as an interface{}
func BuildGauge(name, help, namespace string, labelNames []string, tag reflect.StructTag) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      name,
			Help:      help,
			Namespace: namespace,
		},
		labelNames,
	)

	return func(labels prometheus.Labels) interface{} {
		return gauge.With(labels)
	}, gauge, nil
}

// BuildObserver builds a prometheus.Observer
// The function it returns returns a prometheus.Observer type as an interface{}
func BuildObserver(name, help, namespace string, labelNames []string, tag reflect.StructTag) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
	buckets, err := bucketsFromTag(tag)
	if err != nil {
		return nil, nil, fmt.Errorf("build histogram %q: %s", name, err)
	}

	hist := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      name,
			Help:      help,
			Buckets:   buckets,
			Namespace: namespace,
		},
		labelNames,
	)

	return func(labels prometheus.Labels) interface{} {
		return hist.With(labels)
	}, hist, nil
}

// BuildSummary builds a prometheus.Summary
// The function it returns returns a prometheus.Summary type as an interface{}
func BuildSummary(name, help, namespace string, labelNames []string, tag reflect.StructTag) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
	maxAge, err := maxAgeFromTag(tag)
	if err != nil {
		return nil, nil, fmt.Errorf("build summary %q: %s", name, err)
	}

	sum := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:      name,
			Help:      help,
			Namespace: namespace,
			MaxAge:    maxAge,
		},
		labelNames,
	)

	return func(labels prometheus.Labels) interface{} {
		return sum.With(labels)
	}, sum, nil
}

func bucketsFromTag(tag reflect.StructTag) ([]float64, error) {
	bucketsString, ok := tag.Lookup("buckets")
	if !ok {
		return DefaultBuckets(), nil
	}
	bucketSlice := strings.Split(bucketsString, ",")
	buckets := make([]float64, len(bucketSlice))
	var err error
	for i := 0; i < len(bucketSlice); i++ {
		buckets[i], err = strconv.ParseFloat(bucketSlice[i], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bucket specified: %s", err)
		}
	}

	return buckets, nil
}

// DefaultBuckets provides a list of buckets you can use when you don't know what to use yet.
func DefaultBuckets() []float64 {
	return []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 25}
}

func maxAgeFromTag(tag reflect.StructTag) (time.Duration, error) {
	maxAgeString, ok := tag.Lookup("max_age")
	if !ok {
		return 0, nil
	}
	var maxAgeDuration time.Duration
	maxAgeDuration, err := time.ParseDuration(maxAgeString)
	if err != nil {
		return 0, fmt.Errorf("invalid time duration specified: %s", err)
	}
	return maxAgeDuration, nil
}
