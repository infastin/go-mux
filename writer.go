package mux

import "net/http"

type hijackResponseWriter struct {
	http.ResponseWriter

	ctx  *Context
	done bool
}

func (w *hijackResponseWriter) WriteHeader(statusCode int) {
	if w.done {
		return
	}

	w.ctx.statusCode = statusCode
	if statusCode == http.StatusNotFound || statusCode == http.StatusMethodNotAllowed {
		w.done = true
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *hijackResponseWriter) Write(b []byte) (int, error) {
	if w.done {
		return 0, nil
	}
	return w.ResponseWriter.Write(b)
}
