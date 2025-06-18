package openws

import (
	"encoding/json"
)

const JSONRPCVer = "2.0"

type Event struct {
	Event string `json:"event"`
	Data  any    `json:"data,omitempty"`
}

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func newResponse(id json.RawMessage, result any) *Response {
	return &Response{
		JSONRPC: JSONRPCVer,
		ID:      id,
		Result:  result,
		Error:   nil,
	}
}
