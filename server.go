package openws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coder/websocket"
)

type Server struct {
	handlers map[string]rawHandler
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]rawHandler),
	}
}

type rawHandler func(ctx context.Context, id json.RawMessage, rawParams json.RawMessage) *Response

func Register[Params any, Result any](s *Server, method string, fn func(context.Context, *Params) (Result, error)) {
	if _, exists := s.handlers[method]; exists {
		panic("duplicate registration of method " + method)
	}

	s.handlers[method] = func(ctx context.Context, id json.RawMessage, rawParams json.RawMessage) *Response {
		if len(rawParams) == 0 || string(rawParams) == "null" {
			rawParams = []byte("{}")
		}

		var params Params
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return newErrorResponse(id, &Error{
				Code:    InvalidParamsCode,
				Message: err.Error(),
				Details: nil,
			})
		}

		res, err := fn(ctx, &params)
		if err != nil {
			var rpcErr *Error
			if errors.As(err, &rpcErr) {
				return newErrorResponse(id, rpcErr)
			}

			return newErrorResponse(id, &Error{
				Code:    InternalErrorCode,
				Message: err.Error(),
				Details: nil,
			})
		}

		return newResponse(id, res)
	}
}

func (s Server) ServeConn(ctx context.Context, conn *websocket.Conn) error {
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}

		response := s.handleRequest(ctx, msg)

		if err := conn.Write(ctx, websocket.MessageText, response); err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
	}
}

func (s Server) handleRequest(ctx context.Context, raw []byte) []byte {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return MustMarshal(newErrorResponse(req.ID, &Error{
			Code:    ParseErrorCode,
			Message: err.Error(),
			Details: nil,
		}))
	}

	if req.Method == "" || len(req.ID) == 0 {
		return MustMarshal(newErrorResponse(req.ID, &Error{
			Code:    InvalidRequestCode,
			Message: "Invalid request object",
			Details: nil,
		}))
	}

	h, ok := s.handlers[req.Method]
	if !ok {
		return MustMarshal(newErrorResponse(req.ID, &Error{
			Code:    MethodNotFoundCode,
			Message: fmt.Sprintf("method %s does not exists", req.Method),
			Details: nil,
		}))
	}

	resp := h(ctx, req.ID, req.Params)

	return MustMarshal(resp)
}

func EmitEvent(ctx context.Context, eventName string, data any, conns ...*websocket.Conn) {
	resp := Event{
		Event: eventName,
		Data:  data,
	}

	raw := MustMarshal(resp)
	for _, c := range conns {
		_ = c.Write(ctx, websocket.MessageText, raw)
	}
}
