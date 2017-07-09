package gateway

import "errors"

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrServiceNotAviliable  = errors.New("service not aviliable")
	ErrServiceInternalError = errors.New("service internal error")
)

var (
	ErrEtcdConnectionFailed = errors.New("etcd connect failed")
)

var (
	ErrParseServiceInfoFailed   = errors.New("parse service info failed")
	ErrMappingServiceInfoFailed = errors.New("mapping service info failed")
)
