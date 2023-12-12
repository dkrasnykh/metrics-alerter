package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

func TestCreateCounter(t *testing.T) {
	s := NewCounterStorage()
	err := s.Create("name1", int64(123))
	require.NoError(t, err)
}

func TestGetCounter(t *testing.T) {
	s := NewCounterStorage()
	expected := int64(123)
	err := s.Create("name1", expected)
	require.NoError(t, err)

	value, err := s.Get("name1")
	require.NoError(t, err)
	assert.Equal(t, expected, value)

	_, err = s.Get("name2")
	require.Error(t, err)
}

func TestGetAllCounter(t *testing.T) {
	s := NewCounterStorage()
	err := s.Create("name1", int64(1))
	require.NoError(t, err)
	err = s.Create("name2", int64(2))
	require.NoError(t, err)

	valsAny, err := s.GetAll()
	require.NoError(t, err)

	values, ok := valsAny.([]models.Counter)
	assert.True(t, ok)

	assert.Equal(t, 2, len(values))
	assert.Equal(t, int64(3), values[0].Value+values[1].Value)
}

func TestUpdateCounter(t *testing.T) {
	s := NewCounterStorage()

	err := s.Create("name1", int64(123))
	require.NoError(t, err)
	valueAny, err := s.Get("name1")
	require.NoError(t, err)
	value, ok := valueAny.(int64)
	assert.True(t, ok)
	assert.Equal(t, int64(123), value)

	err = s.Update("name1", int64(255))
	require.NoError(t, err)
	valueAny, err = s.Get("name1")
	value, ok = valueAny.(int64)
	assert.True(t, ok)
	require.NoError(t, err)
	assert.Equal(t, int64(255), value)
}

func TestDeleteCounter(t *testing.T) {
	s := NewCounterStorage()
	err := s.Create("name1", int64(123))
	require.NoError(t, err)

	_, err = s.Get("name1")
	require.NoError(t, err)

	err = s.Delete("name1")
	require.NoError(t, err)

	_, err = s.Get("name1")
	require.Error(t, err)
}
