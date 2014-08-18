package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func Test_migrateCommand_Run(t *testing.T) {
	func() {
		c := &migrateCommand{}
		args := []string{}
		err := c.Run(args)
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("no `up' or `down' specified")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
		}
	}()

	func() {
		c := &migrateCommand{}
		args := []string{"unknown"}
		err := c.Run(args)
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("no `up' or `down' specified")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
		}
	}()

	for _, v := range []struct {
		out   string
		limit int
		dir   string
	}{
		{"call Up: n => -1", 0, "up"},
		{"call Down: n => -1", 0, "down"},
		{"call Up: n => 1", 1, "up"},
		{"call Down: n => 1", 1, "down"},
		{"call Up: n => 111", 111, "up"},
		{"call Down: n => 111", 111, "down"},
	} {
		func() {
			tempDir, err := ioutil.TempDir("", "TestMigrateRun")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)
			appName := "testappname"
			destPath := filepath.Join(tempDir, "src", appName)
			_, filename, _, _ := runtime.Caller(0)
			baseDir := filepath.Dir(filename)
			testdataDir := filepath.Join(baseDir, "testdata")
			if err := copyAll(testdataDir, destPath); err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(destPath); err != nil {
				t.Fatal(err)
			}
			origGOPATH := build.Default.GOPATH
			defer func() {
				build.Default.GOPATH = origGOPATH
				os.Setenv("GOPATH", origGOPATH)
			}()
			build.Default.GOPATH = tempDir + string(filepath.ListSeparator) + build.Default.GOPATH
			os.Setenv("GOPATH", build.Default.GOPATH)
			origStdout := os.Stdout
			defer func() {
				os.Stdout = origStdout
			}()
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			os.Stdout = w
			c := &migrateCommand{}
			c.option.Limit = v.limit
			args := []string{v.dir}
			err = c.Run(args)
			var actual interface{} = err
			var expect interface{} = nil
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
			}
			w.Close()
			out, err := ioutil.ReadAll(r)
			r.Close()
			if err != nil {
				t.Fatal(err)
			}
			actual = string(out)
			expect = v.out
			if !strings.HasPrefix(actual.(string), expect.(string)) {
				t.Errorf(`run(%#v); output => %#v; want %#v`, args, actual, expect)
			}
		}()
	}
}
