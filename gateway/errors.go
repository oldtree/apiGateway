package gateway

import "errors"

//error condig would put error info to etcd cluster , backend node will update error description per 10 min
//the etcd url will be http://hostname/2379/v2/keys/servicename/errordesc
//every backend will use a RWLOCK to modify their error map when they want concurrecy update error info
type ErrDesc struct {
	ErrorDomain      string //auth , server ,db operation ,cache opeartion ?
	ErrorCode        uint   //100011,200022
	ErrorDescription string //auth failed or db connection failed
}

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
