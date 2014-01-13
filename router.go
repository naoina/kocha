package kocha

import (
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

	"github.com/naoina/kocha-urlrouter"
	_ "github.com/naoina/kocha-urlrouter/doublearray"
)

var (
	router            urlrouter.URLRouter
	reverseRouter     ReverseRouter
	controllerMethods = map[string]bool{
		"Get":    true,
		"Post":   true,
		"Put":    true,
		"Delete": true,
		"Head":   true,
		"Patch":  true,
	}
	typeRegexpStrings = map[string]string{
		"string":  `[\w-]+`,
		"int":     `\d+`,
		"url.URL": `[\w-/.]+`,
	}
	typeRegexp map[string]*regexp.Regexp
)

type RouteTable []*Route

func (rt RouteTable) buildRouter() urlrouter.URLRouter {
	records := make([]*urlrouter.Record, len(rt))
	for i, route := range rt {
		records[i] = urlrouter.NewRecord(route.Path, route)
	}
	router := urlrouter.NewURLRouter("doublearray")
	if err := router.Build(records); err != nil {
		panic(err)
	}
	return router
}

func (rt RouteTable) buildReverseRouter() ReverseRouter {
	reverse := make(map[string]*routeInfo)
	for _, route := range rt {
		_, params := router.Lookup(route.Path)
		names := make([]string, len(params))
		values := make([]string, len(params))
		for i, param := range params {
			names[i], values[i] = param.Name, param.Value
		}
		reverse[route.Name] = &routeInfo{
			route:      route,
			params:     values,
			paramNames: names,
		}
	}
	return reverse
}

func (rt RouteTable) GoString() string {
	return fmt.Sprintf("kocha.InitRouteTable(%s)", GoString([]*Route(rt)))
}

// Route represents a route.
type Route struct {
	Name        string
	Path        string
	Controller  interface{}
	MethodTypes map[string]MethodArgs
}

type MethodArgs map[string]string

type ReverseRouter map[string]*routeInfo

type routeInfo struct {
	route      *Route
	paramNames []string
	params     []string
}

// Reverse returns path of route by name and any params.
func Reverse(name string, v ...interface{}) string {
	info := reverseRouter[name]
	if info == nil {
		types := make([]string, len(v))
		for i, value := range v {
			types[i] = reflect.TypeOf(value).Name()
		}
		panic(fmt.Errorf("no match route found: %v (%v)", name, strings.Join(types, ", ")))
	}
	return info.reverse(v...)
}

func (ri *routeInfo) reverse(v ...interface{}) string {
	route := ri.route
	switch vlen, nlen := len(v), len(ri.params); {
	case vlen < nlen:
		panic(fmt.Errorf("too few arguments: %v (controller is %v)", route.Name, reflect.TypeOf(route.Controller).Name()))
	case vlen > nlen:
		panic(fmt.Errorf("too many arguments: %v (controller is %v)", route.Name, reflect.TypeOf(route.Controller).Name()))
	case vlen+nlen == 0:
		return route.Path
	}
	var arg MethodArgs
	for _, arg = range route.MethodTypes {
		break
	}
	for i := 0; i < len(v); i++ {
		t := arg[ri.paramNames[i]]
		re := typeRegexp[t]
		if re == nil {
			panic(fmt.Errorf("regexp for type `%v` is not defined", t))
		}
		if !re.MatchString(fmt.Sprint(v[i])) {
			panic(fmt.Errorf("parameter type mismatch: %v (controller is %v)", route.Name, reflect.TypeOf(route.Controller).Name()))
		}
	}
	var oldnew []string
	for i := 0; i < len(v); i++ {
		oldnew = append(oldnew, ri.params[i], fmt.Sprint(v[i]))
	}
	replacer := strings.NewReplacer(oldnew...)
	path := replacer.Replace(route.Path)
	return normPath(path)
}

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
	router = routeTable.buildRouter()
	reverseRouter = routeTable.buildReverseRouter()
	typeRegexp = make(map[string]*regexp.Regexp)
	for t, s := range typeRegexpStrings {
		typeRegexp[t] = regexp.MustCompile(fmt.Sprintf(`\A%s\z`, s))
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

func dispatch(req *http.Request) (controller *reflect.Value, method *reflect.Value, args []reflect.Value) {
	methodName := strings.Title(strings.ToLower(req.Method))
	path := normPath(req.URL.Path)
	data, params := router.Lookup(path)
	if data == nil {
		return nil, nil, nil
	}
	route := data.(*Route)
	return route.dispatch(methodName, params)
}

func (route *Route) dispatch(methodName string, params []urlrouter.Param) (controller *reflect.Value, method *reflect.Value, args []reflect.Value) {
	methodArgs := route.MethodTypes[methodName]
	if methodArgs == nil {
		return nil, nil, nil
	}
	for _, param := range params {
		var arg interface{}
		switch methodArgs[param.Name] {
		case "int":
			i, err := strconv.Atoi(param.Value)
			if err != nil {
				panic(err)
			}
			arg = i
		case "url.URL":
			u, err := url.Parse(param.Value)
			if err != nil {
				panic(err)
			}
			arg = u
		default:
			arg = param.Value
		}
		args = append(args, reflect.ValueOf(arg))
	}
	t := reflect.TypeOf(route.Controller)
	c := reflect.New(t)
	m := c.MethodByName(methodName)
	return &c, &m, args
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

func (route *Route) validateRouteParameters() error {
	var errors []string
	_, params := router.Lookup(route.Path)
	paramNames := make(map[string]bool)
	for _, param := range params {
		paramNames[param.Name] = true
	}
	for methodName, args := range route.MethodTypes {
		var defNames []string
		for name := range args {
			if !paramNames[name] {
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
