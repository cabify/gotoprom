package gotoprom_test

import (
	"testing"
	"time"

	"github.com/cabify/gotoprom"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func Test_InitHappyCase(t *testing.T) {
	type labels struct {
		Region   string `label:"region"`
		Status   string `label:"status"`
		DuckType string `label:"duck_type"`
	}

	var metrics struct {
		HTTPRequestTime     func(labels) prometheus.Observer `name:"http_request_count" help:"Time taken to serve a HTTP request" metricsbuckets:"0.001,0.005,0.01,0.05,0.1,0.5,1,5,10"`
		DuvelsEmptied       func(labels) prometheus.Counter  `name:"duvels_emptied" help:"Delirium floor sweep count"`
		RubberDuckInTherapy func(labels) prometheus.Gauge    `name:"rubber_ducks_in_therapy" help:"Number of rubber ducks who need help after some intense coding"`
		NoLabels            func() prometheus.Counter        `name:"no_labels" help:"Metric without labels"`
	}

	gotoprom.MustInit(&metrics, "delirium")

	// Elsewhere in the code
	theseLabels := labels{
		Region:   "bruxelles",
		Status:   "recovering",
		DuckType: "definitely_rubber",
	}

	metrics.HTTPRequestTime(theseLabels).Observe(time.Second.Seconds())
	metrics.DuvelsEmptied(theseLabels).Add(4)
	metrics.RubberDuckInTherapy(theseLabels).Set(12)
	metrics.NoLabels().Add(288.88)
}

func Test_NestedMetrics(t *testing.T) {
	type testLabels struct {
		TestLabel string `label:"test_label"`
	}

	var metrics struct {
		Server struct {
			Hits func(testLabels) prometheus.Counter `name:"hits_total" help:"bla bla"`
		} `namespace:"server"`

		Client struct {
			Requests func(testLabels) prometheus.Counter `name:"requests_total" help:"cuac cuac"`
		} `namespace:"client"`

		MemoryConsumption func(testLabels) prometheus.Gauge `name:"memory_consumption_bytes" help:"wololo"`
	}

	gotoprom.MustInit(&metrics, "testservice")

	metrics.Server.Hits(testLabels{}).Add(1.0)
	metrics.Client.Requests(testLabels{}).Add(2.0)
	metrics.MemoryConsumption(testLabels{}).Set(3.0)

	mfs, err := prometheus.DefaultGatherer.Gather()
	assert.Nil(t, err)

	reportedMetrics := map[string]struct{}{}
	for _, m := range mfs {
		for _, mm := range m.Metric {
			for _, l := range mm.Label {
				if *l.Name == "test_label" {
					reportedMetrics[*m.Name] = struct{}{}
				}
			}
		}
	}

	assert.Equal(t, map[string]struct{}{
		"testservice_server_hits_total":        struct{}{},
		"testservice_client_requests_total":    struct{}{},
		"testservice_memory_consumption_bytes": struct{}{},
	}, reportedMetrics)
}

func Test_EmbeddedLabels(t *testing.T) {
	type commonLabels struct {
		CommonValue string `label:"common_value"`
	}

	type specificLabels struct {
		commonLabels
		SpecificValue string `label:"specific_value"`
	}

	var metrics struct {
		WithLabels func(specificLabels) prometheus.Counter `name:"with_labels" help:"Some metric with labels"`
	}

	gotoprom.MustInit(&metrics, "namespace")

	metrics.WithLabels(specificLabels{
		commonLabels:  commonLabels{CommonValue: "common"},
		SpecificValue: "specific",
	}).Add(42.0)

	reportedLabels := retrieveReportedLabels(t, "namespace_with_labels")

	assert.Equal(t, map[string]string{
		"common_value":   "common",
		"specific_value": "specific",
	}, reportedLabels)
}

func Test_LabelsWithBooleans(t *testing.T) {
	type labelsWithBools struct {
		StringValue  string `label:"string_value"`
		BooleanValue bool   `label:"bool_value"`
	}

	var metrics struct {
		WithLabels func(labelsWithBools) prometheus.Observer `name:"with_booleans" help:"Parse booleans as strings"`
	}

	gotoprom.MustInit(&metrics, "testbooleans")

	metrics.WithLabels(labelsWithBools{
		StringValue:  "string",
		BooleanValue: false,
	}).Observe(288.0)

	reportedLabels := retrieveReportedLabels(t, "testbooleans_with_booleans")

	assert.Equal(t, map[string]string{
		"string_value": "string",
		"bool_value":   "false",
	}, reportedLabels)
}

func Test_DefaultLabelValues(t *testing.T) {
	type labelsWithEmptyValues struct {
		StringWithEmpty    string `label:"string_with_default" default:"none"`
		StringWithoutEmpty string `label:"string_without_default"`
	}

	var metrics struct {
		WithLabels func(labelsWithEmptyValues) prometheus.Observer `name:"with_labels" help:"Assign default values"`
	}
	gotoprom.MustInit(&metrics, "testdefault")

	metrics.WithLabels(labelsWithEmptyValues{}).Observe(288.0)

	reportedLabels := retrieveReportedLabels(t, "testdefault_with_labels")

	assert.Equal(t, map[string]string{
		"string_with_default":    "none",
		"string_without_default": "",
	}, reportedLabels)
}

func Test_LabelsWithUnsupportedFields(t *testing.T) {
	type labelsWithUnsupportedFields struct {
		StringValue string `label:"string_value"`
		IntValue    int    `label:"int_value"`
	}

	var metrics struct {
		WithLabels func(labelsWithUnsupportedFields) prometheus.Counter `name:"with_unsupported_fields" help:"Can't parse integer labels"`
	}

	err := gotoprom.Init(&metrics, "test")
	assert.NotNil(t, err)
}

func Test_HistogramWithUnsupportedBuckets(t *testing.T) {
	var metrics struct {
		Histogram func() prometheus.Observer `name:"with_broken_buckets" help:"Wrong buckets" buckets:"0.005, +inf"`
	}
	err := gotoprom.Init(&metrics, "test")
	assert.NotNil(t, err)
}

func retrieveReportedLabels(t *testing.T, metric string) map[string]string {
	mfs, err := prometheus.DefaultGatherer.Gather()
	assert.Nil(t, err)

	reportedLabels := map[string]string{}
	for _, m := range mfs {
		if *m.Name == metric {
			for _, mm := range m.Metric {
				for _, l := range mm.Label {
					reportedLabels[*l.Name] = *l.Value
				}
			}
		}
	}

	return reportedLabels
}
