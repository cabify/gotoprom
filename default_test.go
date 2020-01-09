package gotoprom

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cabify/gotoprom/prometheusvanilla"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMustAddBuilder(t *testing.T) {
	initializerMock, tearDown := mockDefaultInitializer()
	defer tearDown()
	defer initializerMock.AssertExpectations(t)

	expectedErr := errors.New("my err")

	typ := prometheusvanilla.HistogramType
	builder := func(
		name, help, namespace string,
		labelNames []string,
		tag reflect.StructTag,
	) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
		return nil, nil, expectedErr
	}

	initializerMock.On("MustAddBuilder", typ, mock.Anything).Run(func(args mock.Arguments) {
		// we can't assert that two functions are the same, so we invoke it and see if it's ours
		_, _, err := args[1].(Builder)("", "", "", nil, reflect.StructTag(""))
		assert.Equal(t, expectedErr, err)
	}).Once()

	MustAddBuilder(typ, builder)
}

func TestAddBuilder(t *testing.T) {
	initializerMock, tearDown := mockDefaultInitializer()
	defer tearDown()
	defer initializerMock.AssertExpectations(t)

	expectedErr := errors.New("my err")

	typ := prometheusvanilla.HistogramType
	builder := func(
		name, help, namespace string,
		labelNames []string,
		tag reflect.StructTag,
	) (func(prometheus.Labels) interface{}, prometheus.Collector, error) {
		return nil, nil, expectedErr
	}

	initializerMock.On("AddBuilder", typ, mock.Anything).Run(func(args mock.Arguments) {
		// we can't assert that two functions are the same, so we invoke it and see if it's ours
		_, _, err := args[1].(Builder)("", "", "", nil, reflect.StructTag(""))
		assert.Equal(t, expectedErr, err)
	}).Return(expectedErr).Once()

	err := AddBuilder(typ, builder)
	assert.Equal(t, expectedErr, err)
}

func TestInit(t *testing.T) {
	initializerMock, tearDown := mockDefaultInitializer()
	defer tearDown()
	defer initializerMock.AssertExpectations(t)

	expectedErr := errors.New("my err")

	metrics := struct{ whatever int }{}
	namespace := "some namespace"

	initializerMock.On("Init", metrics, namespace).Return(expectedErr).Once()

	err := Init(metrics, namespace)
	assert.Equal(t, expectedErr, err)
}

func TestMustInit(t *testing.T) {
	initializerMock, tearDown := mockDefaultInitializer()
	defer tearDown()
	defer initializerMock.AssertExpectations(t)

	metrics := struct{ whatever int }{}
	namespace := "some namespace"

	initializerMock.On("MustInit", metrics, namespace).Once()

	MustInit(metrics, namespace)
}

func mockDefaultInitializer() (mock *InitializerMock, tearDown func()) {
	original := DefaultInitializer
	mock = &InitializerMock{}
	DefaultInitializer = mock
	return mock, func() { DefaultInitializer = original }
}

type InitializerMock struct {
	mock.Mock
}

func (m *InitializerMock) MustAddBuilder(typ reflect.Type, registerer Builder) {
	m.Called(typ, registerer)
}

func (m *InitializerMock) AddBuilder(typ reflect.Type, registerer Builder) error {
	ret := m.Called(typ, registerer)
	return ret[0].(error)
}

func (m *InitializerMock) MustInit(metrics interface{}, namespace string) {
	m.Called(metrics, namespace)
}

func (m *InitializerMock) Init(metrics interface{}, namespace string) error {
	ret := m.Called(metrics, namespace)
	return ret[0].(error)
}
