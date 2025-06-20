package openws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Server struct {
	handlers map[string]rawHandler
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]rawHandler),
	}
}

type rawHandler func(ctx context.Context, id json.RawMessage, rawParams json.RawMessage) (any, *Error)

func Register[Params any, Result any](s *Server, method string, fn func(context.Context, *Params) (Result, error)) {
	if _, exists := s.handlers[method]; exists {
		panic("duplicate registration of method " + method)
	}

	s.handlers[method] = func(ctx context.Context, id json.RawMessage, rawParams json.RawMessage) (any, *Error) {
		if len(rawParams) == 0 || string(rawParams) == "null" {
			rawParams = []byte("{}")
		}

		var params Params
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, &Error{
				Code:    InvalidParamsCode,
				Message: err.Error(),
				Details: nil,
			}
		}

		res, err := fn(ctx, &params)
		if err != nil {
			var rpc RPCErrorer
			if errors.As(err, &rpc) {
				return nil, rpc.RPCError()
			}

			return nil, &Error{
				Code:    InternalErrorCode,
				Message: err.Error(),
				Details: nil,
			}
		}

		return res, nil
	}
}

func (s Server) ServeConn(ctx context.Context, conn *websocket.Conn) error {
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}

		response := s.handleRequest(ctx, msg)

		if err := wsjson.Write(ctx, conn, response); err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
	}
}

func (s Server) handleRequest(ctx context.Context, raw []byte) *Response {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return newErrorResponse(req.ID, &Error{
			Code:    ParseErrorCode,
			Message: err.Error(),
			Details: nil,
		})
	}

	if req.Method == "" || len(req.ID) == 0 {
		return newErrorResponse(req.ID, &Error{
			Code:    InvalidRequestCode,
			Message: "Invalid request object",
			Details: nil,
		})
	}

	h, ok := s.handlers[req.Method]
	if !ok {
		return newErrorResponse(req.ID, &Error{
			Code:    MethodNotFoundCode,
			Message: fmt.Sprintf("method %s does not exists", req.Method),
			Details: nil,
		})
	}

	res, err := h(ctx, req.ID, req.Params)
	if err != nil {
		return newErrorResponse(req.ID, err)
	}

	return &Response{
		JSONRPC: JSONRPCVer,
		ID:      req.ID,
		Result:  res,
		Error:   nil,
	}
}
