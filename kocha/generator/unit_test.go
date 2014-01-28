package generator

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_unitGenerator(t *testing.T) {
	g := &unitGenerator{}
	var (
		actual   interface{}
		expected interface{}
	)
	actual = g.Usage()
	expected = "unit NAME"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if g.flag != nil {
		t.Fatalf("Expect nil, but %v", g.flag)
	}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	g.DefineFlags(flags)
	flags.Parse([]string{})
	actual = g.flag
	expected = flags
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_unitGenerator_Generate(t *testing.T) {
	func() {
		g := &unitGenerator{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		g.DefineFlags(flags)
		flags.Parse([]string{})
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("Expect panic, but not occurred")
			}
		}()
		g.Generate()
	}()

	func() {
		tempdir, err := ioutil.TempDir("", "Test_unitGenerator_Generate")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(tempdir)
		os.Chdir(tempdir)
		g := &unitGenerator{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		g.DefineFlags(flags)
		flags.Parse([]string{"app_unit"})
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
		expected := filepath.Join("app", "units", "app_unit.go")
		if _, err := os.Stat(expected); os.IsNotExist(err) {
			t.Errorf("Expect %v file exists, but not exists", expected)
		}
	}()
}
