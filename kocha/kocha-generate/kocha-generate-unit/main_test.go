package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGenerate(t *testing.T) {
	func() {
		args := []string{}
		err := generate(args)
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("no NAME given")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
		}
	}()

	func() {
		tempdir, err := ioutil.TempDir("", "TestGenerateUnit")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempdir)
		os.Chdir(tempdir)
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
		args := []string{"app_unit"}
		err = generate(args)
		var actual interface{} = err
		var expect interface{} = nil
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
		}

		expect = filepath.Join("app", "unit", "app_unit.go")
		if _, err := os.Stat(expect.(string)); os.IsNotExist(err) {
			t.Errorf("generate(%#v); file %#v is not exists; want exists", args, expect)
		}
	}()
}
