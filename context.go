package mux

import (
	"context"
	"net/http"
)

type Context struct {
	context.Context

	mux *Mux

	responseWriter *hijackResponseWriter
	request        *http.Request

	statusCode int
	err        error
}

func (ctx *Context) ResponseWriter() http.ResponseWriter {
	return ctx.responseWriter.ResponseWriter
}

func (ctx *Context) Request() *http.Request {
	return ctx.request
}

func (ctx *Context) SetStatus(statusCode int) {
	ctx.statusCode = statusCode
}

func (ctx *Context) Status() int {
	return ctx.statusCode
}

func (ctx *Context) Error(err error) {
	ctx.mux.errorHandler(ctx, err)
}

func extractContext(r *http.Request) *Context {
	ctx, ok := r.Context().(*Context)
	if !ok {
		panic("context not found")
	}
	return ctx
}
