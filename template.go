package kocha

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/naoina/kocha/util"
)

// TemplatePathInfo represents an information of template paths.
type TemplatePathInfo struct {
	Name  string   // name of the application.
	Paths []string // directory paths of the template files.
}

// Template represents the templates information.
type Template struct {
	PathInfo TemplatePathInfo // information of location of template paths.
	FuncMap  TemplateFuncMap  // same as template.FuncMap.

	m   templateMap
	app *Application
}

// Get gets a parsed template.
func (t *Template) Get(appName, layoutName, name, format string) *template.Template {
	return t.m[appName][layoutName][format][util.ToSnakeCase(name)]
}

func (t *Template) Ident(appName, layoutName, name, format string) string {
	return fmt.Sprintf("%s:%s %s.%s", appName, layoutName, util.ToSnakeCase(name), format)
}

func (t *Template) build(app *Application) (*Template, error) {
	t.app = app
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
		"in":              t.in,
		"url":             t.url,
		"nl2br":           t.nl2br,
		"raw":             t.raw,
		"invoke_template": t.invokeTemplate,
		"date":            t.date,
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
	var layoutPaths map[string]map[string]map[string]string
	if data := t.app.ResourceSet.Get("_kocha_template_layout_paths"); data != nil {
		if paths, ok := data.(map[string]map[string]map[string]string); ok {
			layoutPaths = paths
		}
	}
	if layoutPaths == nil {
		layoutPaths = map[string]map[string]map[string]string{
			info.Name: make(map[string]map[string]string),
		}
		for _, rootPath := range info.Paths {
			layoutDir := filepath.Join(rootPath, "layouts")
			if err := t.collectLayoutPaths(layoutPaths[info.Name], layoutDir); err != nil {
				return nil, err
			}
		}
		t.app.ResourceSet.Add("_kocha_template_layout_paths", layoutPaths)
	}
	var templatePaths map[string]map[string]map[string]string
	if data := t.app.ResourceSet.Get("_kocha_template_template_paths"); data != nil {
		if paths, ok := data.(map[string]map[string]map[string]string); ok {
			templatePaths = paths
		}
	}
	if templatePaths == nil {
		templatePaths = map[string]map[string]map[string]string{
			info.Name: make(map[string]map[string]string),
		}
		for _, rootPath := range info.Paths {
			layoutDir := filepath.Join(rootPath, "layouts")
			if err := t.collectTemplatePaths(templatePaths[info.Name], rootPath, layoutDir); err != nil {
				return nil, err
			}
		}
		t.app.ResourceSet.Add("_kocha_template_template_paths", templatePaths)
	}
	templateSet := templateMap{
		info.Name: appTemplateSet{},
	}
	for appName, templates := range templatePaths {
		if err := t.buildSingleAppTemplateSet(templateSet[appName], templates); err != nil {
			return nil, err
		}
	}
	for appName, layouts := range layoutPaths {
		if err := t.buildLayoutAppTemplateSet(templateSet[appName], layouts, templatePaths[appName]); err != nil {
			return nil, err
		}
	}
	t.m = templateSet
	return t, nil
}

// TemplateFuncMap is an alias of templete.FuncMap.
type TemplateFuncMap template.FuncMap

type templateMap map[string]appTemplateSet
type appTemplateSet map[string]map[string]map[string]*template.Template

func (ts appTemplateSet) GoString() string {
	return util.GoString(map[string]map[string]map[string]*template.Template(ts))
}

type layoutTemplateSet map[string]map[string]*template.Template
type fileExtTemplateSet map[string]*template.Template

func (t *Template) collectLayoutPaths(layoutPaths map[string]map[string]string, layoutDir string) error {
	return filepath.Walk(layoutDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		baseName, err := filepath.Rel(layoutDir, path)
		if err != nil {
			return err
		}
		name, ext := util.SplitExt(baseName)
		if _, exists := layoutPaths[name]; !exists {
			layoutPaths[name] = make(map[string]string)
		}
		if layoutPath, exists := layoutPaths[name][ext]; exists {
			return fmt.Errorf("duplicate name of layout file:\n  1. %s\n  2. %s\n", layoutPath, path)
		}
		layoutPaths[name][ext] = path
		return nil
	})
}

func (t *Template) collectTemplatePaths(templatePaths map[string]map[string]string, templateDir, excludeDir string) error {
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path == excludeDir {
				return filepath.SkipDir
			}
			return nil
		}
		baseName, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}
		name, ext := util.SplitExt(baseName)
		if _, exists := templatePaths[ext]; !exists {
			templatePaths[ext] = make(map[string]string)
		}
		if templatePath, exists := templatePaths[ext][name]; exists {
			return fmt.Errorf("duplicate name of template file:\n  1. %s\n  2. %s\n", templatePath, path)
		}
		templatePaths[ext][name] = path
		return nil
	})
}

func (t *Template) buildSingleAppTemplateSet(appTemplateSet appTemplateSet, templates map[string]map[string]string) error {
	layoutTemplateSet := layoutTemplateSet{}
	for ext, templateInfos := range templates {
		layoutTemplateSet[ext] = fileExtTemplateSet{}
		for name, path := range templateInfos {
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
			t := template.Must(template.New(name).Funcs(template.FuncMap(t.FuncMap)).Parse(string(templateBytes)))
			layoutTemplateSet[ext][name] = t
		}
	}
	appTemplateSet[""] = layoutTemplateSet
	return nil
}

func (t *Template) buildLayoutAppTemplateSet(appTemplateSet appTemplateSet, layouts map[string]map[string]string, templates map[string]map[string]string) error {
	for layoutName, layoutInfos := range layouts {
		layoutTemplateSet := layoutTemplateSet{}
		for ext, layoutPath := range layoutInfos {
			layoutTemplateSet[ext] = fileExtTemplateSet{}
			var layoutBytes []byte
			if data := t.app.ResourceSet.Get(fmt.Sprintf("_kocha_template_layout_bytes_%s", layoutPath)); data != nil {
				if b, ok := data.([]byte); ok {
					layoutBytes = b
				}
			}
			if layoutBytes == nil {
				b, err := ioutil.ReadFile(layoutPath)
				if err != nil {
					return err
				}
				layoutBytes = b
				t.app.ResourceSet.Add(fmt.Sprintf("_kocha_template_layout_bytes_%s", layoutPath), b)
			}
			for name, path := range templates[ext] {
				// do not use the layoutTemplate.Clone() in order to retrieve layout as string by `kocha build`
				layout := template.Must(template.New("layout").Funcs(template.FuncMap(t.FuncMap)).Parse(string(layoutBytes)))
				var tmpl *template.Template
				if data := t.app.ResourceSet.Get(fmt.Sprintf("_kocha_%s.%s", path, ext)); data != nil {
					if b, ok := data.([]byte); ok {
						template.Must(layout.New(filepath.Base(path)).Parse(string(b)))
						tmpl = layout
					}
				}
				if tmpl == nil {
					tmpl = template.Must(layout.ParseFiles(path))
				}
				layoutTemplateSet[ext][name] = tmpl
			}
		}
		appTemplateSet[layoutName] = layoutTemplateSet
	}
	return nil
}

// in is for "in" template function.
func (t *Template) in(a, b interface{}) bool {
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		if v.IsNil() {
			return false
		}
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Interface() == b {
				return true
			}
		}
	default:
		panic(fmt.Errorf("invalid type %v: valid types are slice, array and string", v.Type().Name()))
	}
	return false
}

// url is for "url" template function.
func (t *Template) url(name string, v ...interface{}) string {
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
func (t *Template) invokeTemplate(unit Unit, tmplName, defTmplName string, context ...interface{}) (html template.HTML, err error) {
	var ctx interface{}
	switch len(context) {
	case 0: // do nothing.
	case 1:
		ctx = context[0]
	default:
		return "", fmt.Errorf("number of context must be 0 or 1")
	}
	t.app.Invoke(unit, func() {
		if html, err = t.readPartialTemplate(tmplName, ctx); err != nil {
			// TODO: logging error.
			panic(ErrInvokeDefault)
		}
	}, func() {
		html, err = t.readPartialTemplate(defTmplName, ctx)
	})
	return html, err
}

// date is for "date" template function.
func (t *Template) date(date time.Time, layout string) string {
	return date.Format(layout)
}

func (t *Template) readPartialTemplate(name string, ctx interface{}) (template.HTML, error) {
	tmpl := t.Get(t.app.Config.AppName, "", name, "html")
	if tmpl == nil {
		return "", fmt.Errorf("%v: template not found", name)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
