package gotoprom

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Builder is a function that registers a metric and provides a function that
// creates the metric reporter for given values
// Note that the type of the first return value of a Builder should be (in Java words):
// func() interface{} implements <typ>
type Builder func(
	name, help, namespace string,
	labelNames []string,
	tag reflect.StructTag,
) (func(prometheus.Labels) interface{}, prometheus.Collector, error)

// Initializer represents an instance of the initializing functionality
type Initializer interface {
	// MustAddBuilder will AddBuilder and panic if an error occurs
	MustAddBuilder(typ reflect.Type, registerer Builder)
	// AddBuilder adds a new registerer for type typ.
	// Note that the type of the first return value of Builder should be (in Java words):
	// func() interface{} implements <typ>
	AddBuilder(typ reflect.Type, registerer Builder) error

	// MustInit initialises the metrics or panics.
	MustInit(metrics interface{}, namespace string)

	// Init initialises the metrics in the given namespace.
	Init(metrics interface{}, namespace string) error
}

// NewInitializer creates a new Initializer for the prometheus.Registerer provided
func NewInitializer(registerer prometheus.Registerer) Initializer {
	return initializer{
		registerer: registerer,
		builders:   make(map[reflect.Type]Builder),
	}
}

type initializer struct {
	registerer prometheus.Registerer
	builders   map[reflect.Type]Builder
}

// MustAddBuilder will AddBuilder and panic if an error occurs
func (in initializer) MustAddBuilder(typ reflect.Type, builder Builder) {
	if err := in.AddBuilder(typ, builder); err != nil {
		panic(err)
	}
}

// AddBuilder adds a new registerer for type typ.
// Note that the type of the first return value of Builder should be (in Java words):
// func() interface{} implements <typ>
func (in initializer) AddBuilder(typ reflect.Type, builder Builder) error {
	if _, ok := in.builders[typ]; ok {
		return fmt.Errorf("type %q already has a builder", typ.Name())
	}
	in.builders[typ] = builder
	return nil
}

// MustInit initialises the metrics or panics.
func (in initializer) MustInit(metrics interface{}, namespace string) {
	if err := in.Init(metrics, namespace); err != nil {
		panic(err)
	}
}

// Init initialises the metrics in the given namespace.
func (in initializer) Init(metrics interface{}, namespace string) error {
	metricsPtr := reflect.ValueOf(metrics)
	if metricsPtr.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer to metrics struct, got %q", metricsPtr.Kind())
	}

	return in.initMetrics(metricsPtr.Elem(), namespace)
}

func (in initializer) initMetrics(group reflect.Value, namespaces ...string) error {
	if group.Kind() != reflect.Struct {
		return fmt.Errorf("expected group %s to be a struct, got %q", group.Type().Name(), group.Kind())
	}

	for i := 0; i < group.Type().NumField(); i++ {

		field := group.Field(i)
		fieldType := group.Type().Field(i)

		if fieldType.Type.Kind() == reflect.Func {
			if err := in.initMetricFunc(field, fieldType, namespaces...); err != nil {
				return err
			}
		} else if fieldType.Type.Kind() == reflect.Struct {
			namespace, ok := fieldType.Tag.Lookup("namespace")
			if !ok {
				return fmt.Errorf("field %s does not have the namespace tag defined", fieldType.Name)
			}
			if err := in.initMetrics(field, append(namespaces, namespace)...); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("metrics are expected to contain only funcs or nested metric structs, but %s is %s", fieldType.Name, fieldType.Type.Kind())
		}
	}
	return nil
}

func (in initializer) initMetricFunc(field reflect.Value, structField reflect.StructField, namespaces ...string) (err error) {
	namespace := strings.Join(namespaces, "_")
	fieldType := field.Type()

	if !field.CanSet() {
		return fmt.Errorf("field %q needs be exported", structField.Name)
	}

	tag := structField.Tag
	name, ok := tag.Lookup("name")
	if !ok {
		return fmt.Errorf("name tag for %s missing", structField.Name)
	}
	help, ok := tag.Lookup("help")
	if !ok {
		return fmt.Errorf("help tag for %s missing", structField.Name)
	}

	// Validate the input of the metric function, it should have zero or one arguments
	// If it has one argument, it should be a struct correctly tagged with label names
	// If there are no input arguments, this metric will not have labels registered
	var labelIndexes = make(map[label][]int)
	if fieldType.NumIn() > 1 {
		return fmt.Errorf("field %s: expected 1 in arg, got %d", structField.Name, fieldType.NumIn())
	} else if fieldType.NumIn() == 1 {
		inArg := fieldType.In(0)
		err := findLabelIndexes(inArg, labelIndexes)
		if err != nil {
			return fmt.Errorf("build labels for field %q: %s", structField.Name, err)
		}
	}
	labelNames := make([]string, 0, len(labelIndexes))
	for label := range labelIndexes {
		labelNames = append(labelNames, label.name)
	}

	// Validate the output and register the correct metric type based on the output type
	if fieldType.NumOut() != 1 {
		return fmt.Errorf("field %s: expected 1 return arg, got %d", structField.Name, fieldType.NumOut())
	}
	returnArg := fieldType.Out(0)

	builder, ok := in.builders[returnArg]
	if !ok {
		return fmt.Errorf("field %s: no builder found for type %q", structField.Name, returnArg.Name())
	}

	// metric's type is:
	//   func(map[string]string) interface{} implements <returnArg>
	// but there's no use case for generics in Go
	metric, collector, err := builder(name, help, namespace, labelNames, tag)
	if err != nil {
		return fmt.Errorf("build metric %q: %s", name, err)
	}

	err = in.registerer.Register(collector)
	if err != nil {
		return fmt.Errorf("register metric %q: %s", name, err)
	}

	metricFunc := func(args []reflect.Value) []reflect.Value {
		labels := make(prometheus.Labels, len(labelIndexes))
		for label, index := range labelIndexes {

			value := args[0].FieldByIndex(index)

			if label.hasDefaultValue && value.Interface() == label.zeroTypeValueInterface {
				value = label.defaultValue
			}

			switch k := label.kind; k {
			case reflect.Bool:
				labels[label.name] = strconv.FormatBool(value.Bool())
			case reflect.String:
				labels[label.name] = value.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				labels[label.name] = strconv.FormatInt(value.Int(), 10)
			default:
				// Should not happen
				panic(fmt.Errorf("field %s has unsupported kind %v", label.name, label.kind))
			}
		}
		return []reflect.Value{reflect.ValueOf(metric(labels)).Convert(returnArg)}
	}

	field.Set(reflect.MakeFunc(fieldType, metricFunc))
	return nil
}

type label struct {
	kind reflect.Kind
	name string

	// hasDefaultValue indicates that zero values should be replaced by default values
	hasDefaultValue bool
	// zeroTypeValueInterface is the interface value of the zero-value for this field's type
	zeroTypeValueInterface interface{}
	// defaultValue is the value to be assigned if hasDefaultValue is true and provided value is the zeroTypeValue
	defaultValue reflect.Value
}

func findLabelIndexes(typ reflect.Type, indexes map[label][]int, current ...int) error {
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("expected to get a Struct for %s, got %s", typ.Name(), typ.Kind())
	}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type.Kind() == reflect.Struct {
			if err := findLabelIndexes(f.Type, indexes, append(current, i)...); err != nil {
				return err
			}
		} else {

			labelTag, ok := f.Tag.Lookup("label")
			if !ok {
				return fmt.Errorf("field %s does not have the label tag", f.Name)
			}

			switch k := f.Type.Kind(); k {
			case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			default:
				return fmt.Errorf("field %s has unsupported type %v", labelTag, k)
			}

			label := label{
				kind: f.Type.Kind(),
				name: labelTag,
			}

			if emptyTag, ok := f.Tag.Lookup("default"); ok {
				label.hasDefaultValue = true
				label.defaultValue = reflect.ValueOf(emptyTag)
				label.zeroTypeValueInterface = reflect.Zero(f.Type).Interface()
			}

			if _, ok := indexes[label]; ok {
				return fmt.Errorf("label %+v can't be registered twice", label)
			}
			indexes[label] = append(current, i)
		}
	}
	return nil
}
