package main

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"go/build"

	"github.com/jessevdk/go-flags"
	"github.com/naoina/kocha/util"
)

const (
	progName      = "kocha migrate"
	defaultDBConf = "default"
)

var option struct {
	DBConf string `long:"db"`
	Limit  int    `short:"n" long:"limit"`
	Help   bool   `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] (up|down)

Run the migrations.

Commands:
    up                apply the migrations
    down              rollback the migrations

Options:
    -n, --limit=N     limit for the number of migrations to apply
        --db=NAME     config [default: "default"]
    -h, --help        display this help and exit

`, progName)
}

func run(args []string) error {
	if len(args) < 1 || !isValidDirection(args[0]) {
		return fmt.Errorf("no `up' or `down' specified")
	}
	if option.Limit < 1 {
		if option.Limit == 0 {
			option.Limit = -1
		} else {
			return fmt.Errorf("`limit' must be greater than or equal to 1")
		}
	}
	if option.DBConf == "" {
		option.DBConf = defaultDBConf
	}
	direction := args[0]
	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(tmpDir, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	_, filename, _, _ := runtime.Caller(0)
	skeletonDir := filepath.Join(filepath.Dir(filename), "skeleton", "migrate")
	t := template.Must(template.ParseFiles(filepath.Join(skeletonDir, direction+".go.template")))
	mainFilePath := filepath.ToSlash(filepath.Join(tmpDir, "migrate.go"))
	file, err := os.Create(mainFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	appDir, err := util.FindAppDir()
	if err != nil {
		return err
	}
	dbPkg, err := getPackage(path.Join(appDir, "db"))
	if err != nil {
		return err
	}
	migrationPkg, err := getPackage(path.Join(appDir, "db", "migration"))
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"dbImportPath":        dbPkg.ImportPath,
		"migrationImportPath": migrationPkg.ImportPath,
		"dbconf":              option.DBConf,
		"limit":               option.Limit,
	}
	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	if err := execCmd("go", "run", mainFilePath); err != nil {
		return err
	}
	if err := os.RemoveAll(tmpDir); err != nil {
		return err
	}
	util.PrintGreen("All migrations are successful!\n")
	return nil
}

func isValidDirection(dir string) bool {
	switch dir {
	case "up", "down":
		return true
	}
	return false
}

func getPackage(importPath string) (*build.Package, error) {
	pkg, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return nil, fmt.Errorf(`cannot import "%s": %v`, importPath, err)
	}
	return pkg, nil
}

func execCmd(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}
	return nil
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
	if err := run(args); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintf(os.Stderr, "%s: %v\n", progName, err)
			printUsage()
		}
		os.Exit(1)
	}
}
