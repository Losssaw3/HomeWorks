package storage

import (
	"errors"
)

var (
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrAppNotFound   = errors.New("app not found")
	ErrActivateUser  = errors.New("failed to confirm email")
	ErrCreateArticle = errors.New("failed to create article")
	ErrEditArticle   = errors.New("failed to edit article")
	ErrAppExists     = errors.New("app already exists")
	ErrCacheNotFound = errors.New("cache not found")
)
