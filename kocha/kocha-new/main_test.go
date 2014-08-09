package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestRunWithNoAPPPATHGiven(t *testing.T) {
	args := []string{}
	err := run(args)
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("no APP_PATH given")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}
}

func TestRunWithAlreadyExists(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "TestRunWithAlreadyExists")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)
	appPath := filepath.Base(tempdir)
	dstPath := filepath.Join(tempdir, "src", appPath)
	configDir := filepath.Join(dstPath, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(configDir, "app.go"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	origGOPATH := build.Default.GOPATH
	defer func() {
		build.Default.GOPATH = origGOPATH
	}()
	build.Default.GOPATH = tempdir + string(filepath.ListSeparator) + build.Default.GOPATH
	args := []string{appPath}
	err = run(args)
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("Kocha application is already exists")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}
}

func TestRun(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "Test_newRun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)
	appPath := filepath.Base(tempdir)
	dstPath := filepath.Join(tempdir, "src", appPath)
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
	origGOPATH := build.Default.GOPATH
	defer func() {
		build.Default.GOPATH = origGOPATH
	}()
	build.Default.GOPATH = tempdir + string(filepath.ListSeparator) + build.Default.GOPATH
	args := []string{appPath}
	err = run(args)
	var actual interface{} = err
	var expect interface{} = nil
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}

	var actuals []string
	filepath.Walk(dstPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}
		actuals = append(actuals, path)
		return nil
	})
	expects := []string{
		filepath.Join("main.go"),
		filepath.Join("app", "controller", "root.go"),
		filepath.Join("app", "view", "layout", "app.html"),
		filepath.Join("app", "view", "root.html"),
		filepath.Join("config", "app.go"),
		filepath.Join("config", "routes.go"),
		filepath.Join("public", "robots.txt"),
	}
	sort.Strings(actuals)
	sort.Strings(expects)
	for i, _ := range actuals {
		actual := actuals[i]
		expected := filepath.Join(dstPath, expects[i])
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`run(%#v); generated file => %#v; want %#v`, args, actual, expected)
		}
	}
}
