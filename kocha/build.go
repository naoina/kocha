package main

import (
	"flag"
	"fmt"
	"github.com/naoina/kocha"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"text/template"
)

type buildCommand struct {
	flag *flag.FlagSet
}

func (c *buildCommand) Name() string {
	return "build"
}

func (c *buildCommand) Alias() string {
	return "b"
}

func (c *buildCommand) Short() string {
	return "build your application"
}

func (c *buildCommand) Usage() string {
	return fmt.Sprintf("%s ENV", c.Name())
}

func (c *buildCommand) DefineFlags(fs *flag.FlagSet) {
	c.flag = fs
}

func (c *buildCommand) Run() {
	env := c.flag.Arg(0)
	if env == "" {
		kocha.PanicOnError(c, "abort: no ENV given")
	}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	appName := filepath.Base(dir)
	configPkg := c.Package(path.Join(appName, "config", env))
	controllersPkg := c.Package(path.Join(appName, "app", "controllers"))
	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		panic(err)
	}
	if err := os.Mkdir(tmpDir, 0755); err != nil && !os.IsExist(err) {
		kocha.PanicOnError(c, "abort: failed to create directory: %v", err)
	}
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	skeletonDir := filepath.Join(baseDir, "skeleton", "build")
	mainTemplate, err := ioutil.ReadFile(filepath.Join(skeletonDir, "main.go"))
	if err != nil {
		panic(err)
	}
	mainFilePath := filepath.Join(tmpDir, "main.go")
	builderFilePath := filepath.Join(tmpDir, "builder.go")
	file, err := os.Create(builderFilePath)
	if err != nil {
		kocha.PanicOnError(c, "abort: failed to create file: %v", err)
	}
	defer file.Close()
	builderTemplatePath := filepath.Join(skeletonDir, "builder.go")
	t := template.Must(template.ParseFiles(builderTemplatePath))
	data := map[string]string{
		"configImportPath":      configPkg.ImportPath,
		"controllersImportPath": controllersPkg.ImportPath,
		"mainTemplate":          string(mainTemplate),
		"mainFilePath":          mainFilePath,
	}
	if err := t.Execute(file, data); err != nil {
		kocha.PanicOnError(c, "abort: failed to write file: %v", err)
	}
	c.execCmd("go", "run", builderFilePath)
	c.execCmd("go", "build", "-o", appName, mainFilePath)
	fmt.Printf("build all-in-one binary to %v\n", filepath.Join(dir, appName))
	fmt.Println(kocha.Green("Build successful"))
	if err := os.RemoveAll(tmpDir); err != nil {
		panic(err)
	}
}

func (c *buildCommand) Package(importPath string) *build.Package {
	pkg, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		kocha.PanicOnError(c, "abort: cannot import `%s`: %v", importPath, err)
	}
	return pkg
}

func (c *buildCommand) execCmd(cmd string, args ...string) {
	command := exec.Command(cmd, args...)
	if msg, err := command.CombinedOutput(); err != nil {
		kocha.PanicOnError(c, "abort: build failed: %v\n%v", err, string(msg))
	}
}
