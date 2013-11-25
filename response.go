package kocha

import (
	"net/http"
)

// Response represents a response.
type Response struct {
	http.ResponseWriter
	ContentType string
	StatusCode  int
	cookies     []*http.Cookie
}

// NewResponse returns a new Response that responds to rw.
func NewResponse(rw http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: rw,
		ContentType:    "",
		StatusCode:     http.StatusOK,
	}
}

// Cookies returns a slice of *http.Cookie.
func (r *Response) Cookies() []*http.Cookie {
	return r.cookies
}

// SetCookie adds a Set-Cookie header to the response.
func (r *Response) SetCookie(cookie *http.Cookie) {
	r.cookies = append(r.cookies, cookie)
	http.SetCookie(r, cookie)
}
