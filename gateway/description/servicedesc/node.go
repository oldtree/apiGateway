package servicedesc

import (
	"encoding/json"
)

type NodeDesc struct {
	SrvName  string `json:"srv_name,omitempty"`
	Id       int    `json:"id,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Address  string `json:"address,omitempty"`
}

func NewNodeDesc() *NodeDesc {
	return &NodeDesc{}
}

func (n *NodeDesc) FormatFromJson(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}
	return nil
}
