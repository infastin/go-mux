package mux

import (
	"net/http"
)

type HandlerFunc func(ctx *Context) error

func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := extractContext(r)
	ctx.err = fn(ctx)
}

type MiddlewareFunc func(next HandlerFunc) HandlerFunc
