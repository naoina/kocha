package generator

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/naoina/genmai"
	"github.com/naoina/kocha"
)

const DefaultTxType = "genmai"

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
		return nil, fmt.Errorf("kocha: migration: genmai: unsupported driver type: %v", driverName)
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

// migrationGenerator is the generator of migration.
type migrationGenerator struct {
	flag   *flag.FlagSet
	txType string
}

// Usage returns the usage of the migration generator.
func (g *migrationGenerator) Usage() string {
	return "migration [-o TYPE] NAME"
}

func (g *migrationGenerator) DefineFlags(fs *flag.FlagSet) {
	fs.StringVar(&g.txType, "o", DefaultTxType, fmt.Sprintf("specify Transaction type (default: %v)", DefaultTxType))
	g.flag = fs
}

// Generate generates migration templates.
func (g *migrationGenerator) Generate() {
	name := g.flag.Arg(0)
	if name == "" {
		kocha.PanicOnError(g, "abort: no NAME given")
	}
	tx := TxTypeMap[g.txType]
	if tx == nil {
		kocha.PanicOnError(g, "abort: unsupported transaction type: `%v`", g.txType)
	}
	now := Now().Format("20060102150405")
	data := map[string]interface{}{
		"Name":       kocha.ToCamelCase(name),
		"TimeStamp":  now,
		"ImportPath": tx.ImportPath(),
		"TxType":     reflect.TypeOf(tx.TransactionType()).String(),
	}
	kocha.CopyTemplate(g,
		filepath.Join(SkeletonDir("migration"), "migration.go.template"),
		filepath.Join("db", "migrations", fmt.Sprintf("%v_%v.go", now, kocha.ToSnakeCase(name))), data)
	initPath := filepath.Join("db", "migrations", "init.go")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		kocha.CopyTemplate(g,
			filepath.Join(SkeletonDir("migration"), "init.go.template"),
			initPath, nil)
	}
}

// RegisterTransactionType registers a transaction type.
// If already registered, it overwrites.
func RegisterTransactionType(name string, tx Transactioner) {
	TxTypeMap[name] = tx
}

func init() {
	RegisterTransactionType("genmai", &GenmaiTransaction{})
	Register("migration", &migrationGenerator{})
}
