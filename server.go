package openws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

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
				Code:    CodeInvalidParams,
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
				Code:    CodeInternalError,
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
			Code:    CodeParseError,
			Message: err.Error(),
			Details: nil,
		})
	}

	if req.Method == "" || len(req.ID) == 0 {
		return newErrorResponse(req.ID, &Error{
			Code:    CodeInvalidRequest,
			Message: "Invalid request object",
			Details: nil,
		})
	}

	h, ok := s.handlers[req.Method]
	if !ok {
		return newErrorResponse(req.ID, &Error{
			Code:    CodeMethodNotFound,
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

type SendResult struct {
	Conn *websocket.Conn
	Err  error
}

func BroadcastEvent(
	ctx context.Context,
	eventName string,
	data any,
	conns ...*websocket.Conn,
) (<-chan SendResult, error) {
	event := &Event{
		Event: eventName,
		Data:  data,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	resultCh := make(chan SendResult)

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(conns))

		for _, conn := range conns {
			go func(c *websocket.Conn) {
				defer wg.Done()
				err := c.Write(ctx, websocket.MessageText, eventBytes)
				resultCh <- SendResult{Conn: c, Err: err}
			}(conn)
		}

		wg.Wait()
		close(resultCh)
	}()

	return resultCh, nil
}

func SendEvent(
	ctx context.Context,
	eventName string,
	data any,
	conn *websocket.Conn,
) error {
	event := &Event{
		Event: eventName,
		Data:  data,
	}

	if err := wsjson.Write(ctx, conn, event); err != nil {
		return fmt.Errorf("failed to write to connection: %w", err)
	}

	return nil
}
