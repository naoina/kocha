package generator

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_modelGenerator(t *testing.T) {
	g := &modelGenerator{}
	var (
		actual   interface{}
		expected interface{}
	)
	actual = g.Usage()
	expected = "model [-o ORM] NAME [[field:type] [field:type]...]"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Usage() => %q, want %q", actual, expected)
	}
	if g.flag != nil {
		t.Errorf("flag => %q, want nil", g.flag)
	}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	g.DefineFlags(flags)
	flags.Parse([]string{})
	actual = g.flag
	expected = flags
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("flag => %q, want %q", actual, expected)
	}
}

func Test_modelGenerator_Generate(t *testing.T) {
	// test for no arguments.
	func() {
		g := &modelGenerator{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		g.DefineFlags(flags)
		flags.Parse([]string{})
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("it hasn't been panic by empty arguments")
			}
		}()
		g.Generate()
	}()

	// test for unsupported ORM.
	func() {
		g := &modelGenerator{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		g.DefineFlags(flags)
		flags.Parse([]string{"-o", "invlaid", "app_model"})
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("it hasn't been panic by unsupported ORM")
			}
		}()
		g.Generate()
	}()

	// test for invalid argument formats.
	func() {
		for _, v := range []string{
			"field", ":", "field:", ":type", ":string", "",
		} {
			func() {
				g := &modelGenerator{}
				flags := flag.NewFlagSet("testflags", flag.ExitOnError)
				g.DefineFlags(flags)
				flags.Parse([]string{"app_model", v})
				defer func() {
					if err := recover(); err == nil {
						t.Errorf("it hasn't been panic by invalid argument format: %q", v)
					}
				}()
				g.Generate()
			}()
		}
	}()

	// test for default ORM.
	func() {
		tempdir, err := ioutil.TempDir("", "Test_modelGenerator_Generate")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(tempdir)
		os.Chdir(tempdir)
		g := &modelGenerator{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		g.DefineFlags(flags)
		flags.Parse([]string{"app_model"})
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
		expected := filepath.Join("app", "models", "app_model.go")
		if _, err := os.Stat(expected); os.IsNotExist(err) {
			t.Errorf("%v hasn't been exist", expected)
		}
	}()

	// test helper.
	testFieldTypes := func(ORM string, types map[string]fieldType) {
		actual := typeMap[ORM]
		expected := types
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("typeMap[%q] => %q, want %q", ORM, actual, expected)
		}

		for v := range types {
			tempdir, err := ioutil.TempDir("", "Test_modelGenerator_Generate")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempdir)
			os.Chdir(tempdir)
			g := &modelGenerator{}
			flags := flag.NewFlagSet("testflags", flag.ExitOnError)
			g.DefineFlags(flags)
			flags.Parse([]string{"app_model", "fieldname:" + v})
			f, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
			if err != nil {
				panic(err)
			}
			oldStdout, oldStderr := os.Stdout, os.Stderr
			os.Stdout, os.Stderr = f, f
			defer func() {
				os.Stdout, os.Stderr = oldStdout, oldStderr
			}()
			defer func() {
				if err := recover(); err != nil {
					t.Errorf("testcase typeMap[%q][%q]: %v", ORM, v, err)
				}
			}()
			g.Generate()
			expected := filepath.Join("app", "models", "app_model.go")
			if _, err := os.Stat(expected); os.IsNotExist(err) {
				t.Errorf("%v hasn't been exist", expected)
			}
		}
	}

	// test for default ORM with valid field types.
	testFieldTypes(defaultORM, map[string]fieldType{
		"int":        fieldType{"int", nil},
		"integer":    fieldType{"int", nil},
		"int8":       fieldType{"int8", nil},
		"byte":       fieldType{"int8", nil},
		"int16":      fieldType{"int16", nil},
		"smallint":   fieldType{"int16", nil},
		"int32":      fieldType{"int32", nil},
		"int64":      fieldType{"int64", nil},
		"bigint":     fieldType{"int64", nil},
		"string":     fieldType{"string", nil},
		"text":       fieldType{"string", []string{`size:"65533"`}},
		"mediumtext": fieldType{"string", []string{`size:"16777216"`}},
		"longtext":   fieldType{"string", []string{`size:"4294967295"`}},
		"bytea":      fieldType{"[]byte", nil},
		"blob":       fieldType{"[]byte", nil},
		"mediumblob": fieldType{"[]byte", []string{`size:"65533"`}},
		"longblob":   fieldType{"[]byte", []string{`size:"4294967295"`}},
		"bool":       fieldType{"bool", nil},
		"boolean":    fieldType{"bool", nil},
		"float":      fieldType{"float64", nil},
		"float64":    fieldType{"float64", nil},
		"double":     fieldType{"float64", nil},
		"real":       fieldType{"float64", nil},
		"date":       fieldType{"time.Time", nil},
		"time":       fieldType{"time.Time", nil},
		"datetime":   fieldType{"time.Time", nil},
		"timestamp":  fieldType{"time.Time", nil},
		"decimal":    fieldType{"genmai.Rat", nil},
		"numeric":    fieldType{"genmai.Rat", nil},
	})
}
