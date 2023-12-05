package service

import (
	"github.com/dkrasnykh/metrics-alerter/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateAndSave(t *testing.T) {
	r := storage.NewStorage()
	s := NewService(r)
	err := s.ValidateAndSave(storage.Gauge, "test", "123.0")
	require.NoError(t, err)

	err = s.ValidateAndSave(storage.Gauge, "test", "test")
	require.Error(t, err)

	err = s.ValidateAndSave("unknown", "test", "123")
	require.Error(t, err)

	err = s.ValidateAndSave(storage.Counter, "test", "155")
	require.NoError(t, err)

	err = s.ValidateAndSave(storage.Counter, "test", "123.0")
	require.Error(t, err)

	err = s.ValidateAndSave(storage.Counter, "test", "test")
	require.Error(t, err)
}

func TestValidateGaudeValue(t *testing.T) {
	r := storage.NewStorage()
	s := NewService(r)
	err := s.validateGaudeValue("123")
	require.NoError(t, err)

	err = s.validateGaudeValue("123.0")
	require.NoError(t, err)

	err = s.validateGaudeValue("test")
	require.Error(t, err)

	//1.8*10^308
	err = s.validateGaudeValue("9999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999")
	require.Error(t, err)
}

func TestSaveGaudeValue(t *testing.T) {
	r := storage.NewStorage()
	s := NewService(r)
	err := s.saveGaudeValue("test", "123.0")
	require.NoError(t, err)
	value, err := r.Get(storage.Gauge, "test")
	assert.Equal(t, "123.0", value)

	err = s.saveGaudeValue("test", "155.0")
	require.NoError(t, err)
	value, err = r.Get(storage.Gauge, "test")
	assert.Equal(t, "155.0", value)
}

func TestValidateCounterValue(t *testing.T) {
	r := storage.NewStorage()
	s := NewService(r)
	err := s.validateCounterValue("123")
	require.NoError(t, err)

	err = s.validateCounterValue("123.0")
	require.Error(t, err)

	err = s.validateCounterValue("test")
	require.Error(t, err)

	err = s.validateCounterValue("10000000000000000000")
	require.Error(t, err)
}

func TestSaveCounterValue(t *testing.T) {
	r := storage.NewStorage()
	s := NewService(r)
	err := s.saveCounterValue("test", "123")
	require.NoError(t, err)
	value, err := r.Get(storage.Counter, "test")
	assert.Equal(t, "123", value)

	err = s.saveCounterValue("test", "15")
	require.NoError(t, err)
	value, err = r.Get(storage.Counter, "test")
	assert.Equal(t, "138", value)
}
