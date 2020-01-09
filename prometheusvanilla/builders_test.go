package prometheusvanilla

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

const (
	name      = "some_name"
	help      = "Some help"
	nameSpace = "some_namespace"
)

var (
	labels                       = make(prometheus.Labels, 2)
	keys                         = make([]string, 0, len(labels))
	defaultTag reflect.StructTag = `name:"some_name" help:"some help for the metric"`

	bucketsTag          reflect.StructTag = `name:"some_name" help:"some help for the metric" buckets:"0.001,0.005,0.01,0.05,0.1,0.5,1,5,10"`
	emptyBucketsTag     reflect.StructTag = `name:"some_name" help:"some help for the metric" buckets:""`
	malformedBucketsTag reflect.StructTag = `name:"some_name" help:"some help for the metric" buckets:"fourtytwo"`

	maxAgeTag              reflect.StructTag = `name:"some_name" help:"some help for the metric" max_age:"1h"`
	objectivesTag          reflect.StructTag = `name:"some_name" help:"some help for the metric" max_age:"1h" objectives:"0.55,0.95,0.98"`
	objectivesMalformedTag reflect.StructTag = `name:"some_name" help:"some help for the metric" max_age:"1h" objectives:"notFloat"`

	expectedBuckets    = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}
	expectedMaxAge     = time.Hour
	expectedObjectives = map[float64]float64{0.55: 0.045, 0.95: 0.005, 0.98: 0.002}
)

func TestBuilders(t *testing.T) {
	initLabels()

	t.Run("Test building a counter", func(t *testing.T) {
		f, c, err := BuildCounter(name, help, nameSpace, keys, "")
		assert.NoError(t, err)
		assert.Implements(t, (*prometheus.Collector)(nil), c)
		assert.Implements(t, (*prometheus.Counter)(nil), f(labels))
	})

	t.Run("Test building a gauge", func(t *testing.T) {
		f, c, err := BuildGauge(name, help, nameSpace, keys, "")
		assert.NoError(t, err)
		assert.Implements(t, (*prometheus.Collector)(nil), c)
		assert.Implements(t, (*prometheus.Counter)(nil), f(labels))
	})

	t.Run("Test building a histogram", func(t *testing.T) {
		f, c, err := BuildHistogram(name, help, nameSpace, keys, `buckets:""`)
		assert.NoError(t, err)
		assert.Implements(t, (*prometheus.Collector)(nil), c)
		assert.Implements(t, (*prometheus.Histogram)(nil), f(labels))
	})

	t.Run("Test building a histogram with malformed buckets", func(t *testing.T) {
		_, _, err := BuildHistogram(name, help, nameSpace, keys, `buckets:"foo"`)
		assert.Error(t, err)
	})

	t.Run("Test building a summary", func(t *testing.T) {
		f, c, err := BuildSummary(name, help, nameSpace, keys, "")
		assert.NoError(t, err)
		assert.Implements(t, (*prometheus.Collector)(nil), c)
		assert.Implements(t, (*prometheus.Summary)(nil), f(labels))
	})

	t.Run("Test building a summary with malformed max_age", func(t *testing.T) {
		_, _, err := BuildSummary(name, help, nameSpace, keys, `max_age:"one year" objectives:"0.1,0.25"`)
		assert.Error(t, err)
	})
}

func TestBuckets(t *testing.T) {
	t.Run("Test it retrieves custom buckets", func(t *testing.T) {
		buckets, err := bucketsFromTag(bucketsTag)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedBuckets, buckets)
	})

	t.Run("Test empty string generates empty buckets slice", func(t *testing.T) {
		buckets, err := bucketsFromTag(emptyBucketsTag)
		assert.NoError(t, err)
		assert.Len(t, buckets, 0)
	})

	t.Run("Test it returns error when buckets are malformed", func(t *testing.T) {
		_, err := bucketsFromTag(malformedBucketsTag)
		assert.Error(t, err)
	})

	t.Run("Test it returns error when none are found", func(t *testing.T) {
		_, err := bucketsFromTag(defaultTag)
		assert.Error(t, err)
	})
}

func TestMaxAge(t *testing.T) {
	t.Run("Test it retrieves custom max_age", func(t *testing.T) {
		maxAge, err := maxAgeFromTag(maxAgeTag)
		assert.NoError(t, err)
		assert.Equal(t, expectedMaxAge, maxAge)
	})
	t.Run("Test it returns 0 when no max_age is found", func(t *testing.T) {
		maxAge, err := maxAgeFromTag(defaultTag)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), maxAge)
	})
}

func TestObjectives(t *testing.T) {
	t.Run("Test parsing objectives from tag", func(t *testing.T) {
		obj, err := objectivesFromTag(objectivesTag)
		assert.NoError(t, err)
		assert.Equal(t, expectedObjectives, obj)
	})
	t.Run("Test returning default objective values when none are specified", func(t *testing.T) {
		obj, err := objectivesFromTag(defaultTag)
		assert.NoError(t, err)
		assert.Equal(t, DefaultObjectives(), obj)
	})
	t.Run("Test returning default objective values when none are specified", func(t *testing.T) {
		obj, err := objectivesFromTag(objectivesMalformedTag)
		assert.Error(t, err)
		assert.Nil(t, obj)
	})
}

func TestAbsError(t *testing.T) {
	tests := []struct {
		objective float64
		absErr    float64
	}{
		{0.5, 0.05},
		{0.99, 0.001},
		{0.999, 0},
		{0, 0.1},
		{-10, 1.1},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test calculate absolute error for obejctive %2f", test.objective), func(t *testing.T) {
			e := absError(test.objective)
			assert.Equal(t, test.absErr, e)
		})
	}
}

func initLabels() {
	labels["labels1"] = "labels"
	labels["labels2"] = "labels"
	for k := range labels {
		keys = append(keys, k)
	}
}
