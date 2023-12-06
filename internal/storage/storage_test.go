package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreate(t *testing.T) {
	s := NewStorage()
	err := s.Create(Gauge, "name1", "123.0")
	require.NoError(t, err)
}

func TestGet(t *testing.T) {
	s := NewStorage()
	err := s.Create(Counter, "name1", "123")
	require.NoError(t, err)

	value, err := s.Get(Counter, "name1")
	require.NoError(t, err)
	assert.Equal(t, "123", value)

	_, err = s.Get(Counter, "name2")
	require.Error(t, err)
}

func TestGetAll(t *testing.T) {
	s := NewStorage()
	err := s.Create(Counter, "name1", "1")
	require.NoError(t, err)
	err = s.Create(Gauge, "name2", "2")
	require.NoError(t, err)

	values, err := s.GetAll()
	require.NoError(t, err)
	assert.Equal(t, 2, len(values))
	assert.Equal(t, 1, len(values[Counter]))
	assert.Equal(t, 1, len(values[Gauge]))
}

func TestUpdate(t *testing.T) {
	s := NewStorage()
	err := s.Create(Gauge, "name1", "123.0")
	require.NoError(t, err)

	value, err := s.Get(Gauge, "name1")
	require.NoError(t, err)
	assert.Equal(t, "123.0", value)

	err = s.Update(Gauge, "name1", "255.0")
	require.NoError(t, err)

	value, err = s.Get(Gauge, "name1")
	require.NoError(t, err)
	assert.Equal(t, "255.0", value)

}

func TestDelete(t *testing.T) {
	s := NewStorage()
	err := s.Create(Gauge, "name1", "123.0")
	require.NoError(t, err)

	value, err := s.Get(Gauge, "name1")
	require.NoError(t, err)
	assert.Equal(t, "123.0", value)

	err = s.Delete(Gauge, "name1")
	require.NoError(t, err)

	_, err = s.Get(Gauge, "name1")
	require.Error(t, err)
}
