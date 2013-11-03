package kocha

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
)

type Result interface {
	Proc(*Response)
}

type ResultTemplate struct {
	Template *template.Template
	Context  Context
}

func (r *ResultTemplate) Proc(res *Response) {
	if err := r.Template.Execute(res, r.Context); err != nil {
		panic(err)
	}
}

type ResultJSON struct {
	Context interface{}
}

func (r *ResultJSON) Proc(res *Response) {
	if err := json.NewEncoder(res).Encode(r.Context); err != nil {
		panic(err)
	}
}

type ResultXML struct {
	Context interface{}
}

func (r *ResultXML) Proc(res *Response) {
	if err := xml.NewEncoder(res).Encode(r.Context); err != nil {
		panic(err)
	}
}

type ResultText struct {
	Content string
}

func (r *ResultText) Proc(res *Response) {
	if _, err := fmt.Fprint(res, r.Content); err != nil {
		panic(err)
	}
}

type ResultRedirect struct {
	Request     *Request
	URL         string
	Permanently bool
}

func (r *ResultRedirect) Proc(res *Response) {
	statusCode := http.StatusFound
	if r.Permanently {
		statusCode = http.StatusMovedPermanently
	}
	http.Redirect(res, r.Request.Request, r.URL, statusCode)
}
