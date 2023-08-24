/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by MockGen. DO NOT EDIT.
// Source: handlers.go

// Package oidc is a generated GoMock package.
package oidc

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	go_jose_v2 "gopkg.in/square/go-jose.v2"
)

// MockTokenHandler is a mock of TokenHandler interface.
type MockTokenHandler struct {
	ctrl     *gomock.Controller
	recorder *MockTokenHandlerMockRecorder
}

// MockTokenHandlerMockRecorder is the mock recorder for MockTokenHandler.
type MockTokenHandlerMockRecorder struct {
	mock *MockTokenHandler
}

// NewMockTokenHandler creates a new mock instance.
func NewMockTokenHandler(ctrl *gomock.Controller) *MockTokenHandler {
	mock := &MockTokenHandler{ctrl: ctrl}
	mock.recorder = &MockTokenHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTokenHandler) EXPECT() *MockTokenHandlerMockRecorder {
	return m.recorder
}

// Token mocks base method.
func (m *MockTokenHandler) Token() (Token, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Token")
	ret0, _ := ret[0].(Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Token indicates an expected call of Token.
func (mr *MockTokenHandlerMockRecorder) Token() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Token", reflect.TypeOf((*MockTokenHandler)(nil).Token))
}

// MockJWKsHandler is a mock of JWKsHandler interface.
type MockJWKsHandler struct {
	ctrl     *gomock.Controller
	recorder *MockJWKsHandlerMockRecorder
}

// MockJWKsHandlerMockRecorder is the mock recorder for MockJWKsHandler.
type MockJWKsHandlerMockRecorder struct {
	mock *MockJWKsHandler
}

// NewMockJWKsHandler creates a new mock instance.
func NewMockJWKsHandler(ctrl *gomock.Controller) *MockJWKsHandler {
	mock := &MockJWKsHandler{ctrl: ctrl}
	mock.recorder = &MockJWKsHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJWKsHandler) EXPECT() *MockJWKsHandlerMockRecorder {
	return m.recorder
}

// KeySet mocks base method.
func (m *MockJWKsHandler) KeySet() go_jose_v2.JSONWebKeySet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KeySet")
	ret0, _ := ret[0].(go_jose_v2.JSONWebKeySet)
	return ret0
}

// KeySet indicates an expected call of KeySet.
func (mr *MockJWKsHandlerMockRecorder) KeySet() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeySet", reflect.TypeOf((*MockJWKsHandler)(nil).KeySet))
}
