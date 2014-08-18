package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/naoina/kocha"
	"github.com/naoina/kocha/util"
)

type buildCommand struct {
	option struct {
		All  bool   `short:"a" long:"all"`
		Tag  string `short:"t" long:"tag"`
		Help bool   `short:"h" long:"help"`
	}
}

func (c *buildCommand) Name() string {
	return "kocha build"
}

func (c *buildCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS]

Build your application.

Options:
    -a, --all         make the true all-in-one binary
    -t, --tag=TAG     specify version tag
    -h, --help        display this help and exit

`, c.Name())
}

func (c *buildCommand) Option() interface{} {
	return &c.option
}

func (c *buildCommand) Run(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	appDir, err := util.FindAppDir()
	if err != nil {
		return err
	}
	appName := filepath.Base(dir)
	configPkg, err := getPackage(path.Join(appDir, "config"))
	if err != nil {
		return fmt.Errorf(`cannot import "%s": %v`, path.Join(appDir, "config"), err)
	}
	var dbImportPath string
	if dbPkg, err := getPackage(path.Join(appDir, "db")); err == nil {
		dbImportPath = dbPkg.ImportPath
	}
	var migrationImportPath string
	if migrationPkg, err := getPackage(path.Join(appDir, "db", "migration")); err == nil {
		migrationImportPath = migrationPkg.ImportPath
	}
	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		return err
	}
	if err := os.Mkdir(tmpDir, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	skeletonDir := filepath.Join(baseDir, "skeleton", "build")
	mainTemplate, err := ioutil.ReadFile(filepath.Join(skeletonDir, "main.go.template"))
	if err != nil {
		return err
	}
	mainFilePath := filepath.ToSlash(filepath.Join(tmpDir, "main.go"))
	builderFilePath := filepath.ToSlash(filepath.Join(tmpDir, "builder.go"))
	file, err := os.Create(builderFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	builderTemplatePath := filepath.ToSlash(filepath.Join(skeletonDir, "builder.go.template"))
	t := template.Must(template.ParseFiles(builderTemplatePath))
	var resources map[string]string
	if c.option.All {
		resources = collectResourcePaths(filepath.Join(dir, kocha.StaticDir))
	}
	tag, err := c.detectVersionTag()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"configImportPath":    configPkg.ImportPath,
		"dbImportPath":        dbImportPath,
		"migrationImportPath": migrationImportPath,
		"mainTemplate":        string(mainTemplate),
		"mainFilePath":        mainFilePath,
		"resources":           resources,
		"version":             tag,
	}
	if err := t.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	execName := appName
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}
	if err := execCmd("go", "run", builderFilePath); err != nil {
		return err
	}
	if err := execCmd("go", "build", "-o", execName, mainFilePath); err != nil {
		return err
	}
	if err := os.RemoveAll(tmpDir); err != nil {
		return err
	}
	if err := util.PrintSettingEnv(); err != nil {
		return err
	}
	fmt.Printf("build all-in-one binary to %v\n", filepath.Join(dir, execName))
	util.PrintGreen("Build successful!\n")
	return nil
}

func getPackage(importPath string) (*build.Package, error) {
	return build.Import(importPath, "", build.FindOnly)
}

func execCmd(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	if msg, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %v\n%v", err, string(msg))
	}
	return nil
}

func collectResourcePaths(root string) map[string]string {
	result := make(map[string]string)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name()[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result[rel] = filepath.ToSlash(path)
		return nil
	})
	return result
}

func (c *buildCommand) detectVersionTag() (string, error) {
	if c.option.Tag != "" {
		return c.option.Tag, nil
	}
	var repo string
	for _, dir := range []string{".git", ".hg"} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			repo = dir
			break
		}
	}
	version := time.Now().Format(time.RFC1123Z)
	switch repo {
	case ".git":
		bin, err := exec.LookPath("git")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: WARNING: git repository found, but `git` command not found. use \"%s\" as version\n", c.Name(), version)
			break
		}
		line, err := exec.Command(bin, "rev-parse", "HEAD").Output()
		if err != nil {
			return "", fmt.Errorf("unexpected error: %v\nplease specify the version using '--tag' option to avoid the this error", err)
		}
		version = strings.TrimSpace(string(line))
	case ".hg":
		bin, err := exec.LookPath("hg")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: WARNING: hg repository found, but `hg` command not found. use \"%s\" as version\n", c.Name(), version)
			break
		}
		line, err := exec.Command(bin, "identify").Output()
		if err != nil {
			return "", fmt.Errorf("unexpected error: %v\nplease specify version using '--tag' option to avoid the this error", err)
		}
		version = strings.TrimSpace(string(line))
	}
	if version == "" {
		// Probably doesn't reach here.
		version = time.Now().Format(time.RFC1123Z)
		fmt.Fprintf(os.Stderr, `%s: WARNING: version is empty, use "%s" as version`, c.Name(), version)
	}
	return version, nil
}

func main() {
	util.RunCommand(&buildCommand{})
}
