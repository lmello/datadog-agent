// Code generated by mockery v2.20.2. DO NOT EDIT.

package mocks

import (
	eval "github.com/DataDog/datadog-agent/pkg/compliance/eval"
	mock "github.com/stretchr/testify/mock"
)

// Iterator is an autogenerated mock type for the Iterator type
type Iterator struct {
	mock.Mock
}

// Done provides a mock function with given fields:
func (_m *Iterator) Done() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Next provides a mock function with given fields:
func (_m *Iterator) Next() (eval.Instance, error) {
	ret := _m.Called()

	var r0 eval.Instance
	var r1 error
	if rf, ok := ret.Get(0).(func() (eval.Instance, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() eval.Instance); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(eval.Instance)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIterator interface {
	mock.TestingT
	Cleanup(func())
}

// NewIterator creates a new instance of Iterator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIterator(t mockConstructorTestingTNewIterator) *Iterator {
	mock := &Iterator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}