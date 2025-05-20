package openws

import (
	"context"
	"encoding/json"
)

type Server struct {
	handlers map[string]handler
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]handler),
	}
}

type handler func(ctx context.Context, id json.RawMessage, rawParams json.RawMessage) *RPCResponse

func Register[Params any, Result any](s *Server, method string, fn func(context.Context, *Params) (Result, *RPCError)) {
	s.handlers[method] = func(ctx context.Context, id json.RawMessage, rawParams json.RawMessage) *RPCResponse {
		if len(rawParams) == 0 || string(rawParams) == "null" {
			rawParams = []byte("{}")
		}

		var p Params
		if err := json.Unmarshal(rawParams, &p); err != nil {
			return makeErrorResponse(id, ErrInvalidParams)
		}

		res, err := fn(ctx, &p)
		if err != nil {
			return makeErrorResponse(id, err)
		}

		return &RPCResponse{
			ID:     id,
			Result: res,
			Error:  nil,
		}
	}
}

func (s *Server) HandleRequest(ctx context.Context, raw []byte) []byte {
	var req RPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return mustMarshal(makeErrorResponse(nil, ErrParse))
	}

	if req.Method == "" || len(req.ID) == 0 {
		return mustMarshal(makeErrorResponse(req.ID, ErrInvalidReq))
	}

	handler, ok := s.handlers[req.Method]
	if !ok {
		return mustMarshal(makeErrorResponse(req.ID, ErrMethodNotFound))
	}

	resp := handler(ctx, req.ID, req.Params)

	return mustMarshal(resp)
}

func EmitEvent(eventName string, data any) []byte {
	resp := RPCEvent{
		Event: eventName,
		Data:  data,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	return b
}
