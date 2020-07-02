// Code generated by MockGen. DO NOT EDIT.
// Source: resolve.go

// Package internalserver is a generated GoMock package.
package internalserver

import (
	anonymous "github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockAnonymousIdentityProvider is a mock of AnonymousIdentityProvider interface
type MockAnonymousIdentityProvider struct {
	ctrl     *gomock.Controller
	recorder *MockAnonymousIdentityProviderMockRecorder
}

// MockAnonymousIdentityProviderMockRecorder is the mock recorder for MockAnonymousIdentityProvider
type MockAnonymousIdentityProviderMockRecorder struct {
	mock *MockAnonymousIdentityProvider
}

// NewMockAnonymousIdentityProvider creates a new mock instance
func NewMockAnonymousIdentityProvider(ctrl *gomock.Controller) *MockAnonymousIdentityProvider {
	mock := &MockAnonymousIdentityProvider{ctrl: ctrl}
	mock.recorder = &MockAnonymousIdentityProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAnonymousIdentityProvider) EXPECT() *MockAnonymousIdentityProviderMockRecorder {
	return m.recorder
}

// List mocks base method
func (m *MockAnonymousIdentityProvider) List(userID string) ([]*anonymous.Identity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", userID)
	ret0, _ := ret[0].([]*anonymous.Identity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockAnonymousIdentityProviderMockRecorder) List(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockAnonymousIdentityProvider)(nil).List), userID)
}
