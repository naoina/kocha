package kocha

import (
	"net/http"
)

type Response struct {
	http.ResponseWriter
}

func NewResponse(rw http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: rw,
	}
}
