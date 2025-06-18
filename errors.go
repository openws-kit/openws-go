package openws

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func newErrorResponse(id json.RawMessage, err *Error) *Response {
	return &Response{
		JSONRPC: JSONRPCVer,
		ID:      id,
		Result:  nil,
		Error:   err,
	}
}

func (e *Error) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("rpc error %d: %s (%v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

const (
	ParseErrorCode         = -32700
	InvalidRequestCode     = -32600
	MethodNotFoundCode     = -32601
	InvalidParamsCode      = -32602
	InternalErrorCode      = -32603
	UnimplementedErrorCode = -32604
)
