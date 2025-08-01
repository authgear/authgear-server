// Code generated by MockGen. DO NOT EDIT.
// Source: pool.go

// Package db is a generated GoMock package.
package db

import (
	reflect "reflect"

	oteldatabasesql "github.com/authgear/authgear-server/pkg/util/otelutil/oteldatabasesql"
	gomock "github.com/golang/mock/gomock"
)

// MockPool_ is a mock of Pool_ interface.
type MockPool_ struct {
	ctrl     *gomock.Controller
	recorder *MockPool_MockRecorder
}

// MockPool_MockRecorder is the mock recorder for MockPool_.
type MockPool_MockRecorder struct {
	mock *MockPool_
}

// NewMockPool_ creates a new mock instance.
func NewMockPool_(ctrl *gomock.Controller) *MockPool_ {
	mock := &MockPool_{ctrl: ctrl}
	mock.recorder = &MockPool_MockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPool_) EXPECT() *MockPool_MockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockPool_) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockPool_MockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockPool_)(nil).Close))
}

// Open mocks base method.
func (m *MockPool_) Open(info ConnectionInfo, opts ConnectionOptions) (oteldatabasesql.ConnPool_, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Open", info, opts)
	ret0, _ := ret[0].(oteldatabasesql.ConnPool_)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Open indicates an expected call of Open.
func (mr *MockPool_MockRecorder) Open(info, opts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Open", reflect.TypeOf((*MockPool_)(nil).Open), info, opts)
}
