package generators

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_controllerGenerator(t *testing.T) {
	g := &controllerGenerator{}
	var (
		actual   interface{}
		expected interface{}
	)
	actual = g.Usage()
	expected = "controller NAME"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if g.flag != nil {
		t.Fatalf("Expect nil, but %v", g.flag)
	}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	g.DefineFlags(flags)
	flags.Parse([]string{})
	actual = g.flag
	expected = flags
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_controllerGeneratorGenerate_with_no_NAME_given(t *testing.T) {
	g := &controllerGenerator{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	g.DefineFlags(flags)
	flags.Parse([]string{})
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Expect panic, but not occurred")
		}
	}()
	g.Generate()
}

func Test_controllerGeneratorGenerate(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "Test_controllerGeneratorGenerate")
	if err != nil {
		panic(err)
	}
	// defer os.RemoveAll(tempdir)
	os.Chdir(tempdir)
	controllerPath := filepath.Join("app", "controllers")
	if err := os.MkdirAll(controllerPath, 0755); err != nil {
		panic(err)
	}
	viewPath := filepath.Join("app", "views")
	if err := os.MkdirAll(viewPath, 0755); err != nil {
		panic(err)
	}
	configPath := filepath.Join("config")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		panic(err)
	}
	routeFileContent := `package config
import (
	"../app/controllers"
	"github/naoina/kocha"
)
var Routes = []*kocha.Route{
	&kocha.Route{
		Name: "root",
		Path: "/",
		Controller: controllers.Root{},
	},
}`
	if err := ioutil.WriteFile(filepath.Join(configPath, "routes.go"), []byte(routeFileContent), 0644); err != nil {
		panic(err)
	}
	g := &controllerGenerator{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	g.DefineFlags(flags)
	flags.Parse([]string{"app_controller"})
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = f, f
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()
	g.Generate()
	expected := filepath.Join(controllerPath, "app_controller.go")
	if _, err := os.Stat(expected); os.IsNotExist(err) {
		t.Errorf("Expect %v file exists, but not exists", expected)
	}
	expected = filepath.Join(viewPath, "app_controller.html")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("Expect %v file exists, but not exists", expected)
	}
	expected = `package config

import (
	"../app/controllers"
	"github/naoina/kocha"
)

var Routes = []*kocha.Route{
	&kocha.Route{
		Name:       "root",
		Path:       "/",
		Controller: controllers.Root{},
	},
	&kocha.Route{
		Name:       "app_controller",
		Path:       "/app_controller",
		Controller: controllers.AppController{},
	},
}
`
	routesBuf, err := ioutil.ReadFile(filepath.Join(configPath, "routes.go"))
	if err != nil {
		panic(err)
	}
	actual := string(routesBuf)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	// test that duplicated routes are not added
	g.Generate()
	routesBuf, err = ioutil.ReadFile(filepath.Join(configPath, "routes.go"))
	if err != nil {
		panic(err)
	}
	actual = string(routesBuf)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_routeTypeName(t *testing.T) {
	actual := routeTypeName
	expected := "kocha.Route"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
