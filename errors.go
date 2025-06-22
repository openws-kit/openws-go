package openws

import (
	"encoding/json"
	"fmt"
)

type RPCErrorer interface {
	RPCError() *Error
}

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *Error) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("rpc error %d: %s (%v)", e.Code, e.Message, e.Details)
	}

	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

func (e *Error) RPCError() *Error {
	return e
}

func newErrorResponse(id json.RawMessage, err *Error) *Response {
	return &Response{
		JSONRPC: JSONRPCVer,
		ID:      id,
		Result:  nil,
		Error:   err,
	}
}

const (
	CodeParseError         = -32700
	CodeInvalidRequest     = -32600
	CodeMethodNotFound     = -32601
	CodeInvalidParams      = -32602
	CodeInternalError      = -32603
	CodeUnimplementedError = -32604
)
