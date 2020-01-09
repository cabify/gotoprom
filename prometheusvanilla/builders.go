package prometheusvanilla

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HistogramType is the type of prometheus.Histogram interface
	HistogramType = reflect.TypeOf((*prometheus.Histogram)(nil)).Elem()
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

// BuildHistogram builds a prometheus.Histogram
// The function it returns returns a prometheus.Histogram type as an interface{}
// It requires the buckets tag to be provided
// If the buckets tag is explicitly empty, then the Histogram will be built with default prometheus buckets
// which is prometheus.DefBuckets at the time this comment is written.
func BuildHistogram(name, help, namespace string, labelNames []string, tag reflect.StructTag) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
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
// It requires the objectives tag to be provided, and optionally the max_age tag
// If the objectives tag is explicitly empty, then the Summary will be built with default prometheus objectives
// which is no objectives at the time this comment is written.
func BuildSummary(name, help, namespace string, labelNames []string, tag reflect.StructTag) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
	maxAge, err := maxAgeFromTag(tag)
	if err != nil {
		return nil, nil, fmt.Errorf("build summary %q: %s", name, err)
	}
	objectives, err := objectivesFromTag(tag)
	if err != nil {
		return nil, nil, fmt.Errorf("build summary %q: %s", name, err)
	}

	sum := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Namespace:  namespace,
			MaxAge:     maxAge,
			Objectives: objectives,
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
		return nil, fmt.Errorf("buckets not specified")
	}

	if bucketsString == "" {
		return nil, nil
	}

	bucketSlice := strings.Split(bucketsString, ",")
	buckets := make([]float64, len(bucketSlice))

	var err error
	for i := range bucketSlice {
		buckets[i], err = strconv.ParseFloat(bucketSlice[i], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bucket specified: %s", err)
		}
	}

	return buckets, nil
}

func maxAgeFromTag(tag reflect.StructTag) (time.Duration, error) {
	maxAgeString, ok := tag.Lookup("max_age")
	if !ok {
		return 0, nil
	}
	maxAgeDuration, err := time.ParseDuration(maxAgeString)
	if err != nil {
		return 0, fmt.Errorf("invalid max_age tag specified: %s", err)
	}
	return maxAgeDuration, nil
}

// objectivesFromTag will return the objectives from the tag provided
// if there's no objectives tag, it will return an error
// if objectives is an empty string, it will return a nil value instead of an initialized empty map
// this is intended to initialize prometheus metric with default values, as prometheus will
// check for the value to be nil instead of checking for its len to be 0 (like it does for buckets)
func objectivesFromTag(tag reflect.StructTag) (map[float64]float64, error) {
	quantileString, ok := tag.Lookup("objectives")
	if !ok {
		return nil, fmt.Errorf("objectives not specified")
	}

	if quantileString == "" {
		return nil, nil
	}

	quantileSlice := strings.Split(quantileString, ",")
	objectives := make(map[float64]float64, len(quantileSlice))
	for i := range quantileSlice {
		q, err := strconv.ParseFloat(quantileSlice[i], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid objective specified: %s", err)
		}
		objectives[q] = absError(q)
	}
	return objectives, nil
}

// absError will calculate the absolute error for a given objective up to 3 decimal places.
// The variance is calculated based on the given quantile.
// Values in lower quantiles have a higher probability of being similar, so we can apply greater variances.
func absError(obj float64) float64 {
	return math.Round((0.1*(1-obj))*1000) / 1000
}
