package gateway

import "errors"

var (
	ErrServiceNotFound     = errors.New("service not found")
	ErrServiceNotAviliable = errors.New("service not aviliable")
)
