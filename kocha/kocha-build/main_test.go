package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func Test_buildCommand_Name(t *testing.T) {
	c := &buildCommand{}
	var actual interface{} = c.Name()
	var expect interface{} = "kocha build"
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`%T.Name() => %#v; want %#v`, c, actual, expect)
	}
}

func Test_buildCommand_Run_withNoENVGiven(t *testing.T) {
	c := &buildCommand{}
	args := []string{}
	err := c.Run(args)
	actual := err.Error()
	expect := "cannot import "
	if !strings.HasPrefix(actual, expect) {
		t.Errorf(`%T.Run(%#v) => %#v; want %#v`, c, args, actual, expect)
	}
}

func Test_buildCommand_Run(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "Test_buildCommandRun")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	appName := "testappname"
	dstPath := filepath.Join(tempDir, "src", appName)
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	testdataDir := filepath.Join(baseDir, "testdata")
	if err := copyAll(testdataDir, dstPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dstPath); err != nil {
		t.Fatal(err)
	}
	origGOPATH := build.Default.GOPATH
	defer func() {
		build.Default.GOPATH = origGOPATH
		os.Setenv("GOPATH", origGOPATH)
	}()
	build.Default.GOPATH = tempDir + string(filepath.ListSeparator) + build.Default.GOPATH
	os.Setenv("GOPATH", build.Default.GOPATH)
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
	c := &buildCommand{}
	args := []string{}
	err = c.Run(args)
	var actual interface{} = err
	var expect interface{} = nil
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`%T.Run(%#v) => %#v; want %#v`, c, args, actual, expect)
	}

	tmpDir := filepath.Join(dstPath, "tmp")
	if _, err := os.Stat(tmpDir); err == nil {
		t.Errorf("Expect %v was removed, but exists", tmpDir)
	}

	execName := appName
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}
	execPath := filepath.Join(dstPath, execName)
	if _, err := os.Stat(execPath); err != nil {
		t.Fatalf("Expect %v is exists, but not exists", execName)
	}

	output, err := exec.Command(execPath, "-v").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	actual = string(output)
	expect = fmt.Sprintf("%s version ", execName)
	if !strings.HasPrefix(actual.(string), expect.(string)) {
		t.Errorf("Expect starts with %v, but %v", expect, actual)
	}
}
