package main

import (
	"flag"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func Test_newCommand(t *testing.T) {
	cmd := &newCommand{}
	for _, v := range [][]interface{}{
		[]interface{}{"Name", cmd.Name(), "new"},
		[]interface{}{"Alias", cmd.Alias(), ""},
		[]interface{}{"Short", cmd.Short(), "create a new application"},
		[]interface{}{"Usage", cmd.Usage(), "new APP_PATH"},
	} {
		name, actual, expected := v[0], v[1], v[2]
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(".%v expect %v, but %v", name, expected, actual)
		}
	}
	if cmd.flag != nil {
		t.Fatalf("Expect nil, but %v", cmd.flag)
	}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{})
	actual, expected := cmd.flag, flags
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_newCommand_Run_with_no_APP_PATH_given(t *testing.T) {
	cmd := &newCommand{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{})
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Expect panic, but not occurred")
		}
	}()
	cmd.Run()
}

func Test_newCommand_Run_with_already_exists(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "Test_newRun_with_already_exists")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempdir)
	appPath := filepath.Base(tempdir)
	dstPath := filepath.Join(tempdir, "src", appPath)
	configDir := filepath.Join(dstPath, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(filepath.Join(configDir, "app.go"), nil, 0644); err != nil {
		panic(err)
	}
	cmd := &newCommand{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{appPath})
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Expect panic, but not occurred")
		}
	}()
	origGOPATH := build.Default.GOPATH
	defer func() {
		build.Default.GOPATH = origGOPATH
	}()
	build.Default.GOPATH = tempdir + string(filepath.ListSeparator) + build.Default.GOPATH
	cmd.Run()
}

func Test_newCommand_Run(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "Test_newRun")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempdir)
	appPath := filepath.Base(tempdir)
	dstPath := filepath.Join(tempdir, "src", appPath)
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = f, f
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()
	cmd := &newCommand{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{appPath})
	origGOPATH := build.Default.GOPATH
	defer func() {
		build.Default.GOPATH = origGOPATH
	}()
	build.Default.GOPATH = tempdir + string(filepath.ListSeparator) + build.Default.GOPATH
	cmd.Run()
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
		filepath.Join("app", "controllers", "root.go"),
		filepath.Join("app", "views", "layouts", "app.html"),
		filepath.Join("app", "views", "root.html"),
		filepath.Join("config", "app.go"),
		filepath.Join("config", "routes.go"),
		filepath.Join("config", "dev", "app.go"),
		filepath.Join("config", "prod", "app.go"),
		filepath.Join("dev.go"),
		filepath.Join("prod.go"),
		filepath.Join("public", "robots.txt"),
	}
	sort.Strings(actuals)
	sort.Strings(expects)
	for i, _ := range actuals {
		actual := actuals[i]
		expected := filepath.Join(dstPath, expects[i])
		if actual != expected {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}
