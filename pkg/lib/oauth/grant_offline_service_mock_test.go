// Code generated by MockGen. DO NOT EDIT.
// Source: grant_offline_service.go

// Package oauth is a generated GoMock package.
package oauth

import (
	context "context"
	reflect "reflect"
	time "time"

	access "github.com/authgear/authgear-server/pkg/lib/session/access"
	idpsession "github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	gomock "github.com/golang/mock/gomock"
)

// MockServiceIDPSessionProvider is a mock of ServiceIDPSessionProvider interface.
type MockServiceIDPSessionProvider struct {
	ctrl     *gomock.Controller
	recorder *MockServiceIDPSessionProviderMockRecorder
}

// MockServiceIDPSessionProviderMockRecorder is the mock recorder for MockServiceIDPSessionProvider.
type MockServiceIDPSessionProviderMockRecorder struct {
	mock *MockServiceIDPSessionProvider
}

// NewMockServiceIDPSessionProvider creates a new mock instance.
func NewMockServiceIDPSessionProvider(ctrl *gomock.Controller) *MockServiceIDPSessionProvider {
	mock := &MockServiceIDPSessionProvider{ctrl: ctrl}
	mock.recorder = &MockServiceIDPSessionProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockServiceIDPSessionProvider) EXPECT() *MockServiceIDPSessionProviderMockRecorder {
	return m.recorder
}

// CheckSessionExpired mocks base method.
func (m *MockServiceIDPSessionProvider) CheckSessionExpired(session *idpsession.IDPSession) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckSessionExpired", session)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckSessionExpired indicates an expected call of CheckSessionExpired.
func (mr *MockServiceIDPSessionProviderMockRecorder) CheckSessionExpired(session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckSessionExpired", reflect.TypeOf((*MockServiceIDPSessionProvider)(nil).CheckSessionExpired), session)
}

// Get mocks base method.
func (m *MockServiceIDPSessionProvider) Get(ctx context.Context, id string) (*idpsession.IDPSession, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id)
	ret0, _ := ret[0].(*idpsession.IDPSession)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockServiceIDPSessionProviderMockRecorder) Get(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockServiceIDPSessionProvider)(nil).Get), ctx, id)
}

// MockOfflineGrantServiceAccessEventProvider is a mock of OfflineGrantServiceAccessEventProvider interface.
type MockOfflineGrantServiceAccessEventProvider struct {
	ctrl     *gomock.Controller
	recorder *MockOfflineGrantServiceAccessEventProviderMockRecorder
}

// MockOfflineGrantServiceAccessEventProviderMockRecorder is the mock recorder for MockOfflineGrantServiceAccessEventProvider.
type MockOfflineGrantServiceAccessEventProviderMockRecorder struct {
	mock *MockOfflineGrantServiceAccessEventProvider
}

// NewMockOfflineGrantServiceAccessEventProvider creates a new mock instance.
func NewMockOfflineGrantServiceAccessEventProvider(ctrl *gomock.Controller) *MockOfflineGrantServiceAccessEventProvider {
	mock := &MockOfflineGrantServiceAccessEventProvider{ctrl: ctrl}
	mock.recorder = &MockOfflineGrantServiceAccessEventProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOfflineGrantServiceAccessEventProvider) EXPECT() *MockOfflineGrantServiceAccessEventProviderMockRecorder {
	return m.recorder
}

// RecordAccess mocks base method.
func (m *MockOfflineGrantServiceAccessEventProvider) RecordAccess(ctx context.Context, sessionID string, expiry time.Time, event *access.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordAccess", ctx, sessionID, expiry, event)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecordAccess indicates an expected call of RecordAccess.
func (mr *MockOfflineGrantServiceAccessEventProviderMockRecorder) RecordAccess(ctx, sessionID, expiry, event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordAccess", reflect.TypeOf((*MockOfflineGrantServiceAccessEventProvider)(nil).RecordAccess), ctx, sessionID, expiry, event)
}

// MockOfflineGrantServiceMeterService is a mock of OfflineGrantServiceMeterService interface.
type MockOfflineGrantServiceMeterService struct {
	ctrl     *gomock.Controller
	recorder *MockOfflineGrantServiceMeterServiceMockRecorder
}

// MockOfflineGrantServiceMeterServiceMockRecorder is the mock recorder for MockOfflineGrantServiceMeterService.
type MockOfflineGrantServiceMeterServiceMockRecorder struct {
	mock *MockOfflineGrantServiceMeterService
}

// NewMockOfflineGrantServiceMeterService creates a new mock instance.
func NewMockOfflineGrantServiceMeterService(ctrl *gomock.Controller) *MockOfflineGrantServiceMeterService {
	mock := &MockOfflineGrantServiceMeterService{ctrl: ctrl}
	mock.recorder = &MockOfflineGrantServiceMeterServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOfflineGrantServiceMeterService) EXPECT() *MockOfflineGrantServiceMeterServiceMockRecorder {
	return m.recorder
}

// TrackActiveUser mocks base method.
func (m *MockOfflineGrantServiceMeterService) TrackActiveUser(ctx context.Context, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TrackActiveUser", ctx, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// TrackActiveUser indicates an expected call of TrackActiveUser.
func (mr *MockOfflineGrantServiceMeterServiceMockRecorder) TrackActiveUser(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TrackActiveUser", reflect.TypeOf((*MockOfflineGrantServiceMeterService)(nil).TrackActiveUser), ctx, userID)
}
