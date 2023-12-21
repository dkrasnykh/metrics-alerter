package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
)

func TestValidate(t *testing.T) {
	r := storage.New()
	s := New(r)

	err := s.Validate(models.CounterType, "101")
	require.NoError(t, err)

	err = s.Validate(models.GaugeType, "101.0")
	require.NoError(t, err)

	err = s.Validate("unknown", "101")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownMetricType))
}

func TestSave(t *testing.T) {
	r := storage.New()
	s := New(r)

	err := s.Save(models.CounterType, `name1`, `101`)
	require.NoError(t, err)
}

func TestCalculateCounterValue(t *testing.T) {
	r := storage.New()
	s := New(r)

	value := s.calculateCounterValue(`name1`, `250`)
	assert.Equal(t, int64(250), value)

	err := r.Create(models.Metric{Type: models.CounterType, Name: `name1`, ValueInt64: 500})
	require.NoError(t, err)

	value = s.calculateCounterValue(`name1`, `250`)
	assert.Equal(t, int64(750), value)
}

func TestGetCounterMetricValue(t *testing.T) {
	r := storage.New()
	s := New(r)

	_, err := s.GetMetricValue(models.CounterType, "test")
	require.Error(t, err)

	err = s.Save(models.CounterType, "test", "123")
	require.NoError(t, err)
	value, err := s.GetMetricValue(models.CounterType, "test")
	require.NoError(t, err)
	assert.Equal(t, "123", value)
}

func TestGetAll(t *testing.T) {
	r := storage.New()
	s := New(r)

	err := r.Create(models.Metric{Type: models.CounterType, Name: "name1", ValueInt64: 500})
	require.NoError(t, err)
	err = r.Create(models.Metric{Type: models.GaugeType, Name: "name1", ValueFloat64: 500})
	require.NoError(t, err)

	vals, err := s.GetAll()
	require.NoError(t, err)
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, int64(1000),
		vals[0].ValueInt64+int64(vals[0].ValueFloat64)+vals[1].ValueInt64+int64(vals[1].ValueFloat64))
}
