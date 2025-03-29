package mocks

import (
	"reflect"

	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/golang/mock/gomock"
)

// MockLogger is a mock of Logger interface
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Debug mocks base method
func (m *MockLogger) Debug(msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", msg)
}

// Debug indicates an expected call of Debug
func (mr *MockLoggerMockRecorder) Debug(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogger)(nil).Debug), msg)
}

// Info mocks base method
func (m *MockLogger) Info(msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", msg)
}

// Info indicates an expected call of Info
func (mr *MockLoggerMockRecorder) Info(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogger)(nil).Info), msg)
}

// Warn mocks base method
func (m *MockLogger) Warn(msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Warn", msg)
}

// Warn indicates an expected call of Warn
func (mr *MockLoggerMockRecorder) Warn(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLogger)(nil).Warn), msg)
}

// Error mocks base method
func (m *MockLogger) Error(msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", msg)
}

// Error indicates an expected call of Error
func (mr *MockLoggerMockRecorder) Error(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogger)(nil).Error), msg)
}

// Fatal mocks base method
func (m *MockLogger) Fatal(msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Fatal", msg)
}

// Fatal indicates an expected call of Fatal
func (mr *MockLoggerMockRecorder) Fatal(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatal", reflect.TypeOf((*MockLogger)(nil).Fatal), msg)
}

// WithField mocks base method
func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithField", key, value)
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// WithField indicates an expected call of WithField
func (mr *MockLoggerMockRecorder) WithField(key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithField", reflect.TypeOf((*MockLogger)(nil).WithField), key, value)
}

// WithFields mocks base method
func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithFields", fields)
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// WithFields indicates an expected call of WithFields
func (mr *MockLoggerMockRecorder) WithFields(fields interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithFields", reflect.TypeOf((*MockLogger)(nil).WithFields), fields)
}
