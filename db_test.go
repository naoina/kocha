package kocha

import (
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

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

func TestConst_migrationTableName(t *testing.T) {
	actual := MigrationTableName
	expected := "schema_migration"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("MigrationTableName => %q, want %q", actual, expected)
	}
}

func TestMigrate(t *testing.T) {
	config := DatabaseConfig{Driver: "testdrv", DSN: "testdsn"}
	m := "testm"
	actual := Migrate(config, m)
	expected := &Migration{config: config, m: m}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Migrate(%q, %q) => %q, want %q", config, m, actual, expected)
	}
}

type testMigration struct {
	called []string
}

func (m *testMigration) Up_20140303012000_TestMig1(tx *genmai.DB) {
	m.called = append(m.called, "Up_20140303012000_TestMig1")
}

func (m *testMigration) Down_20140303012000_TestMig1(tx *genmai.DB) {
	m.called = append(m.called, "Down_20140303012000_TestMig1")
}

func (m *testMigration) Up_20140309121357_TestMig2(tx *genmai.DB) {
	m.called = append(m.called, "Up_20140309121357_TestMig2")
}

func (m *testMigration) Down_20140309121357_TestMig2(tx *genmai.DB) {
	m.called = append(m.called, "Down_20140309121357_TestMig2")
}

func TestMigration_Up(t *testing.T) {
	func() {
		tempDir, err := ioutil.TempDir("", "TestMigration_Up")
		if err != nil {
			t.Fatal(err)
		}
		dbpath := filepath.Join(tempDir, "TestMigration_Up.sqlite3")
		defer os.RemoveAll(tempDir)
		origStdout, origStderr := os.Stdout, os.Stderr
		defer func() {
			os.Stdout, os.Stderr = origStdout, origStderr
		}()
		devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout, os.Stderr = devnull, devnull
		TxTypeMap["genmai"] = &GenmaiTransaction{}
		config := DatabaseConfig{Driver: "sqlite3", DSN: dbpath}
		m := &testMigration{}
		if err := (&Migration{config: config, m: m}).Up(-1); err != nil {
			t.Fatal(err)
		}
		actual := m.called
		expected := []string{"Up_20140303012000_TestMig1", "Up_20140309121357_TestMig2"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Up(-1) => %#v, want %#v", config, m, actual, expected)
		}

		m.called = nil
		if err := (&Migration{config: config, m: m}).Up(-1); err != nil {
			t.Fatal(err)
		}
		actual = m.called
		expected = nil
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Up(-1) => %#v, want %#v", config, m, actual, expected)
		}
	}()

	func() {
		tempDir, err := ioutil.TempDir("", "TestMigration_Up")
		if err != nil {
			t.Fatal(err)
		}
		dbpath := filepath.Join(tempDir, "TestMigration_Up.sqlite3")
		defer os.RemoveAll(tempDir)
		origStdout, origStderr := os.Stdout, os.Stderr
		defer func() {
			os.Stdout, os.Stderr = origStdout, origStderr
		}()
		devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout, os.Stderr = devnull, devnull
		TxTypeMap["genmai"] = &GenmaiTransaction{}
		config := DatabaseConfig{Driver: "sqlite3", DSN: dbpath}
		m := &testMigration{}
		if err := (&Migration{config: config, m: m}).Up(1); err != nil {
			t.Fatal(err)
		}
		actual := m.called
		expected := []string{"Up_20140303012000_TestMig1"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Up(1) => %#v, want %#v", config, m, actual, expected)
		}

		m.called = nil
		if err := (&Migration{config: config, m: m}).Up(1); err != nil {
			t.Fatal(err)
		}
		actual = m.called
		expected = []string{"Up_20140309121357_TestMig2"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Up(1) => %#v, want %#v", config, m, actual, expected)
		}

		m.called = nil
		if err := (&Migration{config: config, m: m}).Up(1); err != nil {
			t.Fatal(err)
		}
		actual = m.called
		expected = nil
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Up(1) => %#v, want %#v", config, m, actual, expected)
		}
	}()
}

func TestMigration_Down(t *testing.T) {
	init := func(dbpath string) {
		db, err := sql.Open("sqlite3", dbpath)
		if err != nil {
			t.Fatal(err)
		}
		for _, query := range []string{
			`DROP TABLE IF EXISTS schema_migration`,
			`CREATE TABLE IF NOT EXISTS schema_migration (version varchar(255) PRIMARY KEY)`,
			`INSERT INTO schema_migration VALUES ('20140303012000'), ('20140309121357')`,
		} {
			if _, err := db.Exec(query); err != nil {
				t.Fatal(err)
			}
		}
	}

	func() {
		tempDir, err := ioutil.TempDir("", "TestMigration_Down")
		if err != nil {
			t.Fatal(err)
		}
		dbpath := filepath.Join(tempDir, "TestMigration_Down.sqlite3")
		defer os.RemoveAll(tempDir)
		init(dbpath)
		origStdout, origStderr := os.Stdout, os.Stderr
		defer func() {
			os.Stdout, os.Stderr = origStdout, origStderr
		}()
		devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout, os.Stderr = devnull, devnull
		TxTypeMap["genmai"] = &GenmaiTransaction{}
		config := DatabaseConfig{Driver: "sqlite3", DSN: dbpath}
		m := &testMigration{}
		if err := (&Migration{config: config, m: m}).Down(-1); err != nil {
			t.Fatal(err)
		}
		actual := m.called
		expected := []string{"Down_20140309121357_TestMig2"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Down(-1) => %#v, want %#v", config, m, actual, expected)
		}

		m.called = nil
		if err := (&Migration{config: config, m: m}).Down(-1); err != nil {
			t.Fatal(err)
		}
		actual = m.called
		expected = []string{"Down_20140303012000_TestMig1"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Down(-1) => %#v, want %#v", config, m, actual, expected)
		}
	}()

	func() {
		tempDir, err := ioutil.TempDir("", "TestMigration_Down")
		if err != nil {
			t.Fatal(err)
		}
		dbpath := filepath.Join(tempDir, "TestMigration_Down.sqlite3")
		defer os.RemoveAll(tempDir)
		init(dbpath)
		origStdout, origStderr := os.Stdout, os.Stderr
		defer func() {
			os.Stdout, os.Stderr = origStdout, origStderr
		}()
		devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout, os.Stderr = devnull, devnull
		TxTypeMap["genmai"] = &GenmaiTransaction{}
		config := DatabaseConfig{Driver: "sqlite3", DSN: dbpath}
		m := &testMigration{}
		if err := (&Migration{config: config, m: m}).Down(3); err != nil {
			t.Fatal(err)
		}
		actual := m.called
		expected := []string{"Down_20140309121357_TestMig2", "Down_20140303012000_TestMig1"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Down(3) => %#v, want %#v", config, m, actual, expected)
		}

		m.called = nil
		if err := (&Migration{config: config, m: m}).Down(2); err != nil {
			t.Fatal(err)
		}
		actual = m.called
		expected = nil
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Down(2) => %#v, want %#v", config, m, actual, expected)
		}

		m.called = nil
		if err := (&Migration{config: config, m: m}).Down(1); err != nil {
			t.Fatal(err)
		}
		actual = m.called
		expected = nil
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("(&Migration{config: %#v, m: %#v}).Down(1) => %#v, want %#v", config, m, actual, expected)
		}
	}()
}
