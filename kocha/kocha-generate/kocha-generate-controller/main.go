package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/naoina/kocha"
	"github.com/naoina/kocha/util"
)

const (
	progName = "kocha generate controller"
)

var (
	routeTableTypeName = reflect.TypeOf(kocha.RouteTable{}).Name()
)

var option struct {
	Help bool `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] NAME

Generate the skeleton files of controller.

Options:
    -h, --help        display this help and exit

`, progName)
}

// generate generates the controller templates.
func generate(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no NAME given")
	}
	name := args[0]
	camelCaseName := util.ToCamelCase(name)
	snakeCaseName := util.ToSnakeCase(name)
	receiverName := strings.ToLower(name)
	if len(receiverName) > 1 {
		receiverName = receiverName[:2]
	} else {
		receiverName = receiverName[:1]
	}
	data := map[string]interface{}{
		"Name":     camelCaseName,
		"Receiver": receiverName,
	}
	if err := util.CopyTemplate(
		filepath.Join(skeletonDir("controller"), "controller.go.template"),
		filepath.Join("app", "controller", snakeCaseName+".go"), data); err != nil {
		return err
	}
	if err := util.CopyTemplate(
		filepath.Join(skeletonDir("controller"), "view.html"),
		filepath.Join("app", "view", snakeCaseName+".html"), data); err != nil {
		return err
	}
	return addRouteToFile(name)
}

func addRouteToFile(name string) error {
	routeFilePath := filepath.Join("config", "routes.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, routeFilePath, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	routeStructName := util.ToCamelCase(name)
	routeName := util.ToSnakeCase(name)
	routeTableAST, err := findRouteTableAST(f)
	if err != nil {
		return err
	}
	if routeTableAST == nil {
		return nil
	}
	routeASTs := findRouteASTs(routeTableAST)
	if routeASTs == nil {
		return nil
	}
	if isRouteDefined(routeASTs, routeStructName) {
		return nil
	}
	routeFile, err := os.OpenFile(routeFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer routeFile.Close()
	lastRouteAST := routeASTs[len(routeASTs)-1]
	offset := int64(fset.Position(lastRouteAST.End()).Offset)
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, routeFile, offset); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	buf.WriteString(fmt.Sprintf(`, {
	Name:       "%s",
	Path:       "/%s",
	Controller: &controller.%s{},
}`, routeName, routeName, routeStructName))
	if _, err := io.Copy(&buf, routeFile); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format file: %v", err)
	}
	if _, err := routeFile.WriteAt(formatted, 0); err != nil {
		return fmt.Errorf("failed to update file: %v", err)
	}
	return nil
}

var ErrRouteTableASTIsFound = errors.New("route table AST is found")

func findRouteTableAST(file *ast.File) (routeTableAST *ast.CompositeLit, err error) {
	defer func() {
		if e := recover(); e != nil && e != ErrRouteTableASTIsFound {
			err = e.(error)
		}
	}()
	ast.Inspect(file, func(node ast.Node) bool {
		switch aType := node.(type) {
		case *ast.GenDecl:
			if aType.Tok != token.VAR {
				return false
			}
			ast.Inspect(aType, func(n ast.Node) bool {
				switch typ := n.(type) {
				case *ast.CompositeLit:
					switch t := typ.Type.(type) {
					case *ast.Ident:
						if t.Name == routeTableTypeName {
							routeTableAST = typ
							panic(ErrRouteTableASTIsFound)
						}
					}
				}
				return true
			})
		}
		return true
	})
	return routeTableAST, nil
}

func findRouteASTs(clit *ast.CompositeLit) []*ast.CompositeLit {
	var routeASTs []*ast.CompositeLit
	for _, c := range clit.Elts {
		if a, ok := c.(*ast.CompositeLit); ok {
			routeASTs = append(routeASTs, a)
		}
	}
	return routeASTs
}

func isRouteDefined(routeASTs []*ast.CompositeLit, routeStructName string) bool {
	for _, a := range routeASTs {
		for _, elt := range a.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if kv.Key.(*ast.Ident).Name != "Controller" {
				continue
			}
			unary, ok := kv.Value.(*ast.UnaryExpr)
			if !ok {
				continue
			}
			lit, ok := unary.X.(*ast.CompositeLit)
			if !ok {
				continue
			}
			selector, ok := lit.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			if selector.X.(*ast.Ident).Name == "controller" && selector.Sel.Name == routeStructName {
				return true
			}
		}
	}
	return false
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
		fmt.Fprintf(os.Stderr, "%s: %v\n", progName, err)
		printUsage()
		os.Exit(1)
	}
}
