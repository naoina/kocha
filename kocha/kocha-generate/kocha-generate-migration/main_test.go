package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func Test_generateMigrationCommand_Run(t *testing.T) {
	// test for no arguments.
	func() {
		c := &generateMigrationCommand{}
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
		c := &generateMigrationCommand{}
		c.option.ORM = "invalid"
		args := []string{"create_table"}
		err := c.Run(args)
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("unsupported ORM: `invalid'")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
		}
	}()

	// test for default TYPE.
	func() {
		tempdir, err := ioutil.TempDir("", "Test_migrationGenerator_Generate")
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
		fixedTime, err := time.Parse("20060102150405", "20140305090617")
		if err != nil {
			t.Fatal(err)
		}
		_time.Now = func() time.Time { return fixedTime }
		defer func() {
			_time.Now = time.Now
		}()
		c := &generateMigrationCommand{}
		c.option.ORM = ""
		args := []string{"test_create_table"}
		err = c.Run(args)
		var actual interface{} = err
		var expect interface{} = nil
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`generate(%#v) => %#v; want %#v`, args, actual, expect)
		}

		outpath := filepath.Join("db", "migration", fmt.Sprintf("%s_test_create_table.go", fixedTime.Format("20060102150405")))
		if _, err := os.Stat(outpath); os.IsNotExist(err) {
			t.Errorf("generate(%#v); %#v is not exists; want exists", args, outpath)
		}

		body, err := ioutil.ReadFile(outpath)
		if err != nil {
			t.Fatal(err)
		}
		actual = string(body)
		expect = `package migration

import "github.com/naoina/genmai"

func (m *Migration) Up_20140305090617_TestCreateTable(tx *genmai.DB) {
	// FIXME: Update database schema and/or insert seed data.
}

func (m *Migration) Down_20140305090617_TestCreateTable(tx *genmai.DB) {
	// FIXME: Revert the change done by Up_20140305090617_TestCreateTable.
}
`
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("%#v => %#v, want %#v", outpath, actual, expect)
		}
	}()
}
