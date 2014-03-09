package generator

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/naoina/kocha"
)

const DefaultTxType = "genmai"

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
	tx := kocha.TxTypeMap[g.txType]
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
		appDir, err := kocha.FindAppDir()
		if err != nil {
			panic(err)
		}
		kocha.CopyTemplate(g,
			filepath.Join(SkeletonDir("migration"), "init.go.template"),
			initPath, map[string]interface{}{
				"typeName":     g.txType,
				"tx":           strings.TrimSpace(kocha.GoString(tx)),
				"dbImportPath": path.Join(appDir, "db"),
			})
	}
}

func init() {
	kocha.RegisterTransactionType("genmai", &kocha.GenmaiTransaction{})
	Register("migration", &migrationGenerator{})
}
