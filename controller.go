package kocha

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

var contextPool = &sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

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
	GET(c *Context) error
}

// Poster interface is an interface representing a handler for HTTP POST request.
type Poster interface {
	POST(c *Context) error
}

// Putter interface is an interface representing a handler for HTTP PUT request.
type Putter interface {
	PUT(c *Context) error
}

// Deleter interface is an interface representing a handler for HTTP DELETE request.
type Deleter interface {
	DELETE(c *Context) error
}

// Header interface is an interface representing a handler for HTTP HEAD request.
type Header interface {
	HEAD(c *Context) error
}

// Patcher interface is an interface representing a handler for HTTP PATCH request.
type Patcher interface {
	PATCH(c *Context) error
}

type requestHandler func(c *Context) error

// DefaultController implements Controller interface.
// This can be used to save the trouble to implement all of the methods of
// Controller interface.
type DefaultController struct {
}

// GET implements Getter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) GET(c *Context) error {
	return c.RenderError(http.StatusMethodNotAllowed, nil, nil)
}

// POST implements Poster interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) POST(c *Context) error {
	return c.RenderError(http.StatusMethodNotAllowed, nil, nil)
}

// PUT implements Putter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) PUT(c *Context) error {
	return c.RenderError(http.StatusMethodNotAllowed, nil, nil)
}

// DELETE implements Deleter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) DELETE(c *Context) error {
	return c.RenderError(http.StatusMethodNotAllowed, nil, nil)
}

// HEAD implements Header interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) HEAD(c *Context) error {
	return c.RenderError(http.StatusMethodNotAllowed, nil, nil)
}

// PATCH implements Patcher interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) PATCH(c *Context) error {
	return c.RenderError(http.StatusMethodNotAllowed, nil, nil)
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
	Name     string       // route name of the controller.
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

func newContext() *Context {
	c := contextPool.Get().(*Context)
	c.reset()
	return c
}

// ErrorWithLine returns error that added the filename and line to err.
func ErrorWithLine(err error) error {
	return errorWithLine(err, 2)
}

func errorWithLine(err error, calldepth int) error {
	if _, file, line, ok := runtime.Caller(calldepth); ok {
		return fmt.Errorf("%s:%d: %v", file, line, err)
	}
	return err
}

// Render renders a template.
//
// A data to used will be determined the according to the following rules.
//
// 1. If data of any map type is given, it will be merged to Context.Data if possible.
//
// 2. If data of another type is given, it will be set to Context.Data.
//
// 3. If data is nil, Context.Data as is.
//
// Render retrieve a template file from controller name and c.Response.ContentType.
// e.g. If controller name is "root" and ContentType is "application/xml", Render will
// try to retrieve the template file "root.xml".
// Also ContentType set to "text/html" if not specified.
func (c *Context) Render(data interface{}) error {
	if err := c.setData(data); err != nil {
		return c.errorWithLine(err)
	}
	c.setContentTypeIfNotExists("text/html")
	if err := c.setFormatFromContentTypeIfNotExists(); err != nil {
		return c.errorWithLine(err)
	}
	t, err := c.App.Template.Get(c.Layout, c.Name, c.Format)
	if err != nil {
		return c.errorWithLine(err)
	}
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	if err := t.Execute(buf, c); err != nil {
		return fmt.Errorf("%s: %v", t.Name(), err)
	}
	if err := c.render(buf); err != nil {
		return c.errorWithLine(err)
	}
	return nil
}

// RenderJSON renders the data as JSON.
//
// RenderJSON is similar to Render but data will be encoded to JSON.
// ContentType set to "application/json" if not specified.
func (c *Context) RenderJSON(data interface{}) error {
	if err := c.setData(data); err != nil {
		return c.errorWithLine(err)
	}
	c.setContentTypeIfNotExists("application/json")
	buf, err := json.Marshal(c.Data)
	if err != nil {
		return c.errorWithLine(err)
	}
	if err := c.render(bytes.NewReader(buf)); err != nil {
		return c.errorWithLine(err)
	}
	return nil
}

// RenderXML renders the data as XML.
//
// RenderXML is similar to Render but data will be encoded to XML.
// ContentType set to "application/xml" if not specified.
func (c *Context) RenderXML(data interface{}) error {
	if err := c.setData(data); err != nil {
		return c.errorWithLine(err)
	}
	c.setContentTypeIfNotExists("application/xml")
	buf, err := xml.Marshal(c.Data)
	if err != nil {
		return c.errorWithLine(err)
	}
	if err := c.render(bytes.NewReader(buf)); err != nil {
		return c.errorWithLine(err)
	}
	return nil
}

// RenderText renders the content.
//
// ContentType set to "text/plain" if not specified.
func (c *Context) RenderText(content string) error {
	c.setContentTypeIfNotExists("text/plain")
	if err := c.render(strings.NewReader(content)); err != nil {
		return c.errorWithLine(err)
	}
	return nil
}

// RenderError renders an error page with statusCode.
//
// RenderError is similar to Render, but there is the points where some different.
// If err is not nil, RenderError outputs the err to log using c.App.Logger.Error.
// RenderError retrieves a template file from statusCode and c.Response.ContentType.
// e.g. If statusCode is 500 and ContentType is "application/xml", RenderError will
// try to retrieve the template file "errors/500.xml".
// If failed to retrieve the template file, it returns result of text with statusCode.
// Also ContentType set to "text/html" if not specified.
func (c *Context) RenderError(statusCode int, err error, data interface{}) error {
	if err != nil {
		c.App.Logger.Error(c.errorWithLine(err))
	}
	if err := c.setData(data); err != nil {
		return c.errorWithLine(err)
	}
	c.setContentTypeIfNotExists("text/html")
	if err := c.setFormatFromContentTypeIfNotExists(); err != nil {
		return c.errorWithLine(err)
	}
	c.Response.StatusCode = statusCode
	c.Name = errorTemplateName(statusCode)
	t, err := c.App.Template.Get(c.Layout, c.Name, c.Format)
	if err != nil {
		c.Response.ContentType = "text/plain"
		if err := c.render(bytes.NewReader([]byte(http.StatusText(statusCode)))); err != nil {
			return c.errorWithLine(err)
		}
		return nil
	}
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	if err := t.Execute(buf, c); err != nil {
		return fmt.Errorf("%s: %v", t.Name(), err)
	}
	if err := c.render(buf); err != nil {
		return c.errorWithLine(err)
	}
	return nil
}

// SendFile sends a content.
//
// The path argument specifies an absolute or relative path.
// If absolute path, read the content from the path as it is.
// If relative path, First, Try to get the content from included resources and
// returns it if successful. Otherwise, Add AppPath and StaticDir to the prefix
// of the path and then will read the content from the path that.
// Also, set ContentType detect from content if c.Response.ContentType is empty.
func (c *Context) SendFile(path string) error {
	var file io.ReadSeeker
	path = filepath.FromSlash(path)
	if rc := c.App.ResourceSet.Get(path); rc != nil {
		switch b := rc.(type) {
		case string:
			file = strings.NewReader(b)
		case []byte:
			file = bytes.NewReader(b)
		}
	}
	if file == nil {
		if !filepath.IsAbs(path) {
			path = filepath.Join(c.App.Config.AppPath, StaticDir, path)
		}
		if _, err := os.Stat(path); err != nil {
			if err := c.RenderError(http.StatusNotFound, nil, nil); err != nil {
				return c.errorWithLine(err)
			}
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return c.errorWithLine(err)
		}
		defer f.Close()
		file = f
	}
	c.Response.ContentType = mime.TypeByExtension(filepath.Ext(path))
	if c.Response.ContentType == "" {
		ct, err := c.detectContentTypeByBody(file)
		if err != nil {
			return err
		}
		c.Response.ContentType = ct
	}
	if err := c.render(file); err != nil {
		return c.errorWithLine(err)
	}
	return nil
}

// Redirect renders result of redirect.
//
// If permanently is true, redirect to url with 301. (http.StatusMovedPermanently)
// Otherwise redirect to url with 302. (http.StatusFound)
func (c *Context) Redirect(url string, permanently bool) error {
	if permanently {
		c.Response.StatusCode = http.StatusMovedPermanently
	} else {
		c.Response.StatusCode = http.StatusFound
	}
	http.Redirect(c.Response, c.Request.Request, url, c.Response.StatusCode)
	return nil
}

// Invoke is shorthand of c.App.Invoke.
func (c *Context) Invoke(unit Unit, newFunc func(), defaultFunc func()) {
	c.App.Invoke(unit, newFunc, defaultFunc)
}

// ErrorWithLine returns error that added the filename and line to err.
func (c *Context) ErrorWithLine(err error) error {
	return c.errorWithLine(err)
}

func (c *Context) render(r io.Reader) error {
	c.Response.Header().Set("Content-Type", c.Response.ContentType)
	c.Response.WriteHeader(c.Response.StatusCode)
	_, err := io.Copy(c.Response, r)
	return err
}

func (c *Context) detectContentTypeByBody(r io.Reader) (string, error) {
	buf := make([]byte, 512)
	if n, err := io.ReadFull(r, buf); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", err
		}
		buf = buf[:n]
	}
	if rs, ok := r.(io.Seeker); ok {
		if _, err := rs.Seek(0, os.SEEK_SET); err != nil {
			return "", err
		}
	}
	return http.DetectContentType(buf), nil
}

func (c *Context) setContentTypeIfNotExists(contentType string) {
	if c.Response.ContentType == "" {
		c.Response.ContentType = contentType
	}
}

func (c *Context) setData(data interface{}) error {
	if data == nil {
		return nil
	}
	srcType := reflect.TypeOf(data)
	if srcType.Kind() != reflect.Map {
		c.Data = data
		return nil
	}
	if c.Data == nil {
		c.Data = map[interface{}]interface{}{}
	}
	destType := reflect.TypeOf(c.Data)
	if sk, dk := srcType.Key(), destType.Key(); !sk.AssignableTo(dk) {
		return fmt.Errorf("kocha: context: key of type %v is not assignable to type %v", sk, dk)
	}
	src := reflect.ValueOf(data)
	dest := reflect.ValueOf(c.Data)
	dtype := destType.Elem()
	for _, k := range src.MapKeys() {
		v := src.MapIndex(k)
		if vtype := v.Type(); !vtype.AssignableTo(dtype) {
			if !vtype.ConvertibleTo(dtype) {
				return fmt.Errorf("kocha: context: value of type %v is not assignable to type %v", vtype, dtype)
			}
			v = v.Convert(dtype)
		}
		dest.SetMapIndex(k, v)
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

func (c *Context) newParams() *Params {
	if c.Request.Form == nil {
		c.Request.Form = url.Values{}
	}
	return newParams(c, c.Request.Form, "")
}

func (c *Context) errorWithLine(err error) error {
	return errorWithLine(err, 3)
}

func (c *Context) reset() {
	c.Name = ""
	c.Format = ""
	c.Params = nil
	c.Session = nil
	c.Flash = nil
}

func (c *Context) reuse() {
	c.Params.reuse()
	c.Request.reuse()
	contextPool.Put(c)
}

// StaticServe is generic controller for serve a static file.
type StaticServe struct {
	*DefaultController
}

func (ss *StaticServe) GET(c *Context) error {
	path, err := url.Parse(c.Params.Get("path"))
	if err != nil {
		return c.RenderError(http.StatusBadRequest, err, nil)
	}
	return c.SendFile(path.Path)
}

var internalServerErrorController = &ErrorController{
	StatusCode: http.StatusInternalServerError,
}

// ErrorController is generic controller for error response.
type ErrorController struct {
	*DefaultController

	StatusCode int
}

func (ec *ErrorController) GET(c *Context) error {
	return c.RenderError(ec.StatusCode, nil, nil)
}
