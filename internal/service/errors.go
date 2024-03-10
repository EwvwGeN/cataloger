package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credential")
	ErrValidRefresh = errors.New("not valid refresh token")
)