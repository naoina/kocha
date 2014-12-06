package kocha

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/naoina/kocha/util"
)

// Render returns result of template.
//
// A data to used will be determined the according to the following rules.
//
// 1. If data of the Data type is given, it will be merged to Context.Data.
//
// 2. If data of another type is given, it will be set to Context.Data.
//
// 3. If data is nil, Context.Data as is.
//
// Render retrieve a template file from controller name and c.Response.ContentType.
// e.g. If controller name is "root" and ContentType is "application/xml", Render will
// try to retrieve the template file "root.xml".
// Also ContentType set to "text/html" if not specified.
func Render(c *Context, data interface{}) Result {
	c.setData(data)
	c.setContentTypeIfNotExists("text/html")
	if err := c.setFormatFromContentTypeIfNotExists(); err != nil {
		panic(err)
	}
	t := c.App.Template.Get(c.App.Config.AppName, c.Layout, c.Name, c.Format)
	if t == nil {
		panic(errors.New("kocha: no such template: " + c.App.Template.Ident(c.App.Config.AppName, c.Layout, c.Name, c.Format)))
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
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
func RenderJSON(c *Context, data interface{}) Result {
	c.setData(data)
	c.setContentTypeIfNotExists("application/json")
	buf, err := json.Marshal(c.Data)
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
func RenderXML(c *Context, data interface{}) Result {
	c.setData(data)
	c.setContentTypeIfNotExists("application/xml")
	buf, err := xml.Marshal(c.Data)
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
func RenderText(c *Context, content string) Result {
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
func RenderError(c *Context, statusCode int, data interface{}) Result {
	c.setData(data)
	c.setContentTypeIfNotExists("text/html")
	if err := c.setFormatFromContentTypeIfNotExists(); err != nil {
		panic(err)
	}
	c.Response.StatusCode = statusCode
	c.Name = errorTemplateName(statusCode)
	t := c.App.Template.Get(c.App.Config.AppName, c.Layout, c.Name, c.Format)
	if t == nil {
		c.Response.ContentType = "text/plain"
		return &resultContent{
			Body: bytes.NewReader([]byte(http.StatusText(statusCode))),
		}
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		panic(err)
	}
	return &resultContent{
		Body: &buf,
	}
}

// SendFile returns result of any content.
//
// The path argument specifies an absolute or relative path.
// If absolute path, read the content from the path as it is.
// If relative path, First, Try to get the content from included resources and
// returns it if successful. Otherwise, Add AppPath and StaticDir to the prefix
// of the path and then will read the content from the path that.
// Also, set ContentType detect from content if c.Response.ContentType is empty.
func SendFile(c *Context, path string) Result {
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
			return RenderError(c, http.StatusNotFound, nil)
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

// Redirect returns result of redirect.
//
// If permanently is true, redirect to url with 301. (http.StatusMovedPermanently)
// Otherwise redirect to url with 302. (http.StatusFound)
func Redirect(c *Context, url string, permanently bool) Result {
	return &resultRedirect{
		Request:     c.Request,
		URL:         url,
		Permanently: permanently,
	}
}
