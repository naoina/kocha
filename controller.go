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

	// Layout name to use.
	Layout string

	// Context value for template.
	Context Context

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

// String returns string of a map that sorted by keys.
func (c Context) String() string {
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
// The context variadic argument must be without specified or only one.
// A context to used will be determined the according to the following rules.
//
// 1. If context of the Context type is given, it will be merged with Controller.Context and it will be used.
//
// 2. If context of an other type is given and Controller.Context hasn't been set, it will be used as it is.
//    Or it panics if Controller.Context has been set.
//
// 3. If context isn't given, Controller.Context will be used.
//
// Render retrieve a template file from controller name and c.Response.ContentType.
// e.g. If controller name is "root" and ContentType is "application/xml", Render will
// try to retrieve the template file "root.xml".
// Also ContentType set to "text/html" if not specified.
func (c *Controller) Render(context ...interface{}) Result {
	ctx, err := c.buildContext(context)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("text/html")
	format := MimeTypeFormats.Get(c.Response.ContentType)
	if format == "" {
		panic(fmt.Errorf("unknown Content-Type: %v", c.Response.ContentType))
	}
	t := appConfig.templateMap.Get(appConfig.AppName, c.Layout, c.Name, format)
	if t == nil {
		panic(errors.New("no such template: " + appConfig.templateMap.Ident(appConfig.AppName, c.Layout, c.Name, format)))
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
// RenderJSON is similar to Render but context will be encoded to JSON.
// ContentType set to "application/json" if not specified.
func (c *Controller) RenderJSON(context ...interface{}) Result {
	ctx, err := c.buildContext(context)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("application/json")
	buf, err := json.Marshal(ctx)
	if err != nil {
		panic(err)
	}
	return &ResultContent{
		Body: bytes.NewReader(buf),
	}
}

// RenderXML returns result of XML.
//
// RenderXML is similar to Render but context will be encoded to XML.
// ContentType set to "application/xml" if not specified.
func (c *Controller) RenderXML(context ...interface{}) Result {
	ctx, err := c.buildContext(context)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("application/xml")
	buf, err := xml.Marshal(ctx)
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
func (c *Controller) RenderError(statusCode int, context ...interface{}) Result {
	ctx, err := c.buildContext(context)
	if err != nil {
		panic(err)
	}
	c.setContentTypeIfNotExists("text/html")
	format := MimeTypeFormats.Get(c.Response.ContentType)
	if format == "" {
		panic(fmt.Errorf("unknown Content-Type: %v", c.Response.ContentType))
	}
	c.Response.StatusCode = statusCode
	name := filepath.Join("errors", strconv.Itoa(statusCode))
	t := appConfig.templateMap.Get(appConfig.AppName, c.Layout, name, format)
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

func (c *Controller) buildContext(context []interface{}) (interface{}, error) {
	switch len(context) {
	case 1: // do nothing.
	case 0:
		return c.Context, nil
	default: // > 1
		return nil, errors.New("too many arguments")
	}
	if c.Context == nil {
		return context[0], nil
	}
	if ctx, ok := context[0].(Context); ok {
		for k, v := range ctx {
			c.Context[k] = v
		}
		return c.Context, nil
	}
	if len(c.Context) != 0 {
		return nil, fmt.Errorf("contexts of multiple types has been set: Controller.Context has been set,"+
			" but context of other type was given: %v", reflect.TypeOf(context))
	}
	return context[0], nil
}

// Invoke is an alias to Invoke(unit, newFunc, defaultFunc).
func (c *Controller) Invoke(unit Unit, newFunc func(), defaultFunc func()) {
	Invoke(unit, newFunc, defaultFunc)
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
