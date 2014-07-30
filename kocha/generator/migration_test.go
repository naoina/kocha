package generator

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func Test_migrationGenerator(t *testing.T) {
	g := &migrationGenerator{}
	var (
		actual   interface{}
		expected interface{}
	)
	actual = g.Usage()
	expected = "migration [-o TYPE] NAME"
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

func Test_migrationGenerator_Generate(t *testing.T) {
	// test for no arguments.
	func() {
		g := &migrationGenerator{}
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

	// test for unsupported TYPE.
	func() {
		g := &migrationGenerator{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		g.DefineFlags(flags)
		flags.Parse([]string{"-o", "invlaid", "create_table"})
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("it hasn't been panic by unsupported TYPE")
			}
		}()
		g.Generate()
	}()

	// test for default TYPE.
	func() {
		tempdir, err := ioutil.TempDir("", "Test_migrationGenerator_Generate")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(tempdir)
		os.Chdir(tempdir)
		g := &migrationGenerator{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		g.DefineFlags(flags)
		flags.Parse([]string{"test_create_table"})
		f, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		oldStdout, oldStderr := os.Stdout, os.Stderr
		os.Stdout, os.Stderr = f, f
		defer func() {
			os.Stdout, os.Stderr = oldStdout, oldStderr
		}()
		fixedTime, err := time.Parse("20060102150405", "20140305090617")
		if err != nil {
			t.Fatal(err)
		}
		Now = func() time.Time { return fixedTime }
		defer func() {
			Now = time.Now
		}()
		g.Generate()
		outpath := filepath.Join("db", "migration", fmt.Sprintf("%s_test_create_table.go", fixedTime.Format("20060102150405")))
		if _, err := os.Stat(outpath); os.IsNotExist(err) {
			t.Errorf("%v hasn't been exist", outpath)
		}

		body, err := ioutil.ReadFile(outpath)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(body)
		expected := `package migration

import "github.com/naoina/genmai"

func (m *Migration) Up_20140305090617_TestCreateTable(tx *genmai.DB) {
	// FIXME: Update database schema and/or insert seed data.
}

func (m *Migration) Down_20140305090617_TestCreateTable(tx *genmai.DB) {
	// FIXME: Revert the change done by Up_20140305090617_TestCreateTable.
}
`
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%q => %#v, want %#v", outpath, actual, expected)
		}
	}()
}
