package gateway

import (
	"encoding/json"
	"time"
)

type MethodInfo struct {
	InnerMethod string `json:"inner_method,omitempty"`
	OuterMethod string `json:"outer_method,omitempty"`
}

type AuthInfo struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

type RouteDesc struct {
	Description    string        `json:"description,omitempty"`
	Status         uint          `json:"status,omitempty"`
	NeedAuth       bool          `json:"need_auth,omitempty"`
	Auth           *AuthInfo     `json:"auth,omitempty"`
	InnerRoutePath string        `json:"inner_route_path,omitempty"`
	OuterRoutePath string        `json:"outer_route_path,omitempty"`
	HandleMethed   []*MethodInfo `json:"handle_methed,omitempty"`
}

func NewRouteDesc() *RouteDesc {
	return &RouteDesc{}
}

type ApiInfo struct {
	Info []*RouteDesc `json:"info,omitempty"`
}

func NewApiInfo() *ApiInfo {
	return &ApiInfo{}
}

type ServiceInfo struct {
	ServiceName       string   `json:"service_name,omitempty"`
	Version           string   `json:"version,omitempty"`
	Protocal          string   `json:"protocal,omitempty"`
	LoadBlanceType    uint     `json:"load_blance_type,omitempty"`
	ReadWriteTimeout  int      `json:"read_write_timeout,omitempty"`
	ConnectionTimeout int      `json:"connection_timeout,omitempty"`
	Api               *ApiInfo `json:"api,omitempty"`

	Createtime string
}

func NewServiceInfo() *ServiceInfo {
	return &ServiceInfo{
		Createtime: time.Now().String(),
	}
}

func (s *ServiceInfo) Encode() []byte {
	data, _ := json.Marshal(s)
	return data
}

func (s *ServiceInfo) Decode(data []byte) error {
	err := json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	return nil
}
