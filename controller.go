package kocha

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
)

type mimeTypeFormats map[string]string

var MimeTypeFormats = mimeTypeFormats{
	"application/json": "json",
	"application/xml":  "xml",
	"text/html":        "html",
	"text/plain":       "txt",
}

func (m mimeTypeFormats) Get(mimeType string) string {
	return m[mimeType]
}

func (m mimeTypeFormats) Set(mimeType, format string) {
	m[mimeType] = format
}

func (m mimeTypeFormats) Del(mimeType string) {
	delete(m, mimeType)
}

type Controller struct {
	Name     string
	Request  *Request
	Response *Response
	Params   Params
	Session  Session
}

type Context map[string]interface{}

func (c *Controller) Render(context ...Context) Result {
	var ctx Context
	switch len(context) {
	case 0: // do nothing
	case 1:
		ctx = context[0]
	default: // > 1
		panic(errors.New("too many arguments"))
	}
	c.setContentTypeIfNotExists("text/html")
	format := MimeTypeFormats.Get(c.Response.ContentType)
	if format == "" {
		panic(fmt.Errorf("unknown Content-Type: %v", c.Response.ContentType))
	}
	t := appConfig.TemplateSet.Get(appConfig.AppName, c.Name, format)
	if t == nil {
		panic(errors.New("no such template: " + appConfig.TemplateSet.Ident(appConfig.AppName, c.Name, format)))
	}
	return &ResultTemplate{
		Template: t,
		Context:  ctx,
	}
}

func (c *Controller) RenderJSON(context interface{}) Result {
	c.setContentTypeIfNotExists("application/json")
	return &ResultJSON{
		Context: context,
	}
}

func (c *Controller) RenderXML(context interface{}) Result {
	c.setContentTypeIfNotExists("application/xml")
	return &ResultXML{
		Context: context,
	}
}

func (c *Controller) RenderText(content string) Result {
	c.setContentTypeIfNotExists("text/plain")
	return &ResultText{
		Content: content,
	}
}

func (c *Controller) RenderError(statusCode int, context ...Context) Result {
	var ctx Context
	switch len(context) {
	case 0: // do nothing
	case 1:
		ctx = context[0]
	default: // > 1
		panic(errors.New("too many arguments"))
	}
	c.setContentTypeIfNotExists("text/html")
	format := MimeTypeFormats.Get(c.Response.ContentType)
	if format == "" {
		panic(fmt.Errorf("unknown Content-Type: %v", c.Response.ContentType))
	}
	c.Response.StatusCode = statusCode
	name := filepath.Join("errors", strconv.Itoa(statusCode))
	t := appConfig.TemplateSet.Get(appConfig.AppName, name, format)
	if t == nil {
		c.Response.ContentType = "text/plain"
		return &ResultText{
			Content: http.StatusText(statusCode),
		}
	}
	return &ResultTemplate{
		Template: t,
		Context:  ctx,
	}
}

func (c *Controller) setContentTypeIfNotExists(contentType string) {
	if c.Response.ContentType == "" {
		c.Response.ContentType = contentType
	}
}

func (c *Controller) Redirect(url string, permanently bool) Result {
	return &ResultRedirect{
		Request:     c.Request,
		URL:         url,
		Permanently: permanently,
	}
}

type ErrorController struct {
	Controller
	StatusCode int
}

func NewErrorController(statusCode int) *ErrorController {
	return &ErrorController{
		StatusCode: statusCode,
	}
}

func (c *ErrorController) Get() Result {
	return c.RenderError(c.StatusCode)
}

type Params struct {
	url.Values
}
