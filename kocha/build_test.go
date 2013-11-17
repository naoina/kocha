package main

import (
	"flag"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func Test_buildCommand(t *testing.T) {
	cmd := &buildCommand{}
	for _, v := range [][]interface{}{
		{"Name", cmd.Name(), "build"},
		{"Alias", cmd.Alias(), "b"},
		{"Short", cmd.Short(), "build your application"},
		{"Usage", cmd.Usage(), "build [options] ENV"},
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

func Test_buildCommandRun_with_no_ENV_given(t *testing.T) {
	cmd := &buildCommand{}
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

func Test_buildCommandRun(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "Test_buildCommandRun")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	appName := "testappname"
	dstPath := filepath.Join(tempDir, "src", appName)
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	testdataDir := filepath.Join(baseDir, "testdata")
	filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstPath, strings.TrimPrefix(path, testdataDir))
		if info.IsDir() {
			err := os.MkdirAll(filepath.Join(dstPath), 0755)
			return err
		}
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(dstPath, src, 0644); err != nil {
			return err
		}
		return nil
	})
	if err := os.Chdir(dstPath); err != nil {
		panic(err)
	}
	cmd := &buildCommand{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{"dev"})
	origGOPATH := build.Default.GOPATH
	defer func() {
		build.Default.GOPATH = origGOPATH
		os.Setenv("GOPATH", origGOPATH)
	}()
	build.Default.GOPATH = tempDir + string(filepath.ListSeparator) + build.Default.GOPATH
	os.Setenv("GOPATH", build.Default.GOPATH)
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = f, f
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()
	cmd.Run()
	if _, err := os.Stat(filepath.Join(dstPath, appName)); err != nil {
		t.Errorf("Expect %v is exists, but not exists", appName)
	}
	tmpDir := filepath.Join(dstPath, "tmp")
	if _, err := os.Stat(tmpDir); err == nil {
		t.Errorf("Expect %v was removed, but exists", tmpDir)
	}
}
