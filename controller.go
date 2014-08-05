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
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/naoina/denco"
	"github.com/naoina/kocha/util"
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
	return c.RenderError(http.StatusMethodNotAllowed)
}

// POST implements Poster interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) POST(c *Context) Result {
	return c.RenderError(http.StatusMethodNotAllowed)
}

// PUT implements Putter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) PUT(c *Context) Result {
	return c.RenderError(http.StatusMethodNotAllowed)
}

// DELETE implements Deleter interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) DELETE(c *Context) Result {
	return c.RenderError(http.StatusMethodNotAllowed)
}

// HEAD implements Header interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) HEAD(c *Context) Result {
	return c.RenderError(http.StatusMethodNotAllowed)
}

// PATCH implements Patcher interface that renders the HTTP 405 Method Not Allowed.
func (dc *DefaultController) PATCH(c *Context) Result {
	return c.RenderError(http.StatusMethodNotAllowed)
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
	Data     Data         // data for template.
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

// Render returns result of template.
//
// The data variadic argument must be without specified or only one.
// A data to used will be determined the according to the following rules.
//
// 1. If data of the Data type is given, it will be merged with Context.Data and it will be used.
//
// 2. If data of an other type is given and Context.Data hasn't been set, it will be used as it is.
//    Or it panics if Context.Data has been set.
//
// 3. If data isn't given, Context.Data will be used.
//
// Render retrieve a template file from controller name and c.Response.ContentType.
// e.g. If controller name is "root" and ContentType is "application/xml", Render will
// try to retrieve the template file "root.xml".
// Also ContentType set to "text/html" if not specified.
func (c *Context) Render(data ...interface{}) Result {
	ctx, err := c.buildData(data)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("text/html")
	format := MimeTypeFormats.Get(c.Response.ContentType)
	if format == "" {
		panic(fmt.Errorf("unknown Content-Type: %v", c.Response.ContentType))
	}
	t := c.App.Template.Get(c.App.Config.AppName, c.Layout, c.Name, format)
	if t == nil {
		panic(errors.New("no such template: " + c.App.Template.Ident(c.App.Config.AppName, c.Layout, c.Name, format)))
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		panic(err)
	}
	return &resultContent{
		Body: &buf,
	}
}

// RenderJSON returns result of JSON.
//
// RenderJSON is similar to Render but data will be encoded to JSON.
// ContentType set to "application/json" if not specified.
func (c *Context) RenderJSON(data ...interface{}) Result {
	ctx, err := c.buildData(data)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("application/json")
	buf, err := json.Marshal(ctx)
	if err != nil {
		panic(err)
	}
	return &resultContent{
		Body: bytes.NewReader(buf),
	}
}

// RenderXML returns result of XML.
//
// RenderXML is similar to Render but data will be encoded to XML.
// ContentType set to "application/xml" if not specified.
func (c *Context) RenderXML(data ...interface{}) Result {
	ctx, err := c.buildData(data)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("application/xml")
	buf, err := xml.Marshal(ctx)
	if err != nil {
		panic(err)
	}
	return &resultContent{
		Body: bytes.NewReader(buf),
	}
}

// RenderText returns result of text.
//
// ContentType set to "text/plain" if not specified.
func (c *Context) RenderText(content string) Result {
	c.setContentTypeIfNotExists("text/plain")
	return &resultContent{
		Body: strings.NewReader(content),
	}
}

// RenderError returns result of error.
//
// RenderError is similar to Render, but there is a point where some different.
// Render retrieve a template file from statusCode and c.Response.ContentType.
// e.g. If statusCode is 500 and ContentType is "application/xml", Render will
// try to retrieve the template file "errors/500.xml".
// If failed to retrieve the template file, it returns result of text with statusCode.
// Also ContentType set to "text/html" if not specified.
func (c *Context) RenderError(statusCode int, data ...interface{}) Result {
	ctx, err := c.buildData(data)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("text/html")
	format := MimeTypeFormats.Get(c.Response.ContentType)
	if format == "" {
		panic(fmt.Errorf("unknown Content-Type: %v", c.Response.ContentType))
	}
	c.Response.StatusCode = statusCode
	name := filepath.Join("error", strconv.Itoa(statusCode))
	t := c.App.Template.Get(c.App.Config.AppName, c.Layout, name, format)
	if t == nil {
		c.Response.ContentType = "text/plain"
		return &resultContent{
			Body: bytes.NewReader([]byte(http.StatusText(statusCode))),
		}
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		panic(err)
	}
	return &resultContent{
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
func (c *Context) SendFile(path string) Result {
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
			return c.RenderError(http.StatusNotFound)
		}
		var err error
		if file, err = os.Open(path); err != nil {
			panic(err)
		}
	}
	c.Response.ContentType = util.DetectContentTypeByExt(path)
	if c.Response.ContentType == "" {
		c.Response.ContentType = util.DetectContentTypeByBody(file)
	}
	return &resultContent{
		Body: file,
	}
}

func (c *Context) setContentTypeIfNotExists(contentType string) {
	if c.Response.ContentType == "" {
		c.Response.ContentType = contentType
	}
}

// Redirect returns result of redirect.
//
// If permanently is true, redirect to url with 301. (http.StatusMovedPermanently)
// Otherwise redirect to url with 302. (http.StatusFound)
func (c *Context) Redirect(url string, permanently bool) Result {
	return &resultRedirect{
		Request:     c.Request,
		URL:         url,
		Permanently: permanently,
	}
}

func (c *Context) buildData(data []interface{}) (interface{}, error) {
	switch len(data) {
	case 1:
		ctx, ok := data[0].(Data)
		if !ok {
			if len(c.Data) == 0 {
				return data[0], nil
			}
			return nil, fmt.Errorf("data of multiple types has been set: Context.Data has been set,"+
				" but data of other type was given: %v", reflect.TypeOf(data))
		}
		if c.Data == nil {
			c.Data = Data{}
		}
		for k, v := range ctx {
			c.Data[k] = v
		}
	case 0:
		if c.Data == nil {
			c.Data = Data{}
		}
	default: // > 1
		return nil, fmt.Errorf("too many arguments")
	}
	if _, exists := c.Data["errors"]; exists {
		c.App.Logger.Warn("kocha: Data: `errors' key has already been set, skipped")
	} else {
		c.Data["errors"] = c.Errors
	}
	return c.Data, nil
}

// Invoke is shorthand of c.App.Invoke.
func (c *Context) Invoke(unit Unit, newFunc func(), defaultFunc func()) {
	c.App.Invoke(unit, newFunc, defaultFunc)
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

// StaticServe is generic controller for serve a static file.
type StaticServe struct {
	*DefaultController
}

func (ss *StaticServe) GET(c *Context) Result {
	path, err := url.Parse(c.Params.Get("path"))
	if err != nil {
		return c.RenderError(http.StatusBadRequest)
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

func (ec *ErrorController) GET(c *Context) Result {
	return c.RenderError(ec.StatusCode)
}
