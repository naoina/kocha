package kocha

import (
	"net/http"
)

type Response struct {
	http.ResponseWriter
	ContentType string
	StatusCode  int
}

func NewResponse(rw http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: rw,
		ContentType:    "",
		StatusCode:     http.StatusOK,
	}
}
