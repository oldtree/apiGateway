package gateway

import (
	"encoding/json"
)

type Config struct {
	Version string
	Port    string
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
