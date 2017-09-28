package servicedesc

import (
	"encoding/json"
	"time"
)

type MethodDesc struct {
	InnerMethod string `json:"inner_method,omitempty"`
	OuterMethod string `json:"outer_method,omitempty"`
}

type AuthDesc struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	AuthType int    `json:"auth_type,omitempty"`
}

type RouteDesc struct {
	Description    string        `json:"description,omitempty"`
	Status         uint          `json:"status,omitempty"`
	NeedAuth       bool          `json:"need_auth,omitempty"`
	Auth           *AuthDesc     `json:"auth,omitempty"`
	InnerRoutePath string        `json:"inner_route_path,omitempty"`
	OuterRoutePath string        `json:"outer_route_path,omitempty"`
	HandleMethed   []*MethodDesc `json:"handle_methed,omitempty"`
}

func NewRouteDesc() *RouteDesc {
	return &RouteDesc{}
}

type ApiDesc struct {
	Info []*RouteDesc `json:"info,omitempty"`
}

func NewApiDesc() *ApiDesc {
	return &ApiDesc{
		Info: *new([]*RouteDesc),
	}
}

type ServiceDesc struct {
	ServiceName string `json:"service_name,omitempty"`
	Version     string `json:"version,omitempty"`

	Protocal string `json:"protocal,omitempty"` // deafault is http/https ,support tcp or udp

	LoadBalanceType   uint `json:"load_balance_type,omitempty"`
	ReadWriteTimeout  int  `json:"read_write_timeout,omitempty"`
	ConnectionTimeout int  `json:"connection_timeout,omitempty"`

	Api *ApiDesc `json:"api,omitempty"`

	XEtag map[string]string `json:"x_etag,omitempty"`

	Createtime string `json:"createtime,omitempty"`
}

func NewServiceDesc() *ServiceDesc {
	return &ServiceDesc{
		Createtime: time.Now().String(),
	}
}

func (s *ServiceDesc) Encode() []byte {
	data, _ := json.Marshal(s)
	return data
}

func (s *ServiceDesc) Decode(data []byte) error {
	err := json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	return nil
}
