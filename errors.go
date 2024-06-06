package mux

import (
	"net/http"

	"nhooyr.io/websocket"
)

type InternalError interface {
	error
	Internal() error
}

type HTTPError interface {
	error
	HTTPStatus() int
}

type httpError struct {
	statusCode int
	message    string
	internal   error
}

func NewHTTPError(statusCode int, message string) HTTPError {
	return &httpError{
		statusCode: statusCode,
		message:    message,
		internal:   nil,
	}
}

func NewInternalHTTPError(err error) HTTPError {
	return &httpError{
		statusCode: http.StatusInternalServerError,
		message:    http.StatusText(http.StatusInternalServerError),
		internal:   err,
	}
}

func (e *httpError) HTTPStatus() int {
	return e.statusCode
}

func (e *httpError) Error() string {
	return e.message
}

func (e *httpError) Internal() error {
	return e.internal
}

type WebSocketError interface {
	error
	WebSocketStatus() websocket.StatusCode
}

type webSocketError struct {
	statusCode websocket.StatusCode
	message    string
	internal   error
}

func NewWebSocketError(statusCode websocket.StatusCode, message string) WebSocketError {
	return &webSocketError{
		statusCode: statusCode,
		message:    message,
		internal:   nil,
	}
}

func NewInternalWebSocketError(err error) WebSocketError {
	return &webSocketError{
		statusCode: websocket.StatusInternalError,
		message:    websocket.StatusInternalError.String(),
		internal:   err,
	}
}

func (e *webSocketError) WebSocketStatus() websocket.StatusCode {
	return e.statusCode
}

func (e *webSocketError) Error() string {
	return e.message
}

func (e *webSocketError) Internal() error {
	return e.internal
}
