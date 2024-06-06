package mux

import (
	"net/http"
	"strings"
)

type Router struct {
	mux         *http.ServeMux
	handler     http.Handler
	middlewares []MiddlewareFunc
}

func NewRouter() *Router {
	return &Router{
		mux:         http.NewServeMux(),
		handler:     nil,
		middlewares: make([]MiddlewareFunc, 0),
	}
}

func (m *Router) Handle(pattern string, fn HandlerFunc) {
	if m.handler == nil {
		m.buildHandler()
	}
	m.mux.Handle(pattern, fn)
}

func (m *Router) WebSocket(pattern string, fn WebSocketFunc) {
	if m.handler == nil {
		m.buildHandler()
	}
	m.mux.Handle(pattern, fn)
}

func (m *Router) Use(middlewares ...MiddlewareFunc) {
	if m.handler != nil {
		return
	}
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *Router) Mount(prefix string, router *Router) {
	if m.handler == nil {
		m.buildHandler()
	}

	trimmed := strings.TrimRight(prefix, "/")
	prefix = trimmed + "/"

	m.mux.Handle(prefix, http.StripPrefix(trimmed, router))
}

func (m *Router) Route(prefix string, fn func(router *Router)) {
	if m.handler == nil {
		m.buildHandler()
	}

	router := NewRouter()
	fn(router)

	m.Mount(prefix, router)
}

func (m *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.handler == nil {
		m.buildHandler()
	}
	m.handler.ServeHTTP(w, r)
}

func (m *Router) routeHTTP(ctx *Context) error {
	m.mux.ServeHTTP(ctx.responseWriter, ctx.request)
	return ctx.err
}

func (*Router) provideContext(next HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := extractContext(r)
		ctx.request = r
		ctx.err = next(ctx)
	}
}

func (m *Router) buildHandler() {
	handler := m.routeHTTP

	for i := len(m.middlewares) - 1; i >= 0; i-- {
		middleware := m.middlewares[i]
		handler = middleware(handler)
	}

	m.handler = m.provideContext(handler)
}
