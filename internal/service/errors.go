package service

import "errors"

var (
	ErrUserExist = errors.New("user already exist")
	ErrCategoryExist = errors.New("category with code already exist")
	ErrCategoriesCodes = errors.New("categories with some codes not exists")
	ErrCategoryInUse = errors.New("category with this code in use")
	ErrProductExist = errors.New("product with this name already exist")
	ErrInvalidCredentials = errors.New("invalid credential")
	ErrValidRefresh = errors.New("not valid refresh token")
)