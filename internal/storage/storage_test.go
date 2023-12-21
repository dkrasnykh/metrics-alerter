package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

var (
	mCounter = models.Metric{Type: models.CounterType, Name: "name1", ValueInt64: 500}
	mGauge   = models.Metric{Type: models.GaugeType, Name: "name1", ValueFloat64: 500}
)

func TestCreate(t *testing.T) {
	s := New()
	err := s.Create(mCounter)
	require.NoError(t, err)
}

func TestGet(t *testing.T) {
	s := New()
	err := s.Create(mCounter)
	require.NoError(t, err)

	value, err := s.Get(models.CounterType, "name1")
	require.NoError(t, err)
	assert.Equal(t, mCounter, value)

	_, err = s.Get(models.CounterType, "name2")
	require.Error(t, err)
}

func TestGetAll(t *testing.T) {
	s := New()
	err := s.Create(mCounter)
	require.NoError(t, err)
	err = s.Create(mGauge)
	require.NoError(t, err)

	vals, err := s.GetAll()
	require.NoError(t, err)
	assert.Equal(t, 2, len(vals))
	assert.Equal(t, int64(1000),
		vals[0].ValueInt64+int64(vals[0].ValueFloat64)+vals[1].ValueInt64+int64(vals[1].ValueFloat64))
}

func TestUpdate(t *testing.T) {
	s := New()
	err := s.Create(mCounter)
	require.NoError(t, err)
	value, err := s.Get(models.CounterType, `name1`)
	require.NoError(t, err)
	assert.Equal(t, mCounter.ValueInt64, value.ValueInt64)

	updated := models.Metric{Type: models.CounterType, Name: `name1`, ValueInt64: 450}
	err = s.Update(updated)
	require.NoError(t, err)
	value, err = s.Get(models.CounterType, `name1`)
	require.NoError(t, err)
	assert.Equal(t, int64(450), value.ValueInt64)
}

func TestDelete(t *testing.T) {
	s := New()
	err := s.Create(mCounter)
	require.NoError(t, err)

	_, err = s.Get(mCounter.Type, mCounter.Name)
	require.NoError(t, err)

	err = s.Delete(mCounter.Type, mCounter.Name)
	require.NoError(t, err)

	_, err = s.Get(mCounter.Type, mCounter.Name)
	require.Error(t, err)
}
