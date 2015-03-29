package kocha

import (
	"net/http"
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
	case r.Header.Get("Https") == "on", r.Header.Get("X-Forwarded-Ssl") == "on":
		return "https"
	case r.Header.Get("X-Forwarded-Scheme") != "":
		return r.Header.Get("X-Forwarded-Scheme")
	case r.Header.Get("X-Forwarded-Proto") != "":
		return strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]
	}
	return "http"
}

// IsSSL returns whether the current connection is secure.
func (r *Request) IsSSL() bool {
	return r.Scheme() == "https"
}

// IsXHR returns whether the XHR request.
func (r *Request) IsXHR() bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}
