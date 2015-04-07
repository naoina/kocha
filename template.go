package kocha

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/naoina/kocha/util"
)

const (
	LayoutDir        = "layout"
	ErrorTemplateDir = "error"
)

// TemplatePathInfo represents an information of template paths.
type TemplatePathInfo struct {
	Name  string   // name of the application.
	Paths []string // directory paths of the template files.
}

type templateKey struct {
	appName string
	name    string
	format  string
}

// Template represents the templates information.
type Template struct {
	PathInfo   TemplatePathInfo // information of location of template paths.
	FuncMap    TemplateFuncMap  // same as template.FuncMap.
	LeftDelim  string           // left action delimiter.
	RightDelim string           // right action delimiter.

	m   map[templateKey]*template.Template
	app *Application
}

// Get gets a parsed template.
func (t *Template) Get(appName, layout, name, format string) (*template.Template, error) {
	var templateName string
	if layout != "" {
		templateName = filepath.Join(LayoutDir, layout)
	} else {
		templateName = name
	}
	tmpl, exists := t.m[templateKey{
		appName: appName,
		name:    templateName,
		format:  format,
	}]
	if !exists {
		return nil, fmt.Errorf("kocha: template not found: %s:%s/%s.%s", appName, layout, name, format)
	}
	return tmpl, nil
}

func (t *Template) build(app *Application) (*Template, error) {
	if t == nil {
		t = &Template{}
	}
	t.app = app
	if t.LeftDelim == "" {
		t.LeftDelim = "{{"
	}
	if t.RightDelim == "" {
		t.RightDelim = "}}"
	}
	t, err := t.buildFuncMap()
	if err != nil {
		return nil, err
	}
	t, err = t.buildTemplateMap()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Template) buildFuncMap() (*Template, error) {
	m := TemplateFuncMap{
		"yield":           t.yield,
		"in":              t.in,
		"url":             t.url,
		"nl2br":           t.nl2br,
		"raw":             t.raw,
		"invoke_template": t.invokeTemplate,
		"flash":           t.flash,
		"join":            t.join,
	}
	for name, fn := range t.FuncMap {
		m[name] = fn
	}
	t.FuncMap = m
	return t, nil
}

// buildTemplateMap returns templateMap constructed from templateSet.
func (t *Template) buildTemplateMap() (*Template, error) {
	info := t.PathInfo
	var templatePaths map[string]map[string]map[string]string
	if data := t.app.ResourceSet.Get("_kocha_template_paths"); data != nil {
		if paths, ok := data.(map[string]map[string]map[string]string); ok {
			templatePaths = paths
		}
	}
	if templatePaths == nil {
		templatePaths = map[string]map[string]map[string]string{
			info.Name: make(map[string]map[string]string),
		}
		for _, rootPath := range info.Paths {
			if err := t.collectTemplatePaths(templatePaths[info.Name], rootPath); err != nil {
				return nil, err
			}
		}
		t.app.ResourceSet.Add("_kocha_template_paths", templatePaths)
	}
	t.m = map[templateKey]*template.Template{}
	l := len(t.LeftDelim) + len("$ := .Data") + len(t.RightDelim)
	buf := bytes.NewBuffer(append(append(append(make([]byte, 0, l), t.LeftDelim...), "$ := .Data"...), t.RightDelim...))
	for appName, templates := range templatePaths {
		if err := t.buildAppTemplateSet(buf, l, t.m, appName, templates); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// TemplateFuncMap is an alias of templete.FuncMap.
type TemplateFuncMap template.FuncMap

func (t *Template) collectTemplatePaths(templatePaths map[string]map[string]string, templateDir string) error {
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		baseName, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}
		name, ext := util.SplitExt(strings.TrimSuffix(baseName, util.TemplateSuffix))
		if _, exists := templatePaths[ext]; !exists {
			templatePaths[ext] = make(map[string]string)
		}
		templatePaths[ext][name] = path
		return nil
	})
}

func (t *Template) buildAppTemplateSet(buf *bytes.Buffer, l int, m map[templateKey]*template.Template, appName string, templates map[string]map[string]string) error {
	for ext, templateInfos := range templates {
		tmpl := template.New("")
		for name, path := range templateInfos {
			buf.Truncate(l)
			if data := t.app.ResourceSet.Get(path); data != nil {
				if b, ok := data.([]byte); ok {
					buf.Write(b)
				}
			} else {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				_, err = io.Copy(buf, f)
				f.Close()
				if err != nil {
					return err
				}
				t.app.ResourceSet.Add(path, buf.Bytes())
			}
			if _, err := tmpl.New(name).Delims(t.LeftDelim, t.RightDelim).Funcs(template.FuncMap(t.FuncMap)).Parse(buf.String()); err != nil {
				return err
			}
		}
		for _, t := range tmpl.Templates() {
			m[templateKey{
				appName: appName,
				name:    strings.TrimSuffix(t.Name(), ext),
				format:  ext[1:], // truncate the leading dot.
			}] = t
		}
	}
	return nil
}

func (t *Template) yield(c *Context) (template.HTML, error) {
	tmpl, err := t.Get(t.app.Config.AppName, "", c.Name, c.Format)
	if err != nil {
		return "", err
	}
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	if err := tmpl.Execute(buf, c); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// in is for "in" template function.
func (t *Template) in(a, b interface{}) (bool, error) {
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		if v.IsNil() {
			return false, nil
		}
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Interface() == b {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("valid types are slice, array and string, got `%s'", v.Kind())
	}
	return false, nil
}

// url is for "url" template function.
func (t *Template) url(name string, v ...interface{}) (string, error) {
	return t.app.Router.Reverse(name, v...)
}

// nl2br is for "nl2br" template function.
func (t *Template) nl2br(text string) template.HTML {
	return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
}

// raw is for "raw" template function.
func (t *Template) raw(text string) template.HTML {
	return template.HTML(text)
}

// invokeTemplate is for "invoke_template" template function.
func (t *Template) invokeTemplate(unit Unit, tmplName, defTmplName string, ctx ...*Context) (html template.HTML, err error) {
	var c *Context
	switch len(ctx) {
	case 0: // do nothing.
	case 1:
		c = ctx[0]
	default:
		return "", fmt.Errorf("number of context must be 0 or 1")
	}
	t.app.Invoke(unit, func() {
		if html, err = t.readPartialTemplate(tmplName, c); err != nil {
			// TODO: logging error.
			panic(ErrInvokeDefault)
		}
	}, func() {
		html, err = t.readPartialTemplate(defTmplName, c)
	})
	return html, err
}

// flash is for "flash" template function.
// This is a shorthand for {{.Flash.Get "success"}} in template.
func (t *Template) flash(c *Context, key string) string {
	return c.Flash.Get(key)
}

// join is for "join" template function.
func (t *Template) join(a interface{}, sep string) (string, error) {
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		// do nothing.
	default:
		return "", fmt.Errorf("valid types of first argument are slice or array, got `%s'", v.Kind())
	}
	if v.Len() == 0 {
		return "", nil
	}
	buf := append(make([]byte, 0, v.Len()*2-1), fmt.Sprint(v.Index(0).Interface())...)
	for i := 1; i < v.Len(); i++ {
		buf = append(append(buf, sep...), fmt.Sprint(v.Index(i).Interface())...)
	}
	return string(buf), nil
}

func (t *Template) readPartialTemplate(name string, c *Context) (template.HTML, error) {
	tmpl, err := t.Get(t.app.Config.AppName, "", name, "html")
	if err != nil {
		return "", err
	}
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()
	if err := tmpl.Execute(buf, c); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func errorTemplateName(code int) string {
	return filepath.Join(ErrorTemplateDir, strconv.Itoa(code))
}
