package mux

import (
	"net/http"
	"strings"
)

type Mux struct {
	mux                     *http.ServeMux
	handler                 http.Handler
	middlewares             []MiddlewareFunc
	notFoundHandler         HandlerFunc
	methodNotAllowedHandler HandlerFunc
	errorHandler            func(ctx *Context, err error)
	inline                  bool
}

func NewMux() *Mux {
	return &Mux{
		mux:                     http.NewServeMux(),
		handler:                 nil,
		middlewares:             make([]MiddlewareFunc, 0),
		notFoundHandler:         DefaultNotFoundHandler,
		methodNotAllowedHandler: DefaultMethodNotAllowedHandler,
		errorHandler:            DefaultErrorHandler,
		inline:                  false,
	}
}

type Router interface {
	http.Handler

	Handle(pattern string, fn HandlerFunc)
	WebSocket(pattern string, fn WebSocketFunc)
	Use(middlewares ...MiddlewareFunc)
	Mount(prefix string, router Router)
	Route(prefix string, fn func(router Router))
	Static(prefix, root string)
	File(path, file string)
}

func NewRouter() Router {
	return &Mux{
		mux:                     http.NewServeMux(),
		handler:                 nil,
		middlewares:             make([]MiddlewareFunc, 0),
		notFoundHandler:         nil,
		methodNotAllowedHandler: nil,
		errorHandler:            nil,
		inline:                  true,
	}
}

func (m *Mux) Static(prefix, root string) {
	if m.handler == nil {
		m.buildHandler()
	}

	trimmed := strings.TrimRight(prefix, "/")
	prefix = trimmed + "/"

	m.mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(root))))
}

func (m *Mux) File(path, file string) {
	if m.handler == nil {
		m.buildHandler()
	}

	m.mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	}))
}

func (m *Mux) Handle(pattern string, fn HandlerFunc) {
	if m.handler == nil {
		m.buildHandler()
	}
	m.mux.Handle(pattern, fn)
}

func (m *Mux) WebSocket(pattern string, fn WebSocketFunc) {
	if m.handler == nil {
		m.buildHandler()
	}
	m.mux.Handle(pattern, fn)
}

func (m *Mux) Use(middlewares ...MiddlewareFunc) {
	if m.handler != nil {
		return
	}
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *Mux) Mount(prefix string, router Router) {
	if m.handler == nil {
		m.buildHandler()
	}

	trimmed := strings.TrimRight(prefix, "/")
	prefix = trimmed + "/"

	m.mux.Handle(prefix, http.StripPrefix(trimmed, router))
}

func (m *Mux) Route(prefix string, fn func(router Router)) {
	if m.handler == nil {
		m.buildHandler()
	}

	router := NewRouter()
	fn(router)

	m.Mount(prefix, router)
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.handler == nil {
		m.buildHandler()
	}
	m.handler.ServeHTTP(w, r)
}

func (m *Mux) ErrorHandler(fn func(ctx *Context, err error)) {
	m.errorHandler = fn
}

func (m *Mux) NotFound(fn HandlerFunc) {
	m.notFoundHandler = fn
}

func (m *Mux) MethodNotAllowed(fn HandlerFunc) {
	m.methodNotAllowedHandler = fn
}

func (m *Mux) provideHandler() HandlerFunc {
	if m.inline {
		return func(ctx *Context) error {
			m.mux.ServeHTTP(ctx.responseWriter, ctx.request)
			return ctx.err
		}
	}

	return func(ctx *Context) error {
		m.mux.ServeHTTP(ctx.responseWriter, ctx.request)

		switch ctx.statusCode {
		case http.StatusNotFound:
			return m.notFoundHandler(ctx)
		case http.StatusMethodNotAllowed:
			return m.methodNotAllowedHandler(ctx)
		}

		return ctx.err
	}
}

func (m *Mux) provideContext(next HandlerFunc) http.HandlerFunc {
	if m.inline {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := extractContext(r)
			ctx.request = r
			ctx.err = next(ctx)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Context:        r.Context(),
			mux:            m,
			responseWriter: nil,
			request:        nil,
			statusCode:     http.StatusOK,
			err:            nil,
		}

		ctx.responseWriter = &hijackResponseWriter{
			ResponseWriter: w,
			ctx:            ctx,
			done:           false,
		}

		ctx.request = r.WithContext(ctx)
		ctx.err = next(ctx)
	}
}

func (m *Mux) provideErrorHandler(next HandlerFunc) HandlerFunc {
	if m.inline {
		return next
	}

	return func(ctx *Context) error {
		if err := next(ctx); err != nil {
			m.errorHandler(ctx, err)
		}
		return nil
	}
}

func (m *Mux) buildHandler() {
	handler := m.provideHandler()

	for i := len(m.middlewares) - 1; i >= 0; i-- {
		middleware := m.middlewares[i]
		handler = middleware(handler)
	}

	handler = m.provideErrorHandler(handler)
	m.handler = m.provideContext(handler)
}

func DefaultNotFoundHandler(ctx *Context) error {
	return NewHTTPError(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func DefaultMethodNotAllowedHandler(ctx *Context) error {
	return NewHTTPError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
}

func DefaultErrorHandler(ctx *Context, err error) {
	status := http.StatusInternalServerError
	message := http.StatusText(http.StatusInternalServerError)

	if e, ok := err.(HTTPError); ok {
		status = e.HTTPStatus()
		message = e.Error()
	}

	ctx.statusCode = status

	http.Error(ctx.responseWriter.ResponseWriter, message, status)
}
