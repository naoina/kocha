package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/naoina/kocha"
	"github.com/naoina/kocha/util"
)

const (
	progName   = "kocha generate migration"
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

var option struct {
	ORM  string `short:"o" long:"orm"`
	Help bool   `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] NAME

Generate the skeleton files of migration.

Options:
    -o, --orm=ORM     ORM to be used for a transaction [default: "genmai"]
    -h, --help        display this help and exit

`, progName)
}

// generate generates migration templates.
func generate(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no NAME given")
	}
	name := args[0]
	if option.ORM == "" {
		option.ORM = defaultORM
	}
	orm, exists := ORM[option.ORM]
	if !exists {
		return fmt.Errorf("unsupported ORM: `%v'", option.ORM)
	}
	now := _time.Now().Format("20060102150405")
	data := map[string]interface{}{
		"Name":       util.ToCamelCase(name),
		"TimeStamp":  now,
		"ImportPath": orm.ImportPath(),
		"TxType":     reflect.TypeOf(orm.TransactionType()).String(),
	}
	if err := util.CopyTemplate(
		filepath.Join(skeletonDir("migration"), "migration.go.template"),
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
			filepath.Join(skeletonDir("migration"), "init.go.template"),
			initPath, map[string]interface{}{
				"typeName":     option.ORM,
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
	parser := flags.NewNamedParser(progName, flags.PrintErrors|flags.PassDoubleDash)
	if _, err := parser.AddGroup("", "", &option); err != nil {
		panic(err)
	}
	args, err := parser.Parse()
	if err != nil {
		printUsage()
		os.Exit(1)
	}
	if option.Help {
		printUsage()
		os.Exit(0)
	}
	if err := generate(args); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintf(os.Stderr, "%s: %v\n", progName, err)
			printUsage()
		}
		os.Exit(1)
	}
}
