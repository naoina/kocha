package kocha

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
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

// Template represents the templates information.
type Template struct {
	PathInfo   TemplatePathInfo // information of location of template paths.
	FuncMap    TemplateFuncMap  // same as template.FuncMap.
	LeftDelim  string           // left action delimiter.
	RightDelim string           // right action delimiter.

	m   templateMap
	app *Application
}

// Get gets a parsed template.
func (t *Template) Get(layout, name, format string) (*template.Template, error) {
	var r *template.Template
	tmpl := t.m[t.app.Config.AppName][format]
	if tmpl == nil {
		goto ErrNotFound
	}
	if layout != "" {
		r = tmpl.Lookup(filepath.Join(LayoutDir, layout) + "." + format)
	} else {
		r = tmpl.Lookup(name + "." + format)
	}
	if r == nil {
		goto ErrNotFound
	}
	return r, nil
ErrNotFound:
	return nil, fmt.Errorf("kocha: template not found: %s:%s/%s.%s", t.app.Config.AppName, layout, name, format)
}

func (t *Template) build(app *Application) (*Template, error) {
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
	templateSet := templateMap{
		info.Name: make(map[string]*template.Template),
	}
	for appName, templates := range templatePaths {
		if err := t.buildAppTemplateSet(templateSet[appName], templates); err != nil {
			return nil, err
		}
	}
	t.m = templateSet
	return t, nil
}

// TemplateFuncMap is an alias of templete.FuncMap.
type TemplateFuncMap template.FuncMap

type templateMap map[string]map[string]*template.Template

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

func (t *Template) buildAppTemplateSet(appTemplateSet map[string]*template.Template, templates map[string]map[string]string) error {
	for ext, templateInfos := range templates {
		tmpl := template.New("")
		for _, path := range templateInfos {
			var templateBytes []byte
			if data := t.app.ResourceSet.Get(fmt.Sprintf("_kocha_%s.%s", path, ext)); data != nil {
				if b, ok := data.([]byte); ok {
					templateBytes = b
				}
			}
			if templateBytes == nil {
				b, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				templateBytes = b
				t.app.ResourceSet.Add(fmt.Sprintf("_kocha_%s.%s", path, ext), b)
			}
			name := strings.TrimSuffix(t.relativePath(path), util.TemplateSuffix)
			content := fmt.Sprint(
				t.LeftDelim, "$ := .Data", t.RightDelim,
				string(templateBytes),
			)
			if _, err := tmpl.New(name).Delims(t.LeftDelim, t.RightDelim).Funcs(template.FuncMap(t.FuncMap)).Parse(content); err != nil {
				return err
			}
		}
		appTemplateSet[ext] = tmpl
	}
	return nil
}

func (t *Template) relativePath(targpath string) string {
	for _, basepath := range t.PathInfo.Paths {
		if p, err := filepath.Rel(basepath, targpath); err == nil {
			return p
		}
	}
	return targpath
}

func (t *Template) yield(c *Context) (template.HTML, error) {
	tmpl, err := t.Get("", c.Name, c.Format)
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
	tmpl, err := t.Get("", name, "html")
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
