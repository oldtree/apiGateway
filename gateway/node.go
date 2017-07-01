package gateway

import (
	"encoding/json"
	"time"
)

const (
	NodeStatusOK    = 1
	NodeStatusError = -1
)

type Node struct {
	ServeiceName   string `json:"serveice_name,omitempty"`
	ConnectionType string `json:"connection_type,omitempty"`
	Address        string `json:"address,omitempty"`
	NodeId         int    `json:"node_id,omitempty"`
	NodeHost       string `json:"node_host,omitempty"`
	Weight         int    `json:"weight,omitempty"`

	Status       int8  `json:"status,omitempty"` // down,ok
	RecoverTimes uint8 `json:"recover_times,omitempty"`

	SuccessReq   uint64 `json:"success_req,omitempty"`
	FailedReq    uint64 `json:"failed_req,omitempty"`
	ReqPerSecond uint64 `json:"req_per_second,omitempty"`

	UpLineTime int64 `json:"up_line_time,omitempty"`
}

func NewNode() *Node {
	return &Node{
		UpLineTime: time.Now().Unix(),
	}
}

func (n *Node) FormatFromJson(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}
	return nil
}
