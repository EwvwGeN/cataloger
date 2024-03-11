package jwt

import "errors"

var (
	ErrEmptyValue = errors.New("empty value")
	ErrRefreshGenerate = errors.New("failed to generate refresh token")
	ErrParseClaims = errors.New("failed to get claims from token")
)