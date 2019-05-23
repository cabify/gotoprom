package prometheusvanilla

import (
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
	labels                            = make(prometheus.Labels, 2)
	keys                              = make([]string, 0, len(labels))
	defaultTag      reflect.StructTag = `name:"some_name" help:"some help for the metric"`
	bucketsTag      reflect.StructTag = `name:"some_name" help:"some help for the metric" buckets:"0.001,0.005,0.01,0.05,0.1,0.5,1,5,10"`
	maxAgeTag       reflect.StructTag = `name:"some_name" help:"some help for the metric" max_age:"1h"`
	expectedBuckets                   = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}
	expectedMaxAge                    = time.Hour
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

	t.Run("Test building a observer", func(t *testing.T) {
		f, c, err := BuildObserver(name, help, nameSpace, keys, "")
		assert.NoError(t, err)
		assert.Implements(t, (*prometheus.Collector)(nil), c)
		assert.Implements(t, (*prometheus.Observer)(nil), f(labels))
	})

	t.Run("Test building a summary", func(t *testing.T) {
		f, c, err := BuildSummary(name, help, nameSpace, keys, "")
		assert.NoError(t, err)
		assert.Implements(t, (*prometheus.Collector)(nil), c)
		assert.Implements(t, (*prometheus.Summary)(nil), f(labels))
	})
}

func TestBuckets(t *testing.T) {
	t.Run("Test it retrieves custom buckets", func(t *testing.T) {
		buckets, err := bucketsFromTag(bucketsTag)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedBuckets, buckets)
	})

	t.Run("Test it returns default buckets when none are found", func(t *testing.T) {
		buckets, err := bucketsFromTag(defaultTag)
		assert.NoError(t, err)
		assert.ElementsMatch(t, DefaultBuckets(), buckets)
	})
}

func TestMaxAge(t *testing.T) {
	t.Run("Test it retrieves custom max_age", func(t *testing.T) {
		maxAge, err := maxAgeFromTag(maxAgeTag)
		assert.NoError(t, err)
		assert.Equal(t, maxAge, expectedMaxAge)
	})
	t.Run("Test it returns 0 when no max_age is found", func(t *testing.T) {
		maxAge, err := maxAgeFromTag(defaultTag)
		assert.NoError(t, err)
		assert.Equal(t, maxAge, time.Duration(0))
	})
}

func initLabels() {
	labels["labels1"] = "labels"
	labels["labels2"] = "labels"
	for k := range labels {
		keys = append(keys, k)
	}
}
