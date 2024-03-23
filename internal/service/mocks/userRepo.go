// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	models "github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	mock "github.com/stretchr/testify/mock"
)

// UserRepo is an autogenerated mock type for the userRepo type
type UserRepo struct {
	mock.Mock
}

// GetUserByEmail provides a mock function with given fields: ctx, email
func (_m *UserRepo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	ret := _m.Called(ctx, email)

	if len(ret) == 0 {
		panic("no return value specified for GetUserByEmail")
	}

	var r0 models.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (models.User, error)); ok {
		return rf(ctx, email)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) models.User); ok {
		r0 = rf(ctx, email)
	} else {
		r0 = ret.Get(0).(models.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveRefreshToken provides a mock function with given fields: ctx, email, refreshToken, rttl
func (_m *UserRepo) SaveRefreshToken(ctx context.Context, email string, refreshToken string, rttl int64) error {
	ret := _m.Called(ctx, email, refreshToken, rttl)

	if len(ret) == 0 {
		panic("no return value specified for SaveRefreshToken")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64) error); ok {
		r0 = rf(ctx, email, refreshToken, rttl)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveUser provides a mock function with given fields: ctx, email, passHash
func (_m *UserRepo) SaveUser(ctx context.Context, email string, passHash string) error {
	ret := _m.Called(ctx, email, passHash)

	if len(ret) == 0 {
		panic("no return value specified for SaveUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, email, passHash)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewUserRepo creates a new instance of UserRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserRepo {
	mock := &UserRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
