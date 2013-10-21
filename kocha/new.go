package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type newCommand struct {
	flag *flag.FlagSet
}

func (c *newCommand) Name() string {
	return "new"
}

func (c *newCommand) Alias() string {
	return ""
}

func (c *newCommand) Short() string {
	return "create a new application"
}

func (c *newCommand) Usage() string {
	return fmt.Sprintf("%s APP_PATH", c.Name())
}

func (c *newCommand) DefineFlags(fs *flag.FlagSet) {
	c.flag = fs
}

func (c *newCommand) Run() {
	appPath := c.flag.Arg(0)
	if appPath == "" {
		panicOnError(c, "abort: no APP_PATH given")
	}
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	skeletonDir := filepath.Join(baseDir, "skeleton", "new")
	absPath, err := filepath.Abs(appPath)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(filepath.Join(appPath, "config", "app.go")); err == nil {
		panicOnError(c, "abort: Kocha application is already exists")
	}
	data := map[string]interface{}{
		"appName": filepath.Base(absPath),
	}
	filepath.Walk(skeletonDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			return nil
		}
		dstPath := filepath.Join(appPath, strings.TrimPrefix(path, skeletonDir))
		dstDir := filepath.Dir(dstPath)
		dirCreated, err := mkdirAllIfNotExists(dstDir)
		if err != nil {
			panicOnError(c, "abort: failed to create directory: %v", err)
		}
		if dirCreated {
			fmt.Println(green("create directory"), "", dstDir)
		} else {
			fmt.Println(blue("exist"), "", dstDir)
		}
		copyTemplate(c, path, dstPath, data)
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
