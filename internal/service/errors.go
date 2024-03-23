package service

import "errors"

var (
	ErrUserExist = errors.New("user already exist")
	ErrCategoryExist = errors.New("category already exist")

	ErrInvalidCredentials = errors.New("invalid credential")
	ErrValidRefresh = errors.New("not valid refresh token")
)