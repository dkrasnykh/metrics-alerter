package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

var (
	mdelta   = int64(500)
	mvalue   = float64(500)
	mCounter = models.Metrics{MType: models.CounterType, ID: "name1", Delta: &mdelta}
	mGauge   = models.Metrics{MType: models.GaugeType, ID: "name1", Value: &mvalue}
)

func TestCreate(t *testing.T) {
	_ = logger.InitLogger()
	ctx := context.Background()
	s := New("", 0)
	m, err := s.Create(ctx, mCounter)
	require.NoError(t, err)
	assert.Equal(t, mCounter, m)
}

func TestGet(t *testing.T) {
	_ = logger.InitLogger()
	ctx := context.Background()
	s := New("", 0)
	_, err := s.Create(ctx, mCounter)
	require.NoError(t, err)

	value, err := s.Get(ctx, models.CounterType, "name1")
	require.NoError(t, err)
	assert.Equal(t, mCounter, value)

	_, err = s.Get(ctx, models.CounterType, "name2")
	require.Error(t, err)
}

func TestGetAll(t *testing.T) {
	_ = logger.InitLogger()
	ctx := context.Background()
	s := New("", 0)
	_, err := s.Create(ctx, mCounter)
	require.NoError(t, err)
	_, err = s.Create(ctx, mGauge)
	require.NoError(t, err)

	vals, err := s.GetAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, len(vals))
}

func TestUpdate(t *testing.T) {
	_ = logger.InitLogger()
	ctx := context.Background()
	s := New("", 0)
	_, err := s.Create(ctx, mCounter)
	require.NoError(t, err)
	value, err := s.Get(ctx, models.CounterType, `name1`)
	require.NoError(t, err)
	assert.Equal(t, *mCounter.Delta, *value.Delta)

	delta := int64(450)
	updated := models.Metrics{MType: models.CounterType, ID: `name1`, Delta: &delta}
	_, err = s.Create(ctx, updated)
	require.NoError(t, err)
	value, err = s.Get(ctx, models.CounterType, `name1`)
	require.NoError(t, err)
	assert.Equal(t, delta, *value.Delta)
}
