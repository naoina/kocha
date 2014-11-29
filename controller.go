package kocha

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/naoina/denco"
)

// Controller is the interface that the request controller.
type Controller interface {
	Getter
	Poster
	Putter
	Deleter
	Header
	Patcher
}

// Getter interface is an interface representing a handler for HTTP GET request.
type Getter interface {
	GET(c *Context) Result
}

// Poster interface is an interface representing a handler for HTTP POST request.
type Poster interface {
	POST(c *Context) Result
}

// Putter interface is an interface representing a handler for HTTP PUT request.
type Putter interface {
	PUT(c *Context) Result
}

// Deleter interface is an interface representing a handler for HTTP DELETE request.
type Deleter interface {
	DELETE(c *Context) Result
}

// Header interface is an interface representing a handler for HTTP HEAD request.
type Header interface {
	HEAD(c *Context) Result
}

// Patcher interface is an interface representing a handler for HTTP PATCH request.
type Patcher interface {
	PATCH(c *Context) Result
}

type requestHandler func(c *Context) Result

// DefaultController implements Controller interface.
// This can be used to save the trouble to implement all of the methods of
// Controller interface.
type DefaultController struct {
}

// GET implements Getter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) GET(c *Context) Result {
	return RenderError(c, http.StatusMethodNotAllowed)
}

// POST implements Poster interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) POST(c *Context) Result {
	return RenderError(c, http.StatusMethodNotAllowed)
}

// PUT implements Putter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) PUT(c *Context) Result {
	return RenderError(c, http.StatusMethodNotAllowed)
}

// DELETE implements Deleter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) DELETE(c *Context) Result {
	return RenderError(c, http.StatusMethodNotAllowed)
}

// HEAD implements Header interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) HEAD(c *Context) Result {
	return RenderError(c, http.StatusMethodNotAllowed)
}

// PATCH implements Patcher interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) PATCH(c *Context) Result {
	return RenderError(c, http.StatusMethodNotAllowed)
}

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

// Context represents a context of each request.
type Context struct {
	Name     string       // controller name.
	Layout   string       // layout name.
	Format   string       // format of response.
	Data     interface{}  // data for template.
	Request  *Request     // request.
	Response *Response    // response.
	Params   *Params      // parameters of form values.
	Session  Session      // session.
	Flash    Flash        // flash messages.
	App      *Application // an application.

	// Errors represents the map of errors that related to the form values.
	// A map key is field name, and value is slice of errors.
	// Errors will be set by Context.Params.Bind().
	Errors map[string][]*ParamError
}

// Invoke is shorthand of c.App.Invoke.
func (c *Context) Invoke(unit Unit, newFunc func(), defaultFunc func()) {
	c.App.Invoke(unit, newFunc, defaultFunc)
}

func (c *Context) setContentTypeIfNotExists(contentType string) {
	if c.Response.ContentType == "" {
		c.Response.ContentType = contentType
	}
}

func (c *Context) setData(data []interface{}) error {
	switch len(data) {
	case 1:
		d, ok := data[0].(Data)
		if !ok {
			c.Data = data[0]
			return nil
		}
		if data, ok := c.Data.(Data); ok {
			if data == nil {
				data = Data{}
			}
			for k, v := range d {
				data[k] = v
			}
			d = data
		}
		c.Data = d
	case 0:
		// do nothing.
	default: // > 1
		return fmt.Errorf("too many arguments")
	}
	return nil
}

func (c *Context) setFormatFromContentTypeIfNotExists() error {
	if c.Format != "" {
		return nil
	}
	if c.Format = MimeTypeFormats.Get(c.Response.ContentType); c.Format == "" {
		return fmt.Errorf("kocha: unknown Content-Type: %v", c.Response.ContentType)
	}
	return nil
}

func (c *Context) prepareRequest(params denco.Params) error {
	c.Request.Body = http.MaxBytesReader(c.Response, c.Request.Body, c.App.Config.MaxClientBodySize)
	if err := c.Request.ParseMultipartForm(c.App.Config.MaxClientBodySize); err != nil && err != http.ErrNotMultipart {
		return err
	}
	for _, param := range params {
		c.Request.Form.Add(param.Name, param.Value)
	}
	return nil
}

func (c *Context) prepareParams() error {
	c.Params = newParams(c, c.Request.Form, "")
	return nil
}

// Data is shorthand type for map[string]interface{}
type Data map[string]interface{}

// String returns string of a map that sorted by keys.
func (c Data) String() string {
	keys := make([]string, 0, len(c))
	for key, _ := range c {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		keys[i] = fmt.Sprintf("%v:%v", key, c[key])
	}
	return fmt.Sprintf("map[%v]", strings.Join(keys, " "))
}

// StaticServe is generic controller for serve a static file.
type StaticServe struct {
	*DefaultController
}

func (ss *StaticServe) GET(c *Context) Result {
	path, err := url.Parse(c.Params.Get("path"))
	if err != nil {
		return RenderError(c, http.StatusBadRequest)
	}
	return SendFile(c, path.Path)
}

var internalServerErrorController = &ErrorController{
	StatusCode: http.StatusInternalServerError,
}

// ErrorController is generic controller for error response.
type ErrorController struct {
	*DefaultController

	StatusCode int
}

func (ec *ErrorController) GET(c *Context) Result {
	return RenderError(c, ec.StatusCode)
}
