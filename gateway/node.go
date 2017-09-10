package gateway

import (
	"encoding/json"
)

const (
	//NodeStatusOK : node is ok
	NodeStatusOK = 1
	//NodeStatusError : node is not aviliable
	NodeStatusError = -1
)

//Node 只包含节点的运行状态信息，连接属性信息，以及一些描述信息
//不包含路由转发的信息
type Node struct {
	ServeiceName   string `json:"serveice_name,omitempty"`
	ConnectionType string `json:"connection_type,omitempty"`
	Address        string `json:"address,omitempty"`
	NodeID         int    `json:"node_id,omitempty"`
	NodeHost       string `json:"node_host,omitempty"`
	Weight         int    `json:"weight,omitempty"`

	Status       int8  `json:"status,omitempty"`        // down,ok
	RecoverTimes uint8 `json:"recover_times,omitempty"` // seconds

	SuccessReq   uint64 `json:"success_req,omitempty"`    // uint64 number
	FailedReq    uint64 `json:"failed_req,omitempty"`     // uint64 number
	ReqPerSecond uint64 `json:"req_per_second,omitempty"` // uint64 number

}

//NewNode : Create a new node
func NewNode() *Node {
	return &Node{}
}

//FormatFromJSON format node info from []byte data
func (n *Node) FormatFromJSON(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}
	return nil
}
