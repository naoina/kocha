package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/naoina/kocha"
	"github.com/naoina/kocha/util"
)

const (
	defaultORM = "genmai"
)

var (
	ORM = map[string]kocha.Transactioner{
		"genmai": &kocha.GenmaiTransaction{},
	}
)

// for test.
var _time = struct {
	Now func() time.Time
}{
	Now: time.Now,
}

type generateMigrationCommand struct {
	option struct {
		ORM  string `short:"o" long:"orm"`
		Help bool   `short:"h" long:"help"`
	}
}

func (c *generateMigrationCommand) Name() string {
	return "kocha generate migration"
}

func (c *generateMigrationCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS] NAME

Generate the skeleton files of migration.

Options:
    -o, --orm=ORM     ORM to be used for a transaction [default: "genmai"]
    -h, --help        display this help and exit

`, c.Name())
}

func (c *generateMigrationCommand) Option() interface{} {
	return &c.option
}

// Run generates migration templates.
func (c *generateMigrationCommand) Run(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no NAME given")
	}
	name := args[0]
	if c.option.ORM == "" {
		c.option.ORM = defaultORM
	}
	orm, exists := ORM[c.option.ORM]
	if !exists {
		return fmt.Errorf("unsupported ORM: `%v'", c.option.ORM)
	}
	now := _time.Now().Format("20060102150405")
	data := map[string]interface{}{
		"Name":       util.ToCamelCase(name),
		"TimeStamp":  now,
		"ImportPath": orm.ImportPath(),
		"TxType":     reflect.TypeOf(orm.TransactionType()).String(),
	}
	if err := util.CopyTemplate(
		filepath.Join(skeletonDir("migration"), "migration.go"+util.TemplateSuffix),
		filepath.Join("db", "migration", fmt.Sprintf("%v_%v.go", now, util.ToSnakeCase(name))),
		data,
	); err != nil {
		return err
	}
	initPath := filepath.Join("db", "migration", "init.go")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		appDir, err := util.FindAppDir()
		if err != nil {
			return err
		}
		if err := util.CopyTemplate(
			filepath.Join(skeletonDir("migration"), "init.go"+util.TemplateSuffix),
			initPath, map[string]interface{}{
				"typeName":     c.option.ORM,
				"tx":           strings.TrimSpace(util.GoString(orm)),
				"dbImportPath": path.Join(appDir, "db"),
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func skeletonDir(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	return filepath.Join(baseDir, "skeleton", name)
}

func main() {
	util.RunCommand(&generateMigrationCommand{})
}
