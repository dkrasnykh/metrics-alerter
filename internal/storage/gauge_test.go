package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

func TestCreateGauge(t *testing.T) {
	s := NewGaugeStorage()
	err := s.Create("name1", float64(123))
	require.NoError(t, err)
}

func TestGetGauge(t *testing.T) {
	s := NewGaugeStorage()
	expected := float64(123)
	err := s.Create("name1", expected)
	require.NoError(t, err)

	value, err := s.Get("name1")
	require.NoError(t, err)
	assert.Equal(t, expected, value)

	_, err = s.Get("name2")
	require.Error(t, err)
}

func TestGetAllGauge(t *testing.T) {
	s := NewGaugeStorage()
	err := s.Create("name1", float64(1))
	require.NoError(t, err)
	err = s.Create("name2", float64(2))
	require.NoError(t, err)

	valsAny, err := s.GetAll()
	require.NoError(t, err)

	values, ok := valsAny.([]models.Gauge)
	assert.True(t, true, ok)

	assert.Equal(t, 2, len(values))
	assert.Equal(t, float64(3), values[0].Value+values[1].Value)
}

func TestUpdateGuage(t *testing.T) {
	s := NewGaugeStorage()

	err := s.Create("name1", float64(123))
	require.NoError(t, err)
	valueAny, err := s.Get("name1")
	require.NoError(t, err)
	value, ok := valueAny.(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(123), value)

	err = s.Update("name1", float64(255))
	require.NoError(t, err)
	valueAny, err = s.Get("name1")
	value, ok = valueAny.(float64)
	assert.True(t, ok)
	require.NoError(t, err)
	assert.Equal(t, float64(255), value)
}

func TestDeleteGuage(t *testing.T) {
	s := NewGaugeStorage()
	err := s.Create("name1", float64(123))
	require.NoError(t, err)

	_, err = s.Get("name1")
	require.NoError(t, err)

	err = s.Delete("name1")
	require.NoError(t, err)

	_, err = s.Get("name1")
	require.Error(t, err)
}
