package kocha

import (
	"net/http"
)

type Response struct {
	http.ResponseWriter
	ContentType string
}

func NewResponse(rw http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: rw,
		ContentType:    "text/html",
	}
}
