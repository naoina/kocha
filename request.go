package kocha

import (
	"net/http"
	"os"
	"strings"
)

type Request struct {
	*http.Request
}

func NewRequest(req *http.Request) *Request {
	return &Request{
		Request: req,
	}
}

func (r *Request) Scheme() string {
	switch {
	case os.Getenv("HTTPS") == "on", os.Getenv("HTTP_X_FORWARDED_SSL") == "on":
		return "https"
	case os.Getenv("HTTP_X_FORWARDED_SCHEME") != "":
		return os.Getenv("HTTP_X_FORWARDED_SCHEME")
	case os.Getenv("HTTP_X_FORWARDED_PROTO") != "":
		return strings.Split(os.Getenv("HTTP_X_FORWARDED_PROTO"), ",")[0]
	}
	return "http"
}

func (r *Request) IsSSL() bool {
	return r.Scheme() == "https"
}
