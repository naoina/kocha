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

	_ "github.com/mattn/go-sqlite3"
	"github.com/naoina/genmai"
)

func TestGenmaiTransaction_ImportPath(t *testing.T) {
	actual := (&GenmaiTransaction{}).ImportPath()
	expected := "github.com/naoina/genmai"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiTransaction{}).ImportPath() => %q, want %q", actual, expected)
	}
}

func TestGenmaiTransaction_TransactionType(t *testing.T) {
	actual := (&GenmaiTransaction{}).TransactionType()
	expected := &genmai.DB{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiTransaction{}).TransactionType() => %q, want %q", actual, expected)
	}
}

func TestGenmaiTransaction_Begin(t *testing.T) {
	// test for sqlite3.
	func() {
		tx, err := (&GenmaiTransaction{}).Begin("sqlite3", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		actual := reflect.TypeOf(tx)
		expected := reflect.TypeOf(&genmai.DB{})
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&GenmaiTransaction{}).Begin(%q, %q) => %q, nil, want %q, nil", "sqlite3", ":memory:", actual, expected)
		}
	}()

	// test for unsupported driver name.
	func() {
		_, err := (&GenmaiTransaction{}).Begin("unknown", ":memory:")
		if err == nil {
			t.Errorf("(&GenmaiTransaction{}).Begin(%q, %q) => nil, error, want nil, error(%q)", "invalid", ":memory:", err)
		}
	}()

	// TODO: test for mysql and postgres.
}

func TestGenmaiTransaction_Commit(t *testing.T) {
	t.Skipf("it cannot be tested because cannot applies mock to genmai.DB.Commit.")
}

func TestGenmaiTransaction_Rollback(t *testing.T) {
	t.Skipf("it cannot be tested because cannot applies mock to genmai.DB.Rollback.")
}

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
		outpath := filepath.Join("db", "migrations", fmt.Sprintf("%s_test_create_table.go", fixedTime.Format("20060102150405")))
		if _, err := os.Stat(outpath); os.IsNotExist(err) {
			t.Errorf("%v hasn't been exist", outpath)
		}

		body, err := ioutil.ReadFile(outpath)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(body)
		expected := `package migrations

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

type testTransaction struct{}

func (t *testTransaction) ImportPath() string {
	return "path/to/test/transaction"
}

func (t *testTransaction) TransactionType() interface{} {
	return t
}

func (t *testTransaction) Begin(driverName, dsn string) (tx interface{}, err error) {
	return nil, nil
}

func (t *testTransaction) Commit() error {
	return nil
}

func (t *testTransaction) Rollback() error {
	return nil
}

func TestRegisterTransactionType(t *testing.T) {
	bakTxTypeMap := TxTypeMap
	defer func() {
		TxTypeMap = bakTxTypeMap
	}()
	tx := &testTransaction{}
	RegisterTransactionType("testtx", tx)
	actual := TxTypeMap["testtx"]
	expected := tx
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("RegisterTransactionType(%q, %q) => %q, want %q", "testtx", tx, actual, expected)
	}
}
