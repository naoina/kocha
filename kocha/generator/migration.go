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
	"github.com/naoina/kocha/util"
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
		util.PanicOnError(g, "abort: no NAME given")
	}
	tx := kocha.TxTypeMap[g.txType]
	if tx == nil {
		util.PanicOnError(g, "abort: unsupported transaction type: `%v`", g.txType)
	}
	now := Now().Format("20060102150405")
	data := map[string]interface{}{
		"Name":       util.ToCamelCase(name),
		"TimeStamp":  now,
		"ImportPath": tx.ImportPath(),
		"TxType":     reflect.TypeOf(tx.TransactionType()).String(),
	}
	util.CopyTemplate(g,
		filepath.Join(SkeletonDir("migration"), "migration.go.template"),
		filepath.Join("db", "migration", fmt.Sprintf("%v_%v.go", now, util.ToSnakeCase(name))), data)
	initPath := filepath.Join("db", "migration", "init.go")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		appDir, err := util.FindAppDir()
		if err != nil {
			panic(err)
		}
		util.CopyTemplate(g,
			filepath.Join(SkeletonDir("migration"), "init.go.template"),
			initPath, map[string]interface{}{
				"typeName":     g.txType,
				"tx":           strings.TrimSpace(util.GoString(tx)),
				"dbImportPath": path.Join(appDir, "db"),
			})
	}
}

func init() {
	kocha.RegisterTransactionType("genmai", &kocha.GenmaiTransaction{})
	Register("migration", &migrationGenerator{})
}
