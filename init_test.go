package gotoprom

import (
	"testing"

	"github.com/cabify/gotoprom/prometheusvanilla"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestInitializer_MustAddBuilder(t *testing.T) {
	t.Run("adds builder", func(t *testing.T) {
		initializer := NewInitializer(prometheus.NewRegistry())
		initializer.MustAddBuilder(prometheusvanilla.GaugeType, prometheusvanilla.BuildGauge)

		err := initializer.Init(&struct {
			Metric func() prometheus.Gauge `name:"gauge" help:"help"`
		}{}, "namespace")
		assert.NoError(t, err)
	})
	t.Run("same builder twice panics", func(t *testing.T) {
		initializer := NewInitializer(prometheus.NewRegistry())
		initializer.MustAddBuilder(prometheusvanilla.GaugeType, prometheusvanilla.BuildGauge)

		assert.Panics(t, func() {
			initializer.MustAddBuilder(prometheusvanilla.GaugeType, prometheusvanilla.BuildGauge)
		})
	})
}

func TestInitializer_AddBuilder(t *testing.T) {
	t.Run("same builder twice fails", func(t *testing.T) {
		initializer := NewInitializer(prometheus.NewRegistry())
		err := initializer.AddBuilder(prometheusvanilla.GaugeType, prometheusvanilla.BuildGauge)
		assert.NoError(t, err)

		err = initializer.AddBuilder(prometheusvanilla.GaugeType, prometheusvanilla.BuildGauge)
		assert.Error(t, err)
	})
}

func TestInitializer_MustInit(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		initializer := NewInitializer(prometheus.NewRegistry())

		structPtr := &struct{}{}
		assert.NotPanics(t, func() {
			initializer.MustInit(structPtr, "namespace")
		})
	})

	t.Run("panics when can't init", func(t *testing.T) {
		initializer := NewInitializer(prometheus.NewRegistry())

		notAPointer := struct{}{}
		assert.Panics(t, func() {
			initializer.MustInit(notAPointer, "namespace")
		})
	})
}

func TestInitializer_Init(t *testing.T) {
	someInt := 3
	type someLabels struct {
		Label string `label:"label"`
	}
	type someLabelsWithoutLabelTag struct {
		LabelWithoutLabelTag string
	}
	t.Run("fails", func(t *testing.T) {
		for _, tc := range []struct {
			desc    string
			metrics interface{}
		}{
			{
				desc:    "not a pointer to struct",
				metrics: struct{}{},
			},
			{
				desc:    "pointer to a non struct",
				metrics: &someInt,
			},
			{
				desc: "contains a non-supported field, like a string",
				metrics: &struct {
					Something string
				}{},
			},
			{
				desc: "non-exported field",
				metrics: &struct {
					foo func() prometheus.Gauge `name:"unexported" help:"this can't be set"`
				}{},
			},
			{
				desc: "missing name",
				metrics: &struct {
					Foo func() prometheus.Gauge `help:"missing name tag"`
				}{},
			},
			{
				desc: "missing help",
				metrics: &struct {
					Foo func() prometheus.Gauge `name:"nohelp"`
				}{},
			},
			{
				desc: "more than one return argument",
				metrics: &struct {
					Foo func() (prometheus.Gauge, error) `name:"toomanyreturnargs" help:"what is this?"`
				}{},
			},
			{
				desc: "no builders defined",
				metrics: &struct {
					Foo func() prometheus.Counter `name:"unknowntype" help:"Counter is not registered"`
				}{},
			},
			{
				desc: "builder fails",
				metrics: &struct {
					Foo func() prometheus.Summary `name:"unknowntype" help:"max_age is malformed" max_age:"bar"`
				}{},
			},
			{
				desc: "builder fails",
				metrics: &struct {
					Foo func() prometheus.Summary `name:"unknowntype" help:"max_age is malformed" max_age:"bar"`
				}{},
			},
			{
				desc: "too many params",
				metrics: &struct {
					Foo func(someLabels, string) prometheus.Gauge `name:"too many input params" help:"only labels are accepted"`
				}{},
			},
			{
				desc: "labels without label tag",
				metrics: &struct {
					Foo func(someLabelsWithoutLabelTag) prometheus.Gauge `name:"nolabeltag" help:"This labels are missing the label tag"`
				}{},
			},
			{
				desc: "metric registration fails",
				metrics: &struct {
					Foo func() prometheus.Gauge `name:"metric" help:"first gauge"`
					Bar func() prometheus.Gauge `name:"metric" help:"name duplicates the previous one"`
				}{},
			},
		} {
			t.Run(tc.desc, func(t *testing.T) {
				initializer := NewInitializer(prometheus.NewRegistry())
				initializer.MustAddBuilder(prometheusvanilla.GaugeType, prometheusvanilla.BuildGauge)
				initializer.MustAddBuilder(prometheusvanilla.SummaryType, prometheusvanilla.BuildSummary)

				err := initializer.Init(tc.metrics, "namespace")
				assert.Error(t, err)
			})
		}
	})
}
