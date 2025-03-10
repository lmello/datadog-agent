// Code generated by mockery v2.33.1. DO NOT EDIT.

package mocks

import (
	context "context"

	api "github.com/DataDog/datadog-agent/pkg/eventmonitor/proto/api"

	metadata "google.golang.org/grpc/metadata"

	mock "github.com/stretchr/testify/mock"
)

// EventMonitoringModule_GetProcessEventsClient is an autogenerated mock type for the EventMonitoringModule_GetProcessEventsClient type
type EventMonitoringModule_GetProcessEventsClient struct {
	mock.Mock
}

// CloseSend provides a mock function with given fields:
func (_m *EventMonitoringModule_GetProcessEventsClient) CloseSend() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Context provides a mock function with given fields:
func (_m *EventMonitoringModule_GetProcessEventsClient) Context() context.Context {
	ret := _m.Called()

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// Header provides a mock function with given fields:
func (_m *EventMonitoringModule_GetProcessEventsClient) Header() (metadata.MD, error) {
	ret := _m.Called()

	var r0 metadata.MD
	var r1 error
	if rf, ok := ret.Get(0).(func() (metadata.MD, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Recv provides a mock function with given fields:
func (_m *EventMonitoringModule_GetProcessEventsClient) Recv() (*api.ProcessEventMessage, error) {
	ret := _m.Called()

	var r0 *api.ProcessEventMessage
	var r1 error
	if rf, ok := ret.Get(0).(func() (*api.ProcessEventMessage, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *api.ProcessEventMessage); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*api.ProcessEventMessage)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecvMsg provides a mock function with given fields: m
func (_m *EventMonitoringModule_GetProcessEventsClient) RecvMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendMsg provides a mock function with given fields: m
func (_m *EventMonitoringModule_GetProcessEventsClient) SendMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Trailer provides a mock function with given fields:
func (_m *EventMonitoringModule_GetProcessEventsClient) Trailer() metadata.MD {
	ret := _m.Called()

	var r0 metadata.MD
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	return r0
}

// NewEventMonitoringModule_GetProcessEventsClient creates a new instance of EventMonitoringModule_GetProcessEventsClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEventMonitoringModule_GetProcessEventsClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *EventMonitoringModule_GetProcessEventsClient {
	mock := &EventMonitoringModule_GetProcessEventsClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
