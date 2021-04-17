// Code generated by MockGen. DO NOT EDIT.
// Source: manager.go

// Package usage is a generated GoMock package.
package usage

import (
	strfmt "github.com/go-openapi/strfmt"
	gomock "github.com/golang/mock/gomock"
	gorm "github.com/jinzhu/gorm"
	reflect "reflect"
)

// MockAPI is a mock of API interface
type MockAPI struct {
	ctrl     *gomock.Controller
	recorder *MockAPIMockRecorder
}

// MockAPIMockRecorder is the mock recorder for MockAPI
type MockAPIMockRecorder struct {
	mock *MockAPI
}

// NewMockAPI creates a new mock instance
func NewMockAPI(ctrl *gomock.Controller) *MockAPI {
	mock := &MockAPI{ctrl: ctrl}
	mock.recorder = &MockAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAPI) EXPECT() *MockAPIMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockAPI) Add(usages FeatureUsage, name string, data *map[string]interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Add", usages, name, data)
}

// Add indicates an expected call of Add
func (mr *MockAPIMockRecorder) Add(usages, name, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockAPI)(nil).Add), usages, name, data)
}

// Remove mocks base method
func (m *MockAPI) Remove(usages FeatureUsage, name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Remove", usages, name)
}

// Remove indicates an expected call of Remove
func (mr *MockAPIMockRecorder) Remove(usages, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockAPI)(nil).Remove), usages, name)
}

// Save mocks base method
func (m *MockAPI) Save(db *gorm.DB, clusterId strfmt.UUID, usages FeatureUsage) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Save", db, clusterId, usages)
}

// Save indicates an expected call of Save
func (mr *MockAPIMockRecorder) Save(db, clusterId, usages interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockAPI)(nil).Save), db, clusterId, usages)
}
