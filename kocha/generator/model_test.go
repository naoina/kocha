package generator

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestGenmaiModelType_FieldTypeMap(t *testing.T) {
	m := map[string]ModelFieldType{
		"int":        ModelFieldType{"int", nil},
		"integer":    ModelFieldType{"int", nil},
		"int8":       ModelFieldType{"int8", nil},
		"byte":       ModelFieldType{"int8", nil},
		"int16":      ModelFieldType{"int16", nil},
		"smallint":   ModelFieldType{"int16", nil},
		"int32":      ModelFieldType{"int32", nil},
		"int64":      ModelFieldType{"int64", nil},
		"bigint":     ModelFieldType{"int64", nil},
		"string":     ModelFieldType{"string", nil},
		"text":       ModelFieldType{"string", []string{`size:"65533"`}},
		"mediumtext": ModelFieldType{"string", []string{`size:"16777216"`}},
		"longtext":   ModelFieldType{"string", []string{`size:"4294967295"`}},
		"bytea":      ModelFieldType{"[]byte", nil},
		"blob":       ModelFieldType{"[]byte", nil},
		"mediumblob": ModelFieldType{"[]byte", []string{`size:"65533"`}},
		"longblob":   ModelFieldType{"[]byte", []string{`size:"4294967295"`}},
		"bool":       ModelFieldType{"bool", nil},
		"boolean":    ModelFieldType{"bool", nil},
		"float":      ModelFieldType{"genmai.Float64", nil},
		"float64":    ModelFieldType{"genmai.Float64", nil},
		"double":     ModelFieldType{"genmai.Float64", nil},
		"real":       ModelFieldType{"genmai.Float64", nil},
		"date":       ModelFieldType{"time.Time", nil},
		"time":       ModelFieldType{"time.Time", nil},
		"datetime":   ModelFieldType{"time.Time", nil},
		"timestamp":  ModelFieldType{"time.Time", nil},
		"decimal":    ModelFieldType{"genmai.Rat", nil},
		"numeric":    ModelFieldType{"genmai.Rat", nil},
	}
	actual := (&GenmaiModelType{}).FieldTypeMap()
	expected := m
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiModelType{}).FieldTypeMap() => %q, want %q", "genmai", actual, expected)
	}
}

func TestGenmaiModelType_TemplatePath(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(filename)
	path1, path2 := (&GenmaiModelType{}).TemplatePath()
	actual := path1
	expected := filepath.Join(basepath, "skeleton", "model", "genmai", "genmai.go.template")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiModelType{}).TemplatePath() => %q, $, want %q, $", actual, expected)
	}
	actual = path2
	expected = filepath.Join(basepath, "skeleton", "model", "genmai", "config.go.template")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiModelType{}).TemplatePath() => $, %q, want $, %q", actual, expected)
	}
}

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
		expected := filepath.Join("app", "model", "app_model.go")
		if _, err := os.Stat(expected); os.IsNotExist(err) {
			t.Errorf("%v hasn't been exist", expected)
		}
	}()
}

type testModelType struct{}

func (mt *testModelType) FieldTypeMap() map[string]ModelFieldType {
	return nil
}

func (mt *testModelType) TemplatePath() (templatePath string, configTemplatePath string) {
	return "dummy", "dummy"
}

func TestRegisterModelType(t *testing.T) {
	bakModelTypeMap := modelTypeMap
	defer func() {
		modelTypeMap = bakModelTypeMap
	}()
	mt := &testModelType{}
	RegisterModelType("testtype", mt)
	actual := modelTypeMap["testtype"]
	expected := mt
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("RegisterModelType(%q, %q) => %q, want %q", "testtype", mt, actual, expected)
	}
}
