package kocha

import (
	"net/http"
)

type Response struct {
	http.ResponseWriter
	ContentType string
	StatusCode  int
	cookies     []*http.Cookie
}

func NewResponse(rw http.ResponseWriter) *Response {
	return &Response{
		ResponseWriter: rw,
		ContentType:    "",
		StatusCode:     http.StatusOK,
	}
}

func (r *Response) Cookies() []*http.Cookie {
	return r.cookies
}

func (r *Response) SetCookie(cookie *http.Cookie) {
	r.cookies = append(r.cookies, cookie)
	http.SetCookie(r, cookie)
}
