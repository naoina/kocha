package kocha

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

// Result is the interface that result.
type Result interface {
	Proc(*Response)
}

// ResultTemplate represents a result of template render.
type ResultTemplate struct {
	Template *template.Template
	Context  Context
}

// Proc executes the template and write to response.
func (r *ResultTemplate) Proc(res *Response) {
	if err := r.Template.Execute(res, r.Context); err != nil {
		panic(err)
	}
}

// ResultJSON represents a result of JSON render.
type ResultJSON struct {
	Context interface{}
}

// Proc encodes to JSON and write to response.
func (r *ResultJSON) Proc(res *Response) {
	if err := json.NewEncoder(res).Encode(r.Context); err != nil {
		panic(err)
	}
}

// ResultXML represents a result of XML render.
type ResultXML struct {
	Context interface{}
}

// Proc encodes to XML and write to response.
func (r *ResultXML) Proc(res *Response) {
	if err := xml.NewEncoder(res).Encode(r.Context); err != nil {
		panic(err)
	}
}

// ResultText represents a result of text render.
type ResultText struct {
	Content string
}

// Proc writes content to response.
func (r *ResultText) Proc(res *Response) {
	if _, err := fmt.Fprint(res, r.Content); err != nil {
		panic(err)
	}
}

// ResultRedirect represents a result of redirect.
type ResultRedirect struct {
	Request *Request

	// URL for redirect.
	URL string

	// Whether the redirect with 301 Moved Permanently.
	Permanently bool
}

// Proc writes redirect header to response.
func (r *ResultRedirect) Proc(res *Response) {
	statusCode := http.StatusFound
	if r.Permanently {
		statusCode = http.StatusMovedPermanently
	}
	http.Redirect(res, r.Request.Request, r.URL, statusCode)
}

// ResultContent represents a result of any content.
type ResultContent struct {
	// The content body.
	Reader io.Reader
}

// Proc writes content to response.
//
// If Reader implements io.Closer interface, call Reader.Close() in end of Proc.
func (r *ResultContent) Proc(res *Response) {
	if closer, ok := r.Reader.(io.Closer); ok {
		defer closer.Close()
	}
	if _, err := io.Copy(res, r.Reader); err != nil {
		panic(err)
	}
}
