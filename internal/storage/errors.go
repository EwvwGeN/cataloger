package storage

import "errors"

var (
	ErrUserExist = errors.New("user already exist")
	ErrCategoryExist = errors.New("category already exist")
	ErrProductExist = errors.New("product with this name already exist")
	ErrStartTx = errors.New("failed to begin transaction")
	ErrCommitTx = errors.New("error while commiting transaction")
	ErrRollbackTx = errors.New("failed to rollback transaction")
	ErrQuery = errors.New("error while executing query")
)