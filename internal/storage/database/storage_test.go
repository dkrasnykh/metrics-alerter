package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/dkrasnykh/metrics-alerter/internal/logger"
	"github.com/dkrasnykh/metrics-alerter/internal/models"
)

var ErrTest = errors.New("database access error")

func TestCreate(t *testing.T) {
	_ = logger.InitLogger()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	r := Storage{db: sqlxDB}
	ctx := context.Background()

	type args struct {
		ctx context.Context
		m   models.Metrics
	}
	type mockBehavior func(args args)
	delta := int64(500)
	value := float64(500)
	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		wantErr bool
	}{
		{
			name: "ok create counter",
			mock: func(args args) {
				mock.ExpectExec("INSERT INTO metrics").WithArgs(args.m.ID, args.m.MType, *args.m.Delta).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input: args{ctx: ctx, m: models.Metrics{MType: models.CounterType, ID: "name1", Delta: &delta}},
		},
		{
			name: "ok create gauge",
			mock: func(args args) {
				mock.ExpectExec("INSERT INTO metrics").WithArgs(args.m.ID, args.m.MType, *args.m.Value).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			input: args{ctx: ctx, m: models.Metrics{MType: models.GaugeType, ID: "name1", Value: &value}},
		},
		{
			name: "insertion error",
			mock: func(args args) {
				mock.ExpectExec("INSERT INTO metrics").WithArgs(args.m.ID, args.m.MType, *args.m.Value).
					WillReturnError(ErrTest)
			},
			input:   args{ctx: ctx, m: models.Metrics{MType: models.GaugeType, ID: "name1", Value: &value}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input)

			got, err := r.Create(tt.input.ctx, tt.input.m)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.input.m, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
	_ = mockDB.Close()
}

func TestGet(t *testing.T) {
	_ = logger.InitLogger()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	r := Storage{db: sqlxDB}
	ctx := context.Background()

	type args struct {
		ctx   context.Context
		mType string
		mID   string
		delta int64
		value float64
	}
	delta := int64(500)
	value := float64(500)
	type mockBehavior func(a args)
	tests := []struct {
		name    string
		mock    mockBehavior
		input   args
		want    models.Metrics
		wantErr bool
	}{
		{
			name: "ok conter",
			mock: func(a args) {
				rows := sqlmock.NewRows([]string{"delta", "value"}).AddRow(a.delta, nil)
				mock.ExpectQuery("select (.+) from metrics where (.+) ORDER BY time DESC LIMIT 1;").
					WithArgs(a.mID, a.mType).WillReturnRows(rows)
			},
			input: args{
				ctx:   ctx,
				mType: models.CounterType,
				mID:   "name1",
				delta: int64(500),
			},
			want: models.Metrics{MType: models.CounterType, ID: "name1", Delta: &delta},
		},
		{
			name: "ok gauge",
			mock: func(a args) {
				rows := sqlmock.NewRows([]string{"delta", "value"}).AddRow(nil, a.value)
				mock.ExpectQuery("select (.+) from metrics where (.+) ORDER BY time DESC LIMIT 1;").
					WithArgs(a.mID, a.mType).WillReturnRows(rows)
			},
			input: args{
				ctx:   ctx,
				mType: models.GaugeType,
				mID:   "name1",
				value: float64(500),
			},
			want: models.Metrics{MType: models.GaugeType, ID: "name1", Value: &value},
		},
		{
			name: "selection error",
			mock: func(a args) {
				mock.ExpectQuery("select (.+) from metrics where (.+) ORDER BY time DESC LIMIT 1;").
					WithArgs(a.mID, a.mType).WillReturnError(ErrTest)
			},
			input: args{
				ctx:   ctx,
				mType: models.GaugeType,
				mID:   "name1",
				value: float64(500),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock(tt.input)

			got, err := r.Get(tt.input.ctx, tt.input.mType, tt.input.mID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
	_ = mockDB.Close()
}

func TestGetAll(t *testing.T) {
	_ = logger.InitLogger()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	r := Storage{db: sqlxDB}
	ctx := context.Background()

	delta := int64(500)
	value := float64(500)
	counter := models.Metrics{MType: models.CounterType, ID: "name1", Delta: &delta}
	gauge := models.Metrics{MType: models.GaugeType, ID: "name1", Value: &value}
	type mockBehavior func()
	tests := []struct {
		name    string
		mock    mockBehavior
		input   context.Context
		want    []models.Metrics
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				rows := sqlmock.NewRows([]string{"name", "type", "delta", "value"}).
					AddRow("name1", "counter", int64(500), nil).
					AddRow("name1", "gauge", nil, float64(500))
				mock.ExpectQuery(`SELECT (.+) FROM (.+) AS t1 LEFT JOIN metrics AS m ON (.+);`).
					WithoutArgs().WillReturnRows(rows)
			},
			input: ctx,
			want:  []models.Metrics{counter, gauge},
		},
		{
			name: "selection error",
			mock: func() {
				mock.ExpectQuery(`SELECT (.+) FROM (.+) AS t1 LEFT JOIN metrics AS m ON (.+);`).
					WithoutArgs().WillReturnError(ErrTest)
			},
			input:   ctx,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			got, err := r.GetAll(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
	_ = mockDB.Close()
}

func TestLoad(t *testing.T) {
	_ = logger.InitLogger()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	r := Storage{db: sqlxDB}
	ctx := context.Background()

	delta := int64(500)
	value := float64(500)
	counter := models.Metrics{MType: models.CounterType, ID: "name1", Delta: &delta}
	gauge := models.Metrics{MType: models.GaugeType, ID: "name1", Value: &value}

	type mockBehavior func()
	tests := []struct {
		name    string
		mock    mockBehavior
		input   context.Context
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO metrics").WithArgs("name1", models.CounterType, delta).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO metrics").WithArgs("name1", models.GaugeType, value).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			input: ctx,
		},
		{
			name: "insertion error",
			mock: func() {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO metrics").WithArgs("name1", models.CounterType, delta).
					WillReturnError(ErrTest)
				mock.ExpectRollback()
			},
			input: ctx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			err := r.Load(tt.input, []models.Metrics{counter, gauge})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
	_ = mockDB.Close()
}
