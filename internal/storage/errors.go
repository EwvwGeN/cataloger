package storage

import "errors"

var (
	ErrUserExist = errors.New("user already exist")
	ErrCategoryExist = errors.New("category already exist")
	ErrProductExist = errors.New("product already exist")
	ErrQuery = errors.New("error while executing query")
)