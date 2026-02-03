package services

import "errors"

var (
	ErrInvalidTitle = errors.New("invalid title")
	ErrNotFound     = errors.New("not found")
)
