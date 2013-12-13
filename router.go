package kocha

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type (
	MethodArgs map[string]string
)

type RouteTable []*Route

// Route represents a route.
type Route struct {
	Name        string
	Path        string
	Controller  interface{}
	MethodTypes map[string]MethodArgs
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
		"":        `[\w-]+`, // default
		"int":     `\d+`,
		"url.URL": `[\w-/.]+`,
	}
	placeHolderRegexp = regexp.MustCompile(`:[\w-]+|\*[\w-/]+`)
	pathRegexp        = regexp.MustCompile(`/(?:(?::([\w-]+))|(?:\*([\w-/]+))|[\w-]*)`)
)

// InitRouteTable returns initialized RouteTable.
//
// Returned RouteTable is always clean so that validate a route.
func InitRouteTable(routeTable RouteTable) RouteTable {
	for _, route := range routeTable {
		route.normalize()
	}
	for _, route := range routeTable {
		if err := route.validateControllerType(); err != nil {
			panic(err)
		}
		route.buildMethodTypes()
	}
	for _, route := range routeTable {
		route.buildRegexpPath()
	}
	for _, route := range routeTable {
		for _, validator := range []func() error{
			route.validateControllerMethodSignature,
			route.validateRouteParameters,
		} {
			if err := validator(); err != nil {
				panic(err)
			}
		}
	}
	return routeTable
}

// Reverse returns path of route by name and any params.
func Reverse(name string, v ...interface{}) string {
	for _, route := range appConfig.RouteTable {
		if route.Name == name {
			return route.reverse(v...)
		}
	}
	types := make([]string, len(v))
	for i, value := range v {
		types[i] = reflect.TypeOf(value).Name()
	}
	panic(fmt.Errorf("no match route found: %v (%v)", name, strings.Join(types, ", ")))
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
		var arg interface{}
		switch types[i] {
		case "int":
			i, err := strconv.Atoi(v)
			if err != nil {
				panic(err)
			}
			arg = i
		case "url.URL":
			u, err := url.Parse(v)
			if err != nil {
				panic(err)
			}
			arg = u
		default:
			arg = v
		}
		args = append(args, reflect.ValueOf(arg))
	}
	t := reflect.TypeOf(route.Controller)
	c := reflect.New(t)
	m := c.MethodByName(methodName)
	return &c, &m, args
}

func (route *Route) reverse(v ...interface{}) string {
	switch n := route.RegexpPath.NumSubexp(); {
	case len(v) < n:
		panic(fmt.Errorf("too few arguments: %v (controller is %v)", route.Name, reflect.TypeOf(route.Controller).Name()))
	case len(v) > n:
		panic(fmt.Errorf("too many arguments: %v (controller is %v)", route.Name, reflect.TypeOf(route.Controller).Name()))
	case len(v)+n == 0:
		return route.Path
	}
	path := placeHolderRegexp.ReplaceAllStringFunc(route.Path, func(s string) string {
		result := fmt.Sprint(v[0])
		v = v[1:]
		return result
	})
	if !route.RegexpPath.MatchString(path) {
		panic(fmt.Errorf("parameter type mismatch: %v (controller is %v)", route.Name, reflect.TypeOf(route.Controller).Name()))
	}
	return normPath(path)
}

func (route *Route) normalize() {
	v := reflect.ValueOf(route.Controller)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.IsValid() {
		route.Controller = v.Interface()
	} else {
		route.Controller = nil
	}
}

func (route *Route) buildMethodTypes() {
	controller := reflect.TypeOf(route.Controller)
	cname := controller.Name()
	pkgPath := controller.PkgPath()
	pkgDir := findPkgDir(pkgPath)
	if pkgDir == "" {
		panic(fmt.Errorf("%v: package not found", pkgPath))
	}
	pkgInfo, err := build.ImportDir(pkgDir, 0)
	if err != nil {
		panic(err)
	}
	astFiles := make([]*ast.File, len(pkgInfo.GoFiles))
	for i, goFilePath := range pkgInfo.GoFiles {
		if astFiles[i], err = parser.ParseFile(token.NewFileSet(), filepath.Join(pkgInfo.Dir, goFilePath), nil, 0); err != nil {
			panic(err)
		}
	}
	route.MethodTypes = make(map[string]MethodArgs)
	for _, file := range astFiles {
		for _, d := range file.Decls {
			ast.Inspect(d, func(node ast.Node) bool {
				fdecl, ok := node.(*ast.FuncDecl)
				if !ok || fdecl.Recv == nil {
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
				route.MethodTypes[methodName] = make(MethodArgs)
				for _, v := range fdecl.Type.Params.List {
					typeName := astTypeName(v.Type)
					for _, name := range v.Names {
						route.MethodTypes[methodName][name.Name] = typeName
					}
				}
				return false
			})
		}
	}
}

func findPkgDir(pkgPath string) string {
	var pkgDir string
	for _, srcDir := range build.Default.SrcDirs() {
		path, err := filepath.Abs(filepath.Join(srcDir, pkgPath))
		if err != nil {
			panic(err)
		}
		if _, err := os.Stat(path); err == nil {
			pkgDir = path
			break
		}
	}
	return pkgDir
}

func astTypeName(expr ast.Expr) string {
	var typeName string
	switch t := expr.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.SelectorExpr:
		typeName = fmt.Sprintf("%v.%v", t.X.(*ast.Ident).Name, t.Sel.Name)
	case *ast.StarExpr:
		typeName = astTypeName(t.X)
	default:
		panic(fmt.Errorf("sorry, unexpected argument type `%T` found. please report this issue.", t))
	}
	return typeName
}

func (route *Route) buildRegexpPath() {
	var regexpBuf bytes.Buffer
	for _, paths := range pathRegexp.FindAllStringSubmatch(route.Path, -1) {
		name := paths[1] + paths[2]
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
			panic(fmt.Errorf("argument `%s` is not defined in these methods `%s.%s`",
				name, controllerName, strings.Join(methodNames, ", ")))
		}
		regexpBuf.WriteString(fmt.Sprintf(`/(?P<%s>%s)`, regexp.QuoteMeta(name), rePatStr))
	}
	route.RegexpPath = regexp.MustCompile(fmt.Sprintf("^%s$", regexpBuf.String()))
}

func (route *Route) validateRouteParameters() error {
	var (
		errors   []string
		dupNames []string
	)
	params := make(map[string]bool)
	for _, paths := range pathRegexp.FindAllStringSubmatch(route.Path, -1) {
		if name := paths[1] + paths[2]; name != "" {
			if _, found := params[name]; found {
				dupNames = append(dupNames, name)
				continue
			}
			params[name] = true
		}
	}
	if length := len(dupNames); length > 0 {
		var format string
		switch {
		case length == 1:
			format = "route parameter `%v` is duplicated in the route '%v'"
		case length > 1:
			format = "route parameters `%v` are duplicated in the route '%v'"
		}
		names := strings.Join(dupNames, "`, `")
		errors = append(errors, fmt.Sprintf(format, names, route.Name))
	}
	for methodName, args := range route.MethodTypes {
		var defNames []string
		for name := range args {
			if !params[name] {
				defNames = append(defNames, name)
			}
		}
		if length := len(defNames); length > 0 {
			var format string
			switch {
			case length == 1:
				format = "argument `%v` is defined in `%v.%v`, but route parameter is not defined"
			case length > 1:
				format = "arguments `%v` are defined in `%v.%v`, but route parameters are not defined"
			}
			controllerName := reflect.TypeOf(route.Controller).Name()
			names := strings.Join(defNames, "`, `")
			errors = append(errors, fmt.Sprintf(format, names, controllerName, methodName))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"+strings.Repeat(" ", len("panic: "))))
	}
	return nil
}

func (route *Route) validateControllerMethodSignature() error {
	var errors []string
	controller := reflect.TypeOf(route.Controller)
	for methodName, _ := range route.MethodTypes {
		meth, found := reflect.PtrTo(controller).MethodByName(methodName)
		if !found {
			return fmt.Errorf("BUG: method `%v` is not found in `%v.%v`", methodName, path.Base(controller.PkgPath()), controller.Name())
		}
		if num := meth.Type.NumOut(); num != 1 {
			errors = append(errors, fmt.Sprintf("by %v.%v.%v, number of return value must be 1, but %v", path.Base(controller.PkgPath()), controller.Name(), meth.Name, num))
			continue
		}
		resultType := reflect.TypeOf((*Result)(nil)).Elem()
		if rtype := meth.Type.Out(0); !rtype.Implements(resultType) {
			errors = append(errors, fmt.Sprintf("by %v.%v.%v, type of return value must be `%v`, but `%v`", path.Base(controller.PkgPath()), controller.Name(), meth.Name, resultType, rtype))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"+strings.Repeat(" ", len("panic: "))))
	}
	return nil
}

func (route *Route) validateControllerType() error {
	c := reflect.ValueOf(route.Controller)
	if c.Kind() != reflect.Struct || !c.FieldByName("Controller").IsValid() {
		return fmt.Errorf(`Controller of route "%s" must be any type of embedded %T or that pointer, but %T`, route.Name, Controller{}, route.Controller)
	}
	switch cc := c.FieldByName("Controller").Interface().(type) {
	case Controller:
	case *Controller:
	default:
		return fmt.Errorf("Controller field must be struct of %T or that pointer, but %T", Controller{}, cc)
	}
	return nil
}
