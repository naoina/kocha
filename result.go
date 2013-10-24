package kocha

import (
	"encoding/json"
	"encoding/xml"
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
	if err := json.NewEncoder(writer).Encode(r.Context); err != nil {
		panic(err)
	}
}

type ResultXML struct {
	Context interface{}
}

func (r *ResultXML) Proc(writer http.ResponseWriter) {
	if err := xml.NewEncoder(writer).Encode(r.Context); err != nil {
		panic(err)
	}
}
