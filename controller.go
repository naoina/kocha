package kocha

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type mimeTypeFormats map[string]string

// MimeTypeFormats is relation between mime type and file extension.
var MimeTypeFormats = mimeTypeFormats{
	"application/json": "json",
	"application/xml":  "xml",
	"text/html":        "html",
	"text/plain":       "txt",
}

// Get returns the file extension from the mime type.
func (m mimeTypeFormats) Get(mimeType string) string {
	return m[mimeType]
}

// Set set the file extension to the mime type.
func (m mimeTypeFormats) Set(mimeType, format string) {
	m[mimeType] = format
}

// Del delete the mime type and file extension.
func (m mimeTypeFormats) Del(mimeType string) {
	delete(m, mimeType)
}

// Controller is the base controller.
type Controller struct {
	// Name of controller.
	Name string

	// Request.
	Request *Request

	// Response.
	Response *Response

	// Parameters of form values.
	Params Params

	// Session.
	Session Session
}

// Context is shorthand type for map[string]interface{}
type Context map[string]interface{}

// Render returns result of template.
//
// The context variadic argument must be without specified or only one.
// Render retrieve a template file from controller name and c.Response.ContentType.
// e.g. If controller name is "root" and ContentType is "application/xml", Render will
// try to retrieve the template file "root.xml".
// Also ContentType set to "text/html" if not specified.
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
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		panic(err)
	}
	return &ResultContent{
		Body: &buf,
	}
}

// RenderJSON returns result of JSON.
//
// ContentType set to "application/json" if not specified.
func (c *Controller) RenderJSON(context interface{}) Result {
	c.setContentTypeIfNotExists("application/json")
	buf, err := json.Marshal(context)
	if err != nil {
		panic(err)
	}
	return &ResultContent{
		Body: bytes.NewReader(buf),
	}
}

// RenderXML returns result of XML.
//
// ContentType set to "application/xml" if not specified.
func (c *Controller) RenderXML(context interface{}) Result {
	c.setContentTypeIfNotExists("application/xml")
	buf, err := xml.Marshal(context)
	if err != nil {
		panic(err)
	}
	return &ResultContent{
		Body: bytes.NewReader(buf),
	}
}

// RenderText returns result of text.
//
// ContentType set to "text/plain" if not specified.
func (c *Controller) RenderText(content string) Result {
	c.setContentTypeIfNotExists("text/plain")
	return &ResultContent{
		Body: bytes.NewReader([]byte(content)),
	}
}

// RenderError returns result of error.
//
// The context variadic argument must be without specified or only one.
// Render retrieve a template file from statusCode and c.Response.ContentType.
// e.g. If statusCode is 500 and ContentType is "application/xml", Render will
// try to retrieve the template file "errors/500.xml".
// If failed to retrieve the template file, it returns result of text with statusCode.
// Also ContentType set to "text/html" if not specified.
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
		return &ResultContent{
			Body: bytes.NewReader([]byte(http.StatusText(statusCode))),
		}
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		panic(err)
	}
	return &ResultContent{
		Body: &buf,
	}
}

// Sendfile returns result of any content.
//
// The path argument specifies an absolute or relative path.
// If absolute path, read the content from the path as it is.
// If relative path, First, Try to get the content from included resources and
// returns it if successful. Otherwise, Add AppPath and StaticDir to the prefix
// of the path and then will read the content from the path that.
// Also, set ContentType detect from content if c.Response.ContentType is empty.
func (c *Controller) SendFile(path string) Result {
	var file io.ReadSeeker
	path = filepath.FromSlash(path)
	if rc, ok := includedResources[path]; ok {
		file = rc.Open()
	} else {
		if !filepath.IsAbs(path) {
			path = filepath.Join(appConfig.AppPath, StaticDir, path)
		}
		if _, err := os.Stat(path); err != nil {
			return c.RenderError(http.StatusNotFound)
		}
		var err error
		if file, err = os.Open(path); err != nil {
			panic(err)
		}
	}
	if c.Response.ContentType == "" {
		c.Response.ContentType = detectContentType(file)
	}
	return &ResultContent{
		Body: file,
	}
}

func (c *Controller) setContentTypeIfNotExists(contentType string) {
	if c.Response.ContentType == "" {
		c.Response.ContentType = contentType
	}
}

// Redirect returns result of redirect.
//
// If permanently is true, redirect to url with 301. (http.StatusMovedPermanently)
// Otherwise redirect to url with 302. (http.StatusFound)
func (c *Controller) Redirect(url string, permanently bool) Result {
	return &ResultRedirect{
		Request:     c.Request,
		URL:         url,
		Permanently: permanently,
	}
}

// StaticServe is generic controller for serve a static file.
type StaticServe struct {
	*Controller
}

func (c *StaticServe) Get(path *url.URL) Result {
	return c.SendFile(path.Path)
}

// ErrorController is generic controller for error response.
type ErrorController struct {
	*Controller
	StatusCode int
}

// NewErrorController returns a new ErrorController from statusCode.
func NewErrorController(statusCode int) *ErrorController {
	return &ErrorController{
		StatusCode: statusCode,
	}
}

func (c *ErrorController) Get() Result {
	return c.RenderError(c.StatusCode)
}

// Params is represents form values.
type Params struct {
	url.Values
}
