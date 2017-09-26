package nodedesc

import (
	"encoding/json"
)

type NodeInfo struct {
	Id      int64  `json:"id,omitempty"`
	Address string `json:"address,omitempty"`
}

func NewNodeInfo() *NodeInfo {
	return &NodeInfo{}
}

func (n *NodeInfo) FormatFromJson(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}
	return nil
}
