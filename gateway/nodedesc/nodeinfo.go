package nodedesc

import (
	"encoding/json"
	"time"
)

type NodeInfo struct {
	Id            int64  `json:"id,omitempty"`
	Hostname      string `json:"hostname,omitempty"`
	Address       string `json:"address,omitempty"`
	Gorutine      uint   `json:"gorutine,omitempty"`
	CpuNum        uint   `json:"cpu_num,omitempty"`
	Meminfo       uint   `json:"meminfo,omitempty"`
	ConnectionNum uint   `json:"connection_num,omitempty"`
	ReportTime    int64  `json:"report_time,omitempty"`
}

func NewNodeInfo() *NodeInfo {
	return &NodeInfo{
		ReportTime: time.Now().Unix(),
	}
}

func (n *NodeInfo) FormatFromJson(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}
	return nil
}
