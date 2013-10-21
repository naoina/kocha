package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/naoina/kocha"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"text/template"
)

var Generators = map[string]Generator{
	"controller": &controllerGenerator{},
}

type generateCommand struct {
	flag *flag.FlagSet
}

func (c *generateCommand) Name() string {
	return "generate"
}

func (c *generateCommand) Alias() string {
	return "g"
}

func (c *generateCommand) Short() string {
	return "generate files"
}

func (c *generateCommand) Usage() string {
	var buf bytes.Buffer
	template.Must(template.New("usage").Parse(`%s GENERATOR [args]

Generators:
{{range .}}
    {{.Name|printf "%-6s"}}{{end}}
`)).Execute(&buf, Generators)
	return fmt.Sprintf(buf.String(), c.Name())
}

func (c *generateCommand) DefineFlags(fs *flag.FlagSet) {
	c.flag = fs
}

func (c *generateCommand) Run() {
	generatorName := c.flag.Arg(0)
	if generatorName == "" {
		panicOnError(c, "abort: no GENERATOR given")
	}
	generator, ok := Generators[generatorName]
	if !ok {
		panicOnError(c, "abort: could not find generator: %v", generatorName)
	}
	flagSet := flag.NewFlagSet(generatorName, flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s %s %s\n", os.Args[0], c.Name(), generator.Usage())
	}
	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(Error); ok {
				fmt.Fprintln(os.Stderr, err.message)
				fmt.Fprintf(os.Stderr, "usage: %s %s %s\n", os.Args[0], c.Name(), err.usager.Usage())
				os.Exit(1)
			}
			panic(err)
		}
	}()
	generator.DefineFlags(flagSet)
	flagSet.Parse(c.flag.Args()[1:])
	generator.Generate()
}

type Generator interface {
	Name() string
	Usage() string
	DefineFlags(*flag.FlagSet)
	Generate()
}

type controllerGenerator struct {
	flag *flag.FlagSet
}

var routeTypeName = reflect.TypeOf(kocha.Route{}).String()

func (g *controllerGenerator) Name() string {
	return "controller"
}

func (g *controllerGenerator) Usage() string {
	return fmt.Sprintf("%s NAME", g.Name())
}

func (g *controllerGenerator) DefineFlags(fs *flag.FlagSet) {
	g.flag = fs
}

func (g *controllerGenerator) Generate() {
	name := g.flag.Arg(0)
	if name == "" {
		panicOnError(g, "abort: no NAME given")
	}
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	skeletonDir := filepath.Join(baseDir, "skeleton", "generate")
	camelCaseName := kocha.ToCamelCase(name)
	snakeCaseName := kocha.ToSnakeCase(name)
	data := map[string]interface{}{
		"Name": camelCaseName,
	}
	copyTemplate(g,
		filepath.Join(skeletonDir, "controller.go"),
		filepath.Join("app", "controllers", snakeCaseName+".go"), data)
	copyTemplate(g,
		filepath.Join(skeletonDir, "view.html"),
		filepath.Join("app", "views", snakeCaseName+".html"), data)
	g.addRouteToFile(name)
}

func (g *controllerGenerator) addRouteToFile(name string) {
	routeFilePath := filepath.Join("config", "routes.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, routeFilePath, nil, 0)
	if err != nil {
		panicOnError(g, "abort: failed to read file: %v", err)
	}
	routeStructName := kocha.ToCamelCase(name)
	routeName := kocha.ToSnakeCase(name)
	routeTableAST := findRouteTableAST(f)
	if routeTableAST == nil {
		return
	}
	routeASTs := findRouteASTs(routeTableAST)
	if routeASTs == nil {
		return
	}
	if isRouteDefined(routeASTs, routeStructName) {
		return
	}
	routeFile, err := os.OpenFile(routeFilePath, os.O_RDWR, 0644)
	if err != nil {
		panicOnError(g, "abort: failed to open file: %v", err)
	}
	defer routeFile.Close()
	lastRouteAST := routeASTs[len(routeASTs)-1]
	offset := int64(fset.Position(lastRouteAST.End()).Offset)
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, routeFile, offset); err != nil {
		panicOnError(g, "abort: failed to read file: %v", err)
	}
	buf.WriteString(fmt.Sprintf(`,
&%s{
	Name:       "%s",
	Path:       "/%s",
	Controller: controllers.%s{},
}`, routeTypeName, routeName, routeName, routeStructName))
	if _, err := io.Copy(&buf, routeFile); err != nil {
		panicOnError(g, "abort: failed to read file: %v", err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		panicOnError(g, "abort: failed to format file: %v", err)
	}
	if _, err := routeFile.WriteAt(formatted, 0); err != nil {
		panicOnError(g, "abort: failed to update file: %v", err)
	}
}

func findRouteTableAST(file *ast.File) *ast.CompositeLit {
	var routeTableAST *ast.CompositeLit
	ast.Inspect(file, func(node ast.Node) bool {
		switch aType := node.(type) {
		case *ast.GenDecl:
			if aType.Tok != token.VAR {
				return false
			}
		case *ast.CompositeLit:
			switch t := aType.Type.(type) {
			case *ast.ArrayType:
				star, ok := t.Elt.(*ast.StarExpr)
				if !ok {
					return false
				}
				selector, ok := star.X.(*ast.SelectorExpr)
				if !ok {
					return false
				}
				if name := fmt.Sprintf("%s.%s", selector.X.(*ast.Ident).Name, selector.Sel.Name); name == routeTypeName {
					routeTableAST = aType
					return false
				}
			}
		}
		return true
	})
	return routeTableAST
}

func findRouteASTs(clit *ast.CompositeLit) []*ast.CompositeLit {
	var routeASTs []*ast.CompositeLit
	ast.Inspect(clit, func(node ast.Node) bool {
		switch aType := node.(type) {
		case *ast.CompositeLit:
			switch t := aType.Type.(type) {
			case *ast.SelectorExpr:
				if name := fmt.Sprintf("%s.%s", t.X.(*ast.Ident).Name, t.Sel.Name); name == routeTypeName {
					routeASTs = append(routeASTs, aType)
				}
				return false
			}
		}
		return true
	})
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
			lit, ok := kv.Value.(*ast.CompositeLit)
			if !ok {
				continue
			}
			selector, ok := lit.Type.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			if selector.X.(*ast.Ident).Name == "controllers" && selector.Sel.Name == routeStructName {
				return true
			}
		}
	}
	return false
}
