package kocha

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var (
	TemplateFuncs = template.FuncMap{
		"eq": func(a, b interface{}) bool {
			// TODO: remove in Go 1.2
			//       see http://tip.golang.org/pkg/text/template/#hdr-Functions
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			// TODO: remove in Go 1.2
			//       see http://tip.golang.org/pkg/text/template/#hdr-Functions
			return a != b
		},
		"in": func(a, b interface{}) bool {
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
		},
		"url": Reverse,
		"nl2br": func(text string) template.HTML {
			return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
		},
		"raw": func(text string) template.HTML {
			return template.HTML(text)
		},
		"date": func(date time.Time, layout string) string {
			return date.Format(layout)
		},
	}
)

type TemplateSet map[string]map[string]*template.Template

// Get gets a parsed template.
func (t TemplateSet) Get(appName, name, format string) *template.Template {
	return t[appName][fmt.Sprintf("%s.%s", ToSnakeCase(name), format)]
}

func (t TemplateSet) Ident(appName, name, format string) string {
	return fmt.Sprintf("%s:%s.%s", appName, ToSnakeCase(name), format)
}

// TemplateSetFromPaths returns TemplateSet constructed from templateSetPaths.
func TemplateSetFromPaths(templateSetPaths map[string][]string) TemplateSet {
	layoutPaths := make(map[string]map[string]string)
	templatePaths := make(map[string]map[string]string)
	for appName, paths := range templateSetPaths {
		layoutPaths[appName] = make(map[string]string)
		templatePaths[appName] = make(map[string]string)
		for _, rootPath := range paths {
			if err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				name, err := filepath.Rel(rootPath, path)
				if err != nil {
					return err
				}
				if filepath.HasPrefix(filepath.ToSlash(name), "layouts/") {
					if layoutPath, ok := layoutPaths[appName][name]; ok {
						return fmt.Errorf("duplicate name of layout file:\n  1. %s\n  2. %s\n", layoutPath, path)
					}
					layoutPaths[appName][name] = path
				} else {
					if templatePath, ok := templatePaths[appName][name]; ok {
						return fmt.Errorf("duplicate name of template file:\n  1. %s\n  2. %s\n", templatePath, path)
					}
					templatePaths[appName][name] = path
				}
				return nil
			}); err != nil {
				panic(err)
			}
		}
	}
	templateSet := make(TemplateSet)
	for layoutAppName, layouts := range layoutPaths {
		templateSet[layoutAppName] = make(map[string]*template.Template)
		for layoutName, layoutPath := range layouts {
			layoutBytes, err := ioutil.ReadFile(layoutPath)
			if err != nil {
				panic(err)
			}
			layoutTemplate := template.Must(template.New("layout").Funcs(TemplateFuncs).Parse(string(layoutBytes)))
			templateSet[layoutAppName][layoutName] = layoutTemplate
			for templateAppName, templates := range templatePaths {
				templateSet[templateAppName] = make(map[string]*template.Template)
				for templateName, templatePath := range templates {
					// do not use the layoutTemplate.Clone() in order to retrieve layout as string by `kocha build`
					layout := template.Must(template.New("layout").Funcs(TemplateFuncs).Parse(string(layoutBytes)))
					t := template.Must(layout.ParseFiles(templatePath))
					templateSet[templateAppName][templateName] = t
				}
			}
		}
	}
	return templateSet
}
