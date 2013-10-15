package kocha

import (
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
	r.Template.Execute(writer, r.Context)
}
