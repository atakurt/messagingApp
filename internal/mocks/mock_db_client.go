// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/atakurt/messagingApp/internal/infrastructure/db (interfaces: DBInterface)

// Package mocks is a generated GoMock package.
package mocks

import (
	sql "database/sql"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	gorm "gorm.io/gorm"
)

// MockDBInterface is a mock of DBInterface interface.
type MockDBInterface struct {
	ctrl     *gomock.Controller
	recorder *MockDBInterfaceMockRecorder
}

// MockDBInterfaceMockRecorder is the mock recorder for MockDBInterface.
type MockDBInterfaceMockRecorder struct {
	mock *MockDBInterface
}

// NewMockDBInterface creates a new mock instance.
func NewMockDBInterface(ctrl *gomock.Controller) *MockDBInterface {
	mock := &MockDBInterface{ctrl: ctrl}
	mock.recorder = &MockDBInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBInterface) EXPECT() *MockDBInterfaceMockRecorder {
	return m.recorder
}

// Begin mocks base method.
func (m *MockDBInterface) Begin() *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin")
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// Begin indicates an expected call of Begin.
func (mr *MockDBInterfaceMockRecorder) Begin() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockDBInterface)(nil).Begin))
}

// GetDB mocks base method.
func (m *MockDBInterface) GetDB() *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDB")
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// GetDB indicates an expected call of GetDB.
func (mr *MockDBInterfaceMockRecorder) GetDB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDB", reflect.TypeOf((*MockDBInterface)(nil).GetDB))
}

// GetSQLDB mocks base method.
func (m *MockDBInterface) GetSQLDB() (*sql.DB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSQLDB")
	ret0, _ := ret[0].(*sql.DB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSQLDB indicates an expected call of GetSQLDB.
func (mr *MockDBInterfaceMockRecorder) GetSQLDB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSQLDB", reflect.TypeOf((*MockDBInterface)(nil).GetSQLDB))
}
