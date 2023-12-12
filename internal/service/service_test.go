package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
)

func TestValidateAndSave(t *testing.T) {
	cr := storage.NewCounterStorage()
	gr := storage.NewGaugeStorage()
	s := NewService(cr, gr)

	err := s.ValidateAndSave(models.GaugeType, "test", "123.0")
	require.NoError(t, err)

	err = s.ValidateAndSave(models.GaugeType, "test", "test")
	require.Error(t, err)

	err = s.ValidateAndSave("unknown", "test", "123")
	require.Error(t, err)

	err = s.ValidateAndSave(models.CounterType, "test", "155")
	require.NoError(t, err)

	err = s.ValidateAndSave(models.CounterType, "test", "123.0")
	require.Error(t, err)

	err = s.ValidateAndSave(models.CounterType, "test", "test")
	require.Error(t, err)
}

func TestSaveCounterValue(t *testing.T) {
	cr := storage.NewCounterStorage()
	gr := storage.NewGaugeStorage()
	s := NewService(cr, gr)

	err := s.saveCounterValue("test", int64(123))
	require.NoError(t, err)
	valueAny, err := cr.Get("test")
	require.NoError(t, err)
	value, ok := valueAny.(int64)
	assert.True(t, ok)
	assert.Equal(t, int64(123), value)

	err = s.saveCounterValue("test", int64(15))
	require.NoError(t, err)
	valueAny, err = cr.Get("test")
	require.NoError(t, err)
	value, ok = valueAny.(int64)
	assert.True(t, ok)
	assert.Equal(t, int64(138), value)
}

func TestGetCounterMetricValue(t *testing.T) {
	cr := storage.NewCounterStorage()
	gr := storage.NewGaugeStorage()
	s := NewService(cr, gr)

	_, err := s.GetMetricValue(models.CounterType, "test")
	require.Error(t, err)

	err = s.saveCounterValue("test", int64(123))
	require.NoError(t, err)
	value, err := s.GetMetricValue(models.CounterType, "test")
	require.NoError(t, err)
	assert.Equal(t, "123", value)
}

func TestGetGaugeMetricValue(t *testing.T) {
	cr := storage.NewCounterStorage()
	gr := storage.NewGaugeStorage()
	s := NewService(cr, gr)

	_, err := s.GetMetricValue(models.GaugeType, "test")
	require.Error(t, err)

	err = s.ValidateAndSave(models.GaugeType, "test", "123")
	require.NoError(t, err)
	value, err := s.GetMetricValue(models.GaugeType, "test")
	require.NoError(t, err)
	assert.Equal(t, "123", value)
}

func TestGetAllCounter(t *testing.T) {
	cr := storage.NewCounterStorage()
	gr := storage.NewGaugeStorage()
	s := NewService(cr, gr)

	value, err := s.GetAllCounter()
	require.NoError(t, err)
	assert.Equal(t, 0, len(value))

	err = s.ValidateAndSave(models.CounterType, "testCounter", "123")
	require.NoError(t, err)

	value, err = s.GetAllCounter()
	require.NoError(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, "testCounter", value[0].Name)
	assert.Equal(t, int64(123), value[0].Value)
}

func TestGetAllGauge(t *testing.T) {
	cr := storage.NewCounterStorage()
	gr := storage.NewGaugeStorage()
	s := NewService(cr, gr)

	value, err := s.GetAllGauge()
	require.NoError(t, err)
	assert.Equal(t, 0, len(value))

	err = s.ValidateAndSave(models.GaugeType, "testGauge", "123")
	require.NoError(t, err)

	value, err = s.GetAllGauge()
	require.NoError(t, err)
	assert.Equal(t, 1, len(value))
	assert.Equal(t, "testGauge", value[0].Name)
	assert.Equal(t, float64(123), value[0].Value)
}
