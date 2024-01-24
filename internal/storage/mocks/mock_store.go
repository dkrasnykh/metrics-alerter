package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/dkrasnykh/metrics-alerter/internal/models"
	gomock "github.com/golang/mock/gomock"
)

// MockStorager is a mock of Storager interface.
type MockStorager struct {
	ctrl     *gomock.Controller
	recorder *MockStoragerMockRecorder
}

// MockStoragerMockRecorder is the mock recorder for MockStorager.
type MockStoragerMockRecorder struct {
	mock *MockStorager
}

// NewMockStorager creates a new mock instance.
func NewMockStorager(ctrl *gomock.Controller) *MockStorager {
	mock := &MockStorager{ctrl: ctrl}
	mock.recorder = &MockStoragerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorager) EXPECT() *MockStoragerMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockStorager) Create(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, metric)
	ret0, _ := ret[0].(models.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockStoragerMockRecorder) Create(ctx, metric interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockStorager)(nil).Create), ctx, metric)
}

// Get mocks base method.
func (m *MockStorager) Get(ctx context.Context, mType, name string) (models.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, mType, name)
	ret0, _ := ret[0].(models.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStoragerMockRecorder) Get(ctx, mType, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockStorager)(nil).Get), ctx, mType, name)
}

// GetAll mocks base method.
func (m *MockStorager) GetAll(ctx context.Context) ([]models.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx)
	ret0, _ := ret[0].([]models.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockStoragerMockRecorder) GetAll(ctx context.Context) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockStorager)(nil).GetAll), ctx)
}

// Load mocks base method.
func (m *MockStorager) Load(ctx context.Context, metrics []models.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Load", ctx, metrics)
	ret0, _ := ret[0].(error)
	return ret0
}

// Load indicates an expected call of Load.
func (mr *MockStoragerMockRecorder) Load(ctx, metrics interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Load", reflect.TypeOf((*MockStorager)(nil).Load), ctx, metrics)
}

// Ping mocks base method.
func (m *MockStorager) Ping(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockStoragerMockRecorder) Ping(ctx context.Context) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockStorager)(nil).Ping), ctx)
}
