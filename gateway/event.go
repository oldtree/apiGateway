package gateway

import "encoding/json"

const (
	EventServiceGet = iota
	EventServiceAdd
	EventServiceUpdate
	EventServiceDelete

	EventServiceNodeGet
	EventServiceNodeAdd
	EventServiceNodeUpdate
	EventServiceNodeDelete
)

type Event struct {
	EventType int
	TimeStamp string
	Content   interface{}
}

type Result struct {
	Code        int         `json:"code,omitempty"`
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

func (r *Result) Json() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return data
}
