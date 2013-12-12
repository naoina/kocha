// AUTO-GENERATED BY kocha build
// DO NOT EDIT THIS FILE
package main

import (
	"github.com/naoina/kocha"
	"io/ioutil"
	config "{{.configImportPath}}"
	"os"
	"text/template"
)

const (
	mainTemplate = `{{.mainTemplate}}`
)

func main() {
	funcMap := template.FuncMap{
		"goString": kocha.GoString,
	}
	t := template.Must(template.New("main").Funcs(funcMap).Parse(mainTemplate))
	file, err := os.Create("{{.mainFilePath}}")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	appConfig := config.AppConfig
	res := map[string]string{
		{{range $name, $path := .resources}}
		"{{$name}}": "{{$path}}",
		{{end}}
	}
	resources := make(map[string]string)
	for name, path := range res {
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		resources[name] = kocha.Gzip(string(buf))
	}
	data := map[string]interface{}{
		"appConfig":             appConfig,
		"addr":                  config.Addr,
		"port":                  config.Port,
		"config":                kocha.Config(appConfig.AppName),
		"controllersImportPath": "{{.controllersImportPath}}",
		"resources":             resources,
		"version":               "{{.version}}",
	}
	if err := t.Execute(file, data); err != nil {
		panic(err)
	}
}