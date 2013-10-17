package kocha

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type TemplateSet map[string]map[string]*template.Template

func (t TemplateSet) Get(appName, name, format string) *template.Template {
	return t[appName][fmt.Sprintf("%s.%s", toSnakeCase(name), format)]
}

func (t TemplateSet) Ident(appName, name, format string) string {
	return fmt.Sprintf("%s:%s.%s", appName, toSnakeCase(name), format)
}

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
				if filepath.HasPrefix(name, "layouts/") {
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
				log.Panic(err)
			}
		}
	}
	templateSet := make(TemplateSet)
	for layoutAppName, layouts := range layoutPaths {
		templateSet[layoutAppName] = make(map[string]*template.Template)
		for layoutName, layoutPath := range layouts {
			layoutBytes, err := ioutil.ReadFile(layoutPath)
			if err != nil {
				log.Panic(err)
			}
			layoutTemplate := template.Must(template.New("layout").Parse(string(layoutBytes)))
			templateSet[layoutAppName][layoutName] = layoutTemplate
			for templateAppName, templates := range templatePaths {
				templateSet[templateAppName] = make(map[string]*template.Template)
				for templateName, templatePath := range templates {
					layout := template.Must(layoutTemplate.Clone())
					t := template.Must(layout.ParseFiles(templatePath))
					templateSet[templateAppName][templateName] = t
				}
			}
		}
	}
	return templateSet
}
