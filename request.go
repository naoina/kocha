package kocha

import (
	"net/http"
	"os"
	"strings"
)

// Request represents a request.
type Request struct {
	*http.Request
}

// newRequest returns a new Request that given a *http.Request.
func newRequest(req *http.Request) *Request {
	return &Request{
		Request: req,
	}
}

// Scheme returns current scheme of HTTP connection.
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

// IsSSL returns whether the current connection is secure.
func (r *Request) IsSSL() bool {
	return r.Scheme() == "https"
}
