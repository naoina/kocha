package kocha

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type (
	methodArgs map[string]string
)

type Route struct {
	Name        string
	Path        string
	Controller  interface{}
	MethodTypes map[string]methodArgs
	RegexpPath  *regexp.Regexp
}

var (
	controllerMethods = map[string]bool{
		"Get":    true,
		"Post":   true,
		"Put":    true,
		"Delete": true,
		"Head":   true,
		"Patch":  true,
	}
	typeRegexpStrings = map[string]string{
		"":    `[\w-]+`, // default
		"int": `\d+`,
	}
	placeHolderRegexp = regexp.MustCompile(`/(?:(?::([\w-]+))|[\w-]*)`)
)

func InitRouteTable(routeTable []*Route) []*Route {
	for _, route := range routeTable {
		route.buildMethodTypes()
		route.buildRegexpPath()
	}
	return routeTable
}

func dispatch(req *http.Request) (controller *reflect.Value, method *reflect.Value, args []reflect.Value) {
	methodName := strings.Title(strings.ToLower(req.Method))
	path := normPath(req.URL.Path)
	for _, route := range appConfig.RouteTable {
		if controller, method, args = route.dispatch(methodName, path); controller != nil {
			break
		}
	}
	return controller, method, args
}

func (route *Route) dispatch(methodName, path string) (controller *reflect.Value, method *reflect.Value, args []reflect.Value) {
	matchesBase := route.RegexpPath.FindStringSubmatch(path)
	if len(matchesBase) == 0 {
		return nil, nil, nil
	}
	matchesBase = matchesBase[1:]
	matches := make([]string, 0, len(matchesBase))
	subexpNames := route.RegexpPath.SubexpNames()[1:]
	types := make([]string, 0, len(subexpNames))
	if args, ok := route.MethodTypes[methodName]; ok {
		for name, t := range args {
			for i, subexpName := range subexpNames {
				if subexpName == name {
					matches = append(matches, matchesBase[i])
					types = append(types, t)
					break
				}
			}
		}
	}
	for i, v := range matches {
		switch types[i] {
		case "int":
			p, err := strconv.Atoi(v)
			if err != nil {
				log.Panic(err)
			}
			args = append(args, reflect.ValueOf(p))
		default:
			args = append(args, reflect.ValueOf(v))
		}
	}
	t := reflect.TypeOf(route.Controller)
	c := reflect.New(t)
	m := c.MethodByName(methodName)
	return &c, &m, args
}

func (route *Route) buildMethodTypes() {
	controller := reflect.TypeOf(route.Controller)
	cname := controller.Name()
	sname := toSnakeCase(cname)
	pkgPath := controller.PkgPath()
	goFile := sname + ".go"
	var filePath string
	for _, src := range build.Default.SrcDirs() {
		if path, err := filepath.Abs(filepath.Join(src, pkgPath, goFile)); err == nil {
			if _, err := os.Stat(path); err == nil {
				filePath = path
				break
			}
		}
	}
	if filePath == "" {
		log.Panic(fmt.Sprintf("%s: no such file", filepath.Join(pkgPath, goFile)))
	}
	f, err := parser.ParseFile(token.NewFileSet(), filePath, nil, 0)
	if err != nil {
		log.Panic(err)
	}
	route.MethodTypes = make(map[string]methodArgs)
	for _, d := range f.Decls {
		ast.Inspect(d, func(node ast.Node) bool {
			fdecl, ok := node.(*ast.FuncDecl)
			if !ok {
				return false
			}
			var recv string
			switch t := fdecl.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				recv = t.X.(*ast.Ident).Name
			case *ast.Ident:
				recv = t.Name
			}
			if recv != cname {
				return false
			}
			methodName := fdecl.Name.Name
			if _, ok := controllerMethods[methodName]; !ok {
				return false
			}
			route.MethodTypes[methodName] = make(methodArgs)
			for _, v := range fdecl.Type.Params.List {
				name := v.Names[0].Name
				t := v.Type.(*ast.Ident).Name
				route.MethodTypes[methodName][name] = t
			}
			return false
		})
	}
}

func (route *Route) buildRegexpPath() {
	var regexpBuf bytes.Buffer
	for _, paths := range placeHolderRegexp.FindAllStringSubmatch(route.Path, -1) {
		name := paths[1]
		if name == "" {
			regexpBuf.WriteString(regexp.QuoteMeta(paths[0]))
			continue
		}
		var rePatStr string
		for _, args := range route.MethodTypes {
			if t, ok := args[name]; ok {
				if rePatStr = typeRegexpStrings[t]; rePatStr == "" {
					rePatStr = typeRegexpStrings[""]
				}
				break
			}
		}
		if rePatStr == "" {
			methodNames := make([]string, 0, len(route.MethodTypes))
			for methodName, _ := range route.MethodTypes {
				methodNames = append(methodNames, methodName)
			}
			controllerName := reflect.TypeOf(route.Controller).Name()
			log.Panicf("argument `%s` is not defined in these methods `%s.%s`",
				name, controllerName, strings.Join(methodNames, ", "))
		}
		regexpBuf.WriteString(fmt.Sprintf(`/(?P<%s>%s)`, regexp.QuoteMeta(name), rePatStr))
	}
	route.RegexpPath = regexp.MustCompile(fmt.Sprintf("^%s$", regexpBuf.String()))
}
