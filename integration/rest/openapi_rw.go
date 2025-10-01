package rest

import (
	"bytes"
	"net/http"
)

/*
responseWriter wraps the standard http.ResponseWriter so we can store additional
values during the request/response lifecycle, such as the status code and the
the response body.
*/
type responseWriter struct {
	http.ResponseWriter

	// status code is the HTTP status code sets in the response header. This allows
	// to ensure if the status code respects the one defined in the OpenAPI
	// description.
	status int

	// buf is the HTTP response body sets by a handler function. This allows to
	// ensure if the body respects the one defined in the OpenAPI description.
	buf *bytes.Buffer
}

/*
Write writes the data to the connection as part of an HTTP reply.
*/
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.ResponseWriter.Write(b)
	return rw.buf.Write(b)
}

/*
WriteHeader sends an HTTP response header with the provided status code.
*/
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
