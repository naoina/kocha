package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunWithNoENVGiven(t *testing.T) {
	args := []string{}
	err := run(args)
	actual := err.Error()
	expect := "cannot import "
	if !strings.HasPrefix(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}
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
	if err := copyAll(testdataDir, dstPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dstPath); err != nil {
		panic(err)
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
		panic(err)
	}
	oldStdout, oldStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = f, f
	defer func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
	}()
	args := []string{}
	run(args)
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
	actual := string(output)
	expect := fmt.Sprintf("%s version ", execName)
	if !strings.HasPrefix(actual, expect) {
		t.Errorf("Expect starts with %v, but %v", expect, actual)
	}
}
