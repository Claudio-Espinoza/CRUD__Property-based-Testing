package domain

import "errors"

var (
	ErrInvalidUserName  = errors.New("user name must be between 2 and 50 characters and contain only letters and spaces")
	ErrInvalidUserEmail = errors.New("user email must be a valid email address")
	ErrInvalidUserAge   = errors.New("user age must be between 0 and 150")
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)
