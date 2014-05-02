package kocha

import (
	"io"
	"net/http"
)

// Result is the interface that result.
type Result interface {
	Proc(*Response)
}

// resultContent represents a result of any content.
type resultContent struct {
	// The content body.
	Body io.Reader
}

// Proc writes content to response.
//
// If Body implements io.Closer interface, call Body.Close() in end of Proc.
func (r *resultContent) Proc(res *Response) {
	if closer, ok := r.Body.(io.Closer); ok {
		defer closer.Close()
	}
	if _, err := io.Copy(res, r.Body); err != nil {
		panic(err)
	}
}

// resultRedirect represents a result of redirect.
type resultRedirect struct {
	Request *Request

	// URL for redirect.
	URL string

	// Whether the redirect with 301 Moved Permanently.
	Permanently bool
}

// Proc writes redirect header to response.
func (r *resultRedirect) Proc(res *Response) {
	statusCode := http.StatusFound
	if r.Permanently {
		statusCode = http.StatusMovedPermanently
	}
	http.Redirect(res, r.Request.Request, r.URL, statusCode)
}
