package kocha

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
)

type Result interface {
	Proc(http.ResponseWriter)
}

type ResultTemplate struct {
	Template *template.Template
	Context  Context
}

func (r *ResultTemplate) Proc(writer http.ResponseWriter) {
	if err := r.Template.Execute(writer, r.Context); err != nil {
		panic(err)
	}
}

type ResultJSON struct {
	Context interface{}
}

func (r *ResultJSON) Proc(writer http.ResponseWriter) {
	setContentTypeIfNotExists(writer.Header(), "application/json")
	if err := json.NewEncoder(writer).Encode(r.Context); err != nil {
		panic(err)
	}
}

type ResultXML struct {
	Context interface{}
}

func (r *ResultXML) Proc(writer http.ResponseWriter) {
	setContentTypeIfNotExists(writer.Header(), "application/xml")
	if err := xml.NewEncoder(writer).Encode(r.Context); err != nil {
		panic(err)
	}
}

type ResultPlainText struct {
	Content string
}

func (r *ResultPlainText) Proc(writer http.ResponseWriter) {
	setContentTypeIfNotExists(writer.Header(), "text/plain")
	if _, err := fmt.Fprint(writer, r.Content); err != nil {
		panic(err)
	}
}

func setContentTypeIfNotExists(header http.Header, mimeType string) {
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", mimeType)
	}
}
