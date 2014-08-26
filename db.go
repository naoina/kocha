package kocha

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"sort"

	"github.com/naoina/genmai"
)

type DatabaseMap map[string]DatabaseConfig

// DatabaseConfig represents a configuration of the database.
type DatabaseConfig struct {
	// name of database driver such as "mysql".
	Driver string

	// Data Source Name.
	// e.g. such as "travis@/db_name".
	DSN string
}

var TxTypeMap = make(map[string]Transactioner)

// Transactioner is an interface for a transaction type.
type Transactioner interface {
	// ImportPath returns the import path to import to use a transaction.
	// Usually, import path of ORM like "github.com/naoina/genmai".
	ImportPath() string

	// TransactionType returns an object of a transaction type.
	// Return value will only used to determine an argument type for the
	// methods of migration when generate the template.
	TransactionType() interface{}

	// Begin starts a transaction from driver name and data source name.
	Begin(driverName, dsn string) (tx interface{}, err error)

	// Commit commits the transaction.
	Commit() error

	// Rollback rollbacks the transaction.
	Rollback() error
}

// RegisterTransactionType registers a transaction type.
// If already registered, it overwrites.
func RegisterTransactionType(name string, tx Transactioner) {
	TxTypeMap[name] = tx
}

// GenmaiTransaction implements Transactioner interface.
type GenmaiTransaction struct {
	tx *genmai.DB
}

// ImportPath returns the import path of Genmai.
func (t *GenmaiTransaction) ImportPath() string {
	return "github.com/naoina/genmai"
}

// TransactionType returns the transaction type of Genmai.
func (t *GenmaiTransaction) TransactionType() interface{} {
	return &genmai.DB{}
}

// Begin starts a transaction of Genmai.
// If unsupported driver name given or any error occurred, it returns nil and error.
func (t *GenmaiTransaction) Begin(driverName, dsn string) (tx interface{}, err error) {
	var d genmai.Dialect
	switch driverName {
	case "mysql":
		d = &genmai.MySQLDialect{}
	case "postgres":
		d = &genmai.PostgresDialect{}
	case "sqlite3":
		d = &genmai.SQLite3Dialect{}
	default:
		return nil, fmt.Errorf("kocha: migration: genmai: unsupported driver type `%v'", driverName)
	}
	t.tx, err = genmai.New(d, dsn)
	if err != nil {
		return nil, err
	}
	if err := t.tx.Begin(); err != nil {
		return nil, err
	}
	return t.tx, nil
}

// Commit commits the transaction of Genmai.
func (t *GenmaiTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rollbacks the transaction of Genmai.
func (t *GenmaiTransaction) Rollback() error {
	return t.tx.Rollback()
}

var (
	upMethodRegexp   = regexp.MustCompile(`\AUp_(\d{14})_\w+\z`)
	downMethodRegexp = regexp.MustCompile(`\ADown_(\d{14})_\w+\z`)
)

const MigrationTableName = "schema_migration"

type Migration struct {
	config DatabaseConfig
	m      interface{}
}

func Migrate(config DatabaseConfig, m interface{}) *Migration {
	return &Migration{
		config: config,
		m:      m,
	}
}

func (mig *Migration) Up(limit int) error {
	if err := mig.transaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (version varchar(255) PRIMARY KEY)`,
			MigrationTableName))
		if err != nil {
			return err
		}
		defer stmt.Close()
		if _, err := stmt.Exec(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	var version string
	if err := mig.transaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(fmt.Sprintf(`SELECT version FROM %s ORDER BY version DESC LIMIT 1`, MigrationTableName))
		if err != nil {
			return err
		}
		defer stmt.Close()
		if err := stmt.QueryRow().Scan(&version); err != nil && err != sql.ErrNoRows {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	minfos, err := mig.collectInfos(upMethodRegexp, func(p string) bool {
		return p > version
	})
	if err != nil {
		return err
	}
	limit = int(math.Min(float64(limit), float64(len(minfos))))
	if limit < 0 {
		limit = len(minfos)
	}
	if len(minfos[:limit]) < 1 {
		fmt.Fprintf(os.Stderr, "kocha: migrate: there is no need to migrate.\n")
		return nil
	}
	sort.Sort(migrationInfoSlice(minfos))
	return mig.run("migrating", minfos[:limit], func(version string) {
		if err := mig.transaction(func(tx *sql.Tx) error {
			stmt, err := tx.Prepare(fmt.Sprintf(`INSERT INTO %s (version) VALUES (?)`, MigrationTableName))
			if err != nil {
				return err
			}
			defer stmt.Close()
			if _, err := stmt.Exec(version); err != nil {
				return err
			}
			return nil
		}); err != nil {
			panic(err)
		}
	})
}

func (mig *Migration) Down(limit int) error {
	var positions []string
	if err := mig.transaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(fmt.Sprintf(`SELECT version FROM %s ORDER BY version DESC LIMIT ?`, MigrationTableName))
		if err != nil {
			return err
		}
		defer stmt.Close()
		if limit < 1 {
			limit = 1
		}
		rows, err := stmt.Query(limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var version string
			if err := rows.Scan(&version); err != nil {
				return err
			}
			positions = append(positions, version)
		}
		return rows.Err()
	}); err != nil {
		return err
	}
	minfos, err := mig.collectInfos(downMethodRegexp, func(p string) bool {
		for _, version := range positions {
			if p == version {
				return true
			}
		}
		return false
	})
	if err != nil {
		return err
	}
	if len(minfos) < 1 {
		fmt.Fprintf(os.Stderr, "kocha: migrate: there is no need to migrate.\n")
		return nil
	}
	sort.Sort(sort.Reverse(migrationInfoSlice(minfos)))
	return mig.run("rollback", minfos, func(version string) {
		if err := mig.transaction(func(tx *sql.Tx) error {
			stmt, err := tx.Prepare(fmt.Sprintf(`DELETE FROM %s WHERE version = ?`, MigrationTableName))
			if err != nil {
				return err
			}
			defer stmt.Close()
			if _, err := stmt.Exec(version); err != nil {
				return err
			}
			return nil
		}); err != nil {
			panic(err)
		}
	})
}

func (mig *Migration) transaction(f func(tx *sql.Tx) error) error {
	db, err := sql.Open(mig.config.Driver, mig.config.DSN)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
		tx.Commit()
	}()
	if err := f(tx); err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

func (mig *Migration) run(msg string, minfos []migrationInfo, afterFunc func(version string)) error {
	v := reflect.ValueOf(mig.m)
	for _, mi := range minfos {
		func(mi migrationInfo) {
			tx, err := mi.tx.Begin(mig.config.Driver, mig.config.DSN)
			if err != nil {
				panic(err)
			}
			defer func() {
				if err := recover(); err != nil {
					mi.tx.Rollback()
					panic(err)
				}
				mi.tx.Commit()
			}()
			fmt.Printf("%v by %v...\n", msg, mi.methodName)
			meth := v.MethodByName(mi.methodName)
			meth.Call([]reflect.Value{reflect.ValueOf(tx)})
		}(mi)
		afterFunc(mi.version)
	}
	return nil
}

func (mig *Migration) collectInfos(r *regexp.Regexp, isTarget func(string) bool) ([]migrationInfo, error) {
	v := reflect.ValueOf(mig.m)
	t := v.Type()
	var minfos []migrationInfo
	for i := 0; i < t.NumMethod(); i++ {
		meth := t.Method(i)
		name := meth.Name
		matches := r.FindStringSubmatch(name)
		if matches == nil || !isTarget(matches[1]) {
			continue
		}
		if meth.Type.NumIn() != 2 {
			return nil, fmt.Errorf("kocha: migrate: %v: arguments number must be 1", meth.Name)
		}
		argType := meth.Type.In(1)
		tx := mig.findTransactioner(argType)
		if tx == nil {
			return nil, fmt.Errorf("kocha: migrate: argument type `%v' is undefined", argType)
		}
		minfos = append(minfos, migrationInfo{
			methodName: name,
			version:    matches[1],
			tx:         tx,
		})
	}
	return minfos, nil
}

func (mig *Migration) findTransactioner(t reflect.Type) Transactioner {
	for _, tx := range TxTypeMap {
		if t == reflect.TypeOf(tx.TransactionType()) {
			return tx
		}
	}
	return nil
}

// migrationInfo is an intermediate information of a migration.
type migrationInfo struct {
	methodName string
	version    string
	tx         Transactioner
}

// migrationInfoSlice implements sort.Interface interface.
type migrationInfoSlice []migrationInfo

// Len implements sort.Interface.Len.
func (ms migrationInfoSlice) Len() int {
	return len(ms)
}

// Less implements sort.Interface.Less.
func (ms migrationInfoSlice) Less(i, j int) bool {
	return ms[i].version < ms[j].version
}

// Swap implements sort.Interface.Swap.
func (ms migrationInfoSlice) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}
