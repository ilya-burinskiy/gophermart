// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ilya-burinskiy/gophermart/internal/storage (interfaces: Storage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/ilya-burinskiy/gophermart/internal/models"
	pgx "github.com/jackc/pgx/v5"
)

// MockStorage is a mock of Storage interface.
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage.
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance.
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// BeginTranscaction mocks base method.
func (m *MockStorage) BeginTranscaction(arg0 context.Context) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginTranscaction", arg0)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTranscaction indicates an expected call of BeginTranscaction.
func (mr *MockStorageMockRecorder) BeginTranscaction(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTranscaction", reflect.TypeOf((*MockStorage)(nil).BeginTranscaction), arg0)
}

// CreateBalance mocks base method.
func (m *MockStorage) CreateBalance(arg0 context.Context, arg1, arg2 int) (models.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateBalance", arg0, arg1, arg2)
	ret0, _ := ret[0].(models.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateBalance indicates an expected call of CreateBalance.
func (mr *MockStorageMockRecorder) CreateBalance(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateBalance", reflect.TypeOf((*MockStorage)(nil).CreateBalance), arg0, arg1, arg2)
}

// CreateOrder mocks base method.
func (m *MockStorage) CreateOrder(arg0 context.Context, arg1 int, arg2 string, arg3 models.OrderStatus) (models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrder", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrder indicates an expected call of CreateOrder.
func (mr *MockStorageMockRecorder) CreateOrder(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrder", reflect.TypeOf((*MockStorage)(nil).CreateOrder), arg0, arg1, arg2, arg3)
}

// CreateUser mocks base method.
func (m *MockStorage) CreateUser(arg0 context.Context, arg1, arg2 string) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockStorageMockRecorder) CreateUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockStorage)(nil).CreateUser), arg0, arg1, arg2)
}

// FindOrderByNumber mocks base method.
func (m *MockStorage) FindOrderByNumber(arg0 context.Context, arg1 string) (models.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOrderByNumber", arg0, arg1)
	ret0, _ := ret[0].(models.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindOrderByNumber indicates an expected call of FindOrderByNumber.
func (mr *MockStorageMockRecorder) FindOrderByNumber(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOrderByNumber", reflect.TypeOf((*MockStorage)(nil).FindOrderByNumber), arg0, arg1)
}

// FindUserByLogin mocks base method.
func (m *MockStorage) FindUserByLogin(arg0 context.Context, arg1 string) (models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindUserByLogin", arg0, arg1)
	ret0, _ := ret[0].(models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindUserByLogin indicates an expected call of FindUserByLogin.
func (mr *MockStorageMockRecorder) FindUserByLogin(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUserByLogin", reflect.TypeOf((*MockStorage)(nil).FindUserByLogin), arg0, arg1)
}

// UpdateBalanceCurrentAmount mocks base method.
func (m *MockStorage) UpdateBalanceCurrentAmount(arg0 context.Context, arg1, arg2 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateBalanceCurrentAmount", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateBalanceCurrentAmount indicates an expected call of UpdateBalanceCurrentAmount.
func (mr *MockStorageMockRecorder) UpdateBalanceCurrentAmount(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateBalanceCurrentAmount", reflect.TypeOf((*MockStorage)(nil).UpdateBalanceCurrentAmount), arg0, arg1, arg2)
}
