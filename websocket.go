package mux

import (
	"net/http"

	"nhooyr.io/websocket"
)

type WebSocketContext struct {
	*Context
	websocketConn *websocket.Conn
}

func (ctx *WebSocketContext) WebSocketConn() *websocket.Conn {
	return ctx.websocketConn
}

type WebSocketFunc func(ctx *WebSocketContext)

func (fn WebSocketFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := extractContext(r)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		ctx.err = err
		return
	}

	wctx := &WebSocketContext{
		Context:       ctx,
		websocketConn: conn,
	}

	defer func() {
		_ = conn.CloseNow()
		wctx.websocketConn = nil
	}()

	fn(wctx)
}
