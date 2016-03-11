package notifiers

import "github.com/stretchr/testify/mock"

import "github.com/ashwanthkumar/marathon-alerts/checks"

type MockNotifier struct {
	mock.Mock
}

// Notify provides a mock function with given fields: check
func (_m *MockNotifier) Notify(check checks.AppCheck) {
	_m.Called(check)
}

// Name provides a mock function with given fields:
func (_m *MockNotifier) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
