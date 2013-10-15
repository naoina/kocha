package kocha

import (
	"net/http"
)

type Request struct {
	*http.Request
}

func NewRequest(req *http.Request) *Request {
	return &Request{
		Request: req,
	}
}
