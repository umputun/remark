// Code generated by mockery v1.0.0. DO NOT EDIT.

package engine

import mock "github.com/stretchr/testify/mock"
import store "github.com/umputun/remark/backend/app/store"

// MockInterface is an autogenerated mock type for the Interface type
type MockInterface struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockInterface) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Count provides a mock function with given fields: req
func (_m *MockInterface) Count(req FindRequest) (int, error) {
	ret := _m.Called(req)

	var r0 int
	if rf, ok := ret.Get(0).(func(FindRequest) int); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(FindRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Create provides a mock function with given fields: comment
func (_m *MockInterface) Create(comment store.Comment) (string, error) {
	ret := _m.Called(comment)

	var r0 string
	if rf, ok := ret.Get(0).(func(store.Comment) string); ok {
		r0 = rf(comment)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(store.Comment) error); ok {
		r1 = rf(comment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: req
func (_m *MockInterface) Delete(req DeleteRequest) error {
	ret := _m.Called(req)

	var r0 error
	if rf, ok := ret.Get(0).(func(DeleteRequest) error); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Find provides a mock function with given fields: req
func (_m *MockInterface) Find(req FindRequest) ([]store.Comment, error) {
	ret := _m.Called(req)

	var r0 []store.Comment
	if rf, ok := ret.Get(0).(func(FindRequest) []store.Comment); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.Comment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(FindRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Flag provides a mock function with given fields: req
func (_m *MockInterface) Flag(req FlagRequest) (bool, error) {
	ret := _m.Called(req)

	var r0 bool
	if rf, ok := ret.Get(0).(func(FlagRequest) bool); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(FlagRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: req
func (_m *MockInterface) Get(req GetRequest) (store.Comment, error) {
	ret := _m.Called(req)

	var r0 store.Comment
	if rf, ok := ret.Get(0).(func(GetRequest) store.Comment); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Get(0).(store.Comment)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(GetRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Info provides a mock function with given fields: req
func (_m *MockInterface) Info(req InfoRequest) ([]store.PostInfo, error) {
	ret := _m.Called(req)

	var r0 []store.PostInfo
	if rf, ok := ret.Get(0).(func(InfoRequest) []store.PostInfo); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]store.PostInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(InfoRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDetails provides a mock function with given fields: loc
func (_m *MockInterface) ListDetails(loc store.Locator) (map[string]UserDetailEntry, error) {
	ret := _m.Called(loc)

	var r0 map[string]UserDetailEntry
	if rf, ok := ret.Get(0).(func(store.Locator) map[string]UserDetailEntry); ok {
		r0 = rf(loc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]UserDetailEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(store.Locator) error); ok {
		r1 = rf(loc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListFlags provides a mock function with given fields: req
func (_m *MockInterface) ListFlags(req FlagRequest) ([]interface{}, error) {
	ret := _m.Called(req)

	var r0 []interface{}
	if rf, ok := ret.Get(0).(func(FlagRequest) []interface{}); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(FlagRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: comment
func (_m *MockInterface) Update(comment store.Comment) error {
	ret := _m.Called(comment)

	var r0 error
	if rf, ok := ret.Get(0).(func(store.Comment) error); ok {
		r0 = rf(comment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UserDetail provides a mock function with given fields: req
func (_m *MockInterface) UserDetail(req UserDetailRequest) (string, error) {
	ret := _m.Called(req)

	var r0 string
	if rf, ok := ret.Get(0).(func(UserDetailRequest) string); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(UserDetailRequest) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
