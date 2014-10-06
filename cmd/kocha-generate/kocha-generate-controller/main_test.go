package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_generateControllerCommand_Run_withNoNAMEGiven(t *testing.T) {
	c := &generateControllerCommand{}
	args := []string{}
	err := c.Run(args)
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("no NAME given")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
	}
}

func Test_generateControllerCommand_Run(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "Test_controllerGeneratorGenerate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)
	os.Chdir(tempdir)
	controllerPath := filepath.Join("app", "controller")
	if err := os.MkdirAll(controllerPath, 0755); err != nil {
		t.Fatal(err)
	}
	viewPath := filepath.Join("app", "view")
	if err := os.MkdirAll(viewPath, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join("config")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatal(err)
	}
	routeFileContent := `package config
import (
	"../app/controller"
	"github/naoina/kocha"
)
type RouteTable kocha.RouteTable
var routes = RouteTable{
	{
		Name: "root",
		Path: "/",
		Controller: &controller.Root{},
	},
}
func Routes() RouteTable {
	return append(routes, RouteTable{}...)
}
`
	if err := ioutil.WriteFile(filepath.Join(configPath, "routes.go"), []byte(routeFileContent), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = f, f
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()
	c := &generateControllerCommand{}
	args := []string{"app_controller"}
	err = c.Run(args)
	var actual interface{} = err
	var expect interface{} = nil
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
	}

	for _, v := range []string{
		filepath.Join(controllerPath, "app_controller.go"),
		filepath.Join(viewPath, "app_controller.html"),
	} {
		if _, err := os.Stat(v); os.IsNotExist(err) {
			t.Errorf("generate(%#v); file %#v is not exists; want exists", args, v)
		}
	}

	content := `package config

import (
	"../app/controller"
	"github/naoina/kocha"
)

type RouteTable kocha.RouteTable

var routes = RouteTable{
	{
		Name:       "root",
		Path:       "/",
		Controller: &controller.Root{},
	}, {
		Name:       "app_controller",
		Path:       "/app_controller",
		Controller: &controller.AppController{},
	},
}

func Routes() RouteTable {
	return append(routes, RouteTable{}...)
}
`
	fname := filepath.Join(configPath, "routes.go")
	routesBuf, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(routesBuf)
	expect = content
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`generate(%#v); %#v => %#v; want %#v`, args, fname, actual, expect)
	}

	// test that duplicated routes are not added
	actual = c.Run(args)
	expect = nil
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
	}

	routesBuf, err = ioutil.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(routesBuf)
	expect = content
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`generate(%#v); %#v => %#v; want %#v`, args, fname, actual, expect)
	}
}

func TestRouteTableTypeName(t *testing.T) {
	actual := routeTableTypeName
	expect := "RouteTable"
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`routeTableTypeName => %#v; want %#v`, actual, expect)
	}
}
