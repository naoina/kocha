package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/naoina/kocha/util"
)

type newCommand struct {
	option struct {
		Help bool `short:"h" long:"help"`
	}
}

func (c *newCommand) Name() string {
	return "kocha new"
}

func (c *newCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS] APP_PATH

Create a new application.

Options:
    -h, --help        display this help and exit

`, c.Name())
}

func (c *newCommand) Option() interface{} {
	return &c.option
}

func (c *newCommand) Run(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no APP_PATH given")
	}
	appPath := args[0]
	dstBasePath := filepath.Join(filepath.SplitList(build.Default.GOPATH)[0], "src", appPath)
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	skeletonDir := filepath.Join(baseDir, "skeleton", "new")
	if _, err := os.Stat(filepath.Join(dstBasePath, "config", "app.go")); err == nil {
		return fmt.Errorf("Kocha application is already exists")
	}
	data := map[string]interface{}{
		"appName":   filepath.Base(appPath),
		"appPath":   appPath,
		"secretKey": fmt.Sprintf("%q", string(util.GenerateRandomKey(32))), // AES-256
		"signedKey": fmt.Sprintf("%q", string(util.GenerateRandomKey(16))),
	}
	return filepath.Walk(skeletonDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dstPath := filepath.Join(dstBasePath, strings.TrimSuffix(strings.TrimPrefix(path, skeletonDir), ".template"))
		dstDir := filepath.Dir(dstPath)
		dirCreated, err := mkdirAllIfNotExists(dstDir)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		if dirCreated {
			util.PrintCreateDirectory(dstDir)
		}
		return util.CopyTemplate(path, dstPath, data)
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

func main() {
	util.RunCommand(&newCommand{})
}
