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
	res.Header().Set("Content-Type", res.ContentType)
	if err := r.Template.Execute(res, r.Context); err != nil {
		panic(err)
	}
}

type ResultJSON struct {
	Context interface{}
}

func (r *ResultJSON) Proc(res *Response) {
	setContentTypeIfNotExists(res.Header(), "application/json")
	if err := json.NewEncoder(res).Encode(r.Context); err != nil {
		panic(err)
	}
}

type ResultXML struct {
	Context interface{}
}

func (r *ResultXML) Proc(res *Response) {
	setContentTypeIfNotExists(res.Header(), "application/xml")
	if err := xml.NewEncoder(res).Encode(r.Context); err != nil {
		panic(err)
	}
}

type ResultPlainText struct {
	Content string
}

func (r *ResultPlainText) Proc(res *Response) {
	setContentTypeIfNotExists(res.Header(), "text/plain")
	if _, err := fmt.Fprint(res, r.Content); err != nil {
		panic(err)
	}
}

func setContentTypeIfNotExists(header http.Header, mimeType string) {
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", mimeType)
	}
}
