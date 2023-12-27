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

	delta := int64(10)
	value := float64(100)

	err := s.Validate(models.Metrics{MType: models.CounterType, ID: `test`, Delta: &delta})
	require.NoError(t, err)

	err = s.Validate(models.Metrics{MType: models.CounterType, ID: `test`, Value: &value})
	require.Error(t, err)

	err = s.Validate(models.Metrics{MType: models.CounterType, ID: ``, Delta: &delta})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrIDIsEmpty))

	err = s.Validate(models.Metrics{MType: models.GaugeType, ID: `test`, Value: &value})
	require.NoError(t, err)

	err = s.Validate(models.Metrics{MType: models.GaugeType, ID: `test`, Delta: &delta})
	require.Error(t, err)

	err = s.Validate(models.Metrics{MType: `unknown`, ID: `test`, Value: &value})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnknownMetricType))
}

func TestSave(t *testing.T) {
	r := storage.New()
	s := New(r)
	value := float64(100)
	m := models.Metrics{MType: models.GaugeType, ID: `test`, Value: &value}
	saved, err := s.Save(m)
	require.NoError(t, err)
	assert.Equal(t, m, saved)
}

func TestCalculateCounterValue(t *testing.T) {
	r := storage.New()
	s := New(r)

	value := s.calculateCounterValue(`name1`, 250)
	assert.Equal(t, int64(250), value)
	delta := int64(500)
	_, err := r.Create(models.Metrics{MType: models.CounterType, ID: `name1`, Delta: &delta})
	require.NoError(t, err)

	value = s.calculateCounterValue(`name1`, 250)
	assert.Equal(t, int64(750), value)
}

func TestGetCounterMetricValue(t *testing.T) {
	r := storage.New()
	s := New(r)

	_, err := s.GetMetricValue(models.CounterType, "test")
	require.Error(t, err)

	delta := int64(123)
	_, err = s.Save(models.Metrics{MType: models.CounterType, ID: `test`, Delta: &delta})
	require.NoError(t, err)
	value, err := s.GetMetricValue(models.CounterType, "test")
	require.NoError(t, err)
	assert.Equal(t, "123", value)
}

func TestGetAll(t *testing.T) {
	r := storage.New()
	s := New(r)
	delta := int64(500)
	value := float64(500)

	_, err := r.Create(models.Metrics{MType: models.CounterType, ID: "name1", Delta: &delta})
	require.NoError(t, err)
	_, err = r.Create(models.Metrics{MType: models.GaugeType, ID: "name1", Value: &value})
	require.NoError(t, err)

	vals, err := s.GetAll()
	require.NoError(t, err)
	assert.Equal(t, 2, len(vals))
}
