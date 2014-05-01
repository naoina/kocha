package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/naoina/kocha"
	"github.com/naoina/kocha/util"
)

// newCommand implements `command` interface for `new` command.
type newCommand struct {
	flag *flag.FlagSet
}

// Name returns name of `new` command.
func (c *newCommand) Name() string {
	return "new"
}

// Alias returns alias of `new` command.
func (c *newCommand) Alias() string {
	return ""
}

// Short returns short description for help.
func (c *newCommand) Short() string {
	return "create a new application"
}

// Usage returns usage of `new` command.
func (c *newCommand) Usage() string {
	return fmt.Sprintf("%s APP_PATH", c.Name())
}

func (c *newCommand) DefineFlags(fs *flag.FlagSet) {
	c.flag = fs
}

// Run execute the process for `new` command.
func (c *newCommand) Run() {
	appPath := c.flag.Arg(0)
	if appPath == "" {
		util.PanicOnError(c, "abort: no APP_PATH given")
	}
	dstBasePath := filepath.Join(filepath.SplitList(build.Default.GOPATH)[0], "src", appPath)
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	skeletonDir := filepath.Join(baseDir, "skeleton", "new")
	if _, err := os.Stat(filepath.Join(dstBasePath, "config", "app.go")); err == nil {
		util.PanicOnError(c, "abort: Kocha application is already exists")
	}
	data := map[string]interface{}{
		"appName":   filepath.Base(appPath),
		"appPath":   appPath,
		"secretKey": fmt.Sprintf("%q", string(kocha.GenerateRandomKey(32))), // AES-256
		"signedKey": fmt.Sprintf("%q", string(kocha.GenerateRandomKey(16))),
	}
	filepath.Walk(skeletonDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}
		dstPath := filepath.Join(dstBasePath, strings.TrimSuffix(strings.TrimPrefix(path, skeletonDir), ".template"))
		dstDir := filepath.Dir(dstPath)
		dirCreated, err := mkdirAllIfNotExists(dstDir)
		if err != nil {
			util.PanicOnError(c, "abort: failed to create directory: %v", err)
		}
		if dirCreated {
			util.PrintCreateDirectory(dstDir)
		}
		util.CopyTemplate(c, path, dstPath, data)
		return nil
	})
}

func mkdirAllIfNotExists(dstDir string) (created bool, err error) {
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
