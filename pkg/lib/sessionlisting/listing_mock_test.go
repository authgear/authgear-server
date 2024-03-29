// Code generated by MockGen. DO NOT EDIT.
// Source: listing.go

// Package sessionlisting_test is a generated GoMock package.
package sessionlisting_test

import (
	reflect "reflect"
	time "time"

	oauth "github.com/authgear/authgear-server/pkg/lib/oauth"
	idpsession "github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	gomock "github.com/golang/mock/gomock"
)

// MockIDPSessionProvider is a mock of IDPSessionProvider interface.
type MockIDPSessionProvider struct {
	ctrl     *gomock.Controller
	recorder *MockIDPSessionProviderMockRecorder
}

// MockIDPSessionProviderMockRecorder is the mock recorder for MockIDPSessionProvider.
type MockIDPSessionProviderMockRecorder struct {
	mock *MockIDPSessionProvider
}

// NewMockIDPSessionProvider creates a new mock instance.
func NewMockIDPSessionProvider(ctrl *gomock.Controller) *MockIDPSessionProvider {
	mock := &MockIDPSessionProvider{ctrl: ctrl}
	mock.recorder = &MockIDPSessionProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIDPSessionProvider) EXPECT() *MockIDPSessionProviderMockRecorder {
	return m.recorder
}

// CheckSessionExpired mocks base method.
func (m *MockIDPSessionProvider) CheckSessionExpired(session *idpsession.IDPSession) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckSessionExpired", session)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckSessionExpired indicates an expected call of CheckSessionExpired.
func (mr *MockIDPSessionProviderMockRecorder) CheckSessionExpired(session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckSessionExpired", reflect.TypeOf((*MockIDPSessionProvider)(nil).CheckSessionExpired), session)
}

// MockOfflineGrantService is a mock of OfflineGrantService interface.
type MockOfflineGrantService struct {
	ctrl     *gomock.Controller
	recorder *MockOfflineGrantServiceMockRecorder
}

// MockOfflineGrantServiceMockRecorder is the mock recorder for MockOfflineGrantService.
type MockOfflineGrantServiceMockRecorder struct {
	mock *MockOfflineGrantService
}

// NewMockOfflineGrantService creates a new mock instance.
func NewMockOfflineGrantService(ctrl *gomock.Controller) *MockOfflineGrantService {
	mock := &MockOfflineGrantService{ctrl: ctrl}
	mock.recorder = &MockOfflineGrantServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOfflineGrantService) EXPECT() *MockOfflineGrantServiceMockRecorder {
	return m.recorder
}

// CheckSessionExpired mocks base method.
func (m *MockOfflineGrantService) CheckSessionExpired(session *oauth.OfflineGrant) (bool, time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckSessionExpired", session)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CheckSessionExpired indicates an expected call of CheckSessionExpired.
func (mr *MockOfflineGrantServiceMockRecorder) CheckSessionExpired(session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckSessionExpired", reflect.TypeOf((*MockOfflineGrantService)(nil).CheckSessionExpired), session)
}
