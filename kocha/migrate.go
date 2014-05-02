package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"go/build"
	"github.com/naoina/kocha/util"
)

type migrateCommand struct {
	flag   *flag.FlagSet
	dbconf string
	limit  int
}

func (c *migrateCommand) Name() string {
	return "migrate"
}

func (c *migrateCommand) Alias() string {
	return ""
}

func (c *migrateCommand) Short() string {
	return "run the migrations"
}

func (c *migrateCommand) Usage() string {
	return fmt.Sprintf("%s [-db confname] [-n n] {up|down}", c.Name())
}

func (c *migrateCommand) DefineFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.dbconf, "db", "default", `specify a database config (default: "default")`)
	fs.IntVar(&c.limit, "n", -1, `number of migrations to be run`)
	c.flag = fs
}

func (c *migrateCommand) Run() {
	direction := c.flag.Arg(0)
	switch direction {
	case "up", "down":
		// do nothing.
	default:
		util.PanicOnError(c, `abort: no "up" or "down" specified`)
	}
	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(tmpDir, 0755); err != nil && !os.IsExist(err) {
		util.PanicOnError(c, "abort: failed to create directory: %v", err)
	}
	_, filename, _, _ := runtime.Caller(0)
	skeletonDir := filepath.Join(filepath.Dir(filename), "skeleton", "migrate")
	t := template.Must(template.ParseFiles(filepath.Join(skeletonDir, direction+".go.template")))
	mainFilePath := filepath.ToSlash(filepath.Join(tmpDir, "migrate.go"))
	file, err := os.Create(mainFilePath)
	if err != nil {
		util.PanicOnError(c, "abort: failed to create file: %v", err)
	}
	defer file.Close()
	appDir, err := util.FindAppDir()
	if err != nil {
		panic(err)
	}
	data := map[string]interface{}{
		"dbImportPath":         c.Package(path.Join(appDir, "db")).ImportPath,
		"migrationsImportPath": c.Package(path.Join(appDir, "db", "migrations")).ImportPath,
		"dbconf":               c.dbconf,
		"limit":                c.limit,
	}
	if err := t.Execute(file, data); err != nil {
		util.PanicOnError(c, "abort: failed to write file: %v", err)
	}
	c.execCmd("go", "run", mainFilePath)
	if err := os.RemoveAll(tmpDir); err != nil {
		panic(err)
	}
	util.PrintGreen("All migrations are successful!\n")
}

func (c *migrateCommand) Package(importPath string) *build.Package {
	pkg, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		util.PanicOnError(c, "abort: cannot import `%s`: %v", importPath, err)
	}
	return pkg
}

func (c *migrateCommand) execCmd(cmd string, args ...string) {
	command := exec.Command(cmd, args...)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		util.PanicOnError(c, "abort: migration failed: %v", err)
	}
}
