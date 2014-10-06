package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_generateModelCommand_Run(t *testing.T) {
	// test for no arguments.
	func() {
		c := &generateModelCommand{}
		args := []string{}
		err := c.Run(args)
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("no NAME given")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
		}
	}()

	// test for unsupported ORM.
	func() {
		c := &generateModelCommand{}
		c.option.ORM = "invalid"
		args := []string{"app_model"}
		err := c.Run(args)
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("unsupported ORM: `invalid'")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
		}
	}()

	// test for invalid argument formats.
	func() {
		c := &generateModelCommand{}
		c.option.ORM = ""
		for _, v := range []struct {
			arg    string
			expect interface{}
		}{
			{"field", fmt.Errorf("invalid argument format is specified: `field'")},
			{":", fmt.Errorf("field name isn't specified: `:'")},
			{"field:", fmt.Errorf("field type isn't specified: `field:'")},
			{":type", fmt.Errorf("field name isn't specified: `:type'")},
			{":string", fmt.Errorf("field name isn't specified: `:string'")},
			{"", fmt.Errorf("invalid argument format is specified: `'")},
		} {
			func() {
				args := []string{"app_model", v.arg}
				err := c.Run(args)
				var actual interface{} = err
				var expect interface{} = v.expect
				if !reflect.DeepEqual(actual, expect) {
					t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
				}
			}()
		}
	}()

	// test for default ORM.
	func() {
		tempdir, err := ioutil.TempDir("", "TestModelGeneratorGenerate")
		if err != nil {
			panic(err)
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
		c := &generateModelCommand{}
		args := []string{"app_model"}
		err = c.Run(args)
		var actual interface{} = err
		var expect interface{} = nil
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
		}

		expect = filepath.Join("app", "model", "app_model.go")
		if _, err := os.Stat(expect.(string)); os.IsNotExist(err) {
			t.Errorf("generate(%#v); file %#v is not exists; want exists", args, expect)
		}
	}()
}
