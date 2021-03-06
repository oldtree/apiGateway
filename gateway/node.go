package gateway

import (
	"encoding/json"

	"github.com/oldtree/apiGateway/gateway/description/servicedesc"
)

const (
	//NodeStatusError : node is not aviliable
	NodeStatusError = 1 << iota
	//NodeStatusOK : node is ok
	NodeStatusOK
	//node is temp not reach able
	NodeStatusUnReachable
	//node access number per minitues too hot
	NodeStatusOverHit
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

func NewNodeFromData(data []byte) *Node {
	nodedesc := servicedesc.NewNodeDesc()
	err := json.Unmarshal(data, nodedesc)
	if err != nil {
		return nil
	}
	node := NewDefaultNode(nodedesc.SrvName, nodedesc.Address, nodedesc.Id, nodedesc.Hostname)
	return node
}

func NewDefaultNode(srv string, address string, id int, hostname string) *Node {
	return &Node{
		ServeiceName:   srv,
		ConnectionType: "http",
		Address:        address,
		NodeID:         id,
		NodeHost:       hostname,
		Weight:         5,
		Status:         1,
		RecoverTimes:   30,
	}
}

//FormatFromJSON format node info from []byte data
func (n *Node) FormatFromJSON(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}
	return nil
}
