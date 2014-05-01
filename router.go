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
	"github.com/naoina/kocha/util"
)

var (
	controllerMethods = map[string]bool{
		"Get":    true,
		"Post":   true,
		"Put":    true,
		"Delete": true,
		"Head":   true,
		"Patch":  true,
	}
	typeValidateParsers = map[string]TypeValidateParser{
		"string":   &StringTypeValidateParser{regexp.MustCompile(`\A[\w-]+\z`)},
		"int":      &IntTypeValidateParser{regexp.MustCompile(`\A\d+\z`)},
		"*url.URL": &URLTypeValidateParser{regexp.MustCompile(`\A[\w-/.]+\z`)},
	}
)

type RouteTable []*Route

func (rt RouteTable) buildRouter() (*Router, error) {
	for _, route := range rt {
		route.normalize()
	}
	for _, route := range rt {
		if err := route.validateControllerType(); err != nil {
			return nil, err
		}
		if route.MethodTypes == nil {
			if err := route.buildMethodTypes(); err != nil {
				return nil, err
			}
		}
	}
	router, err := newRouter(rt)
	if err != nil {
		return nil, err
	}
	for _, route := range rt {
		info := router.reverse[route.Name]
		route.paramNames = info.paramNames
	}
	for _, route := range rt {
		for _, validator := range []func() error{
			route.validateTypeValidateParser,
			route.validateControllerMethodSignature,
			route.validateRouteParameters,
			route.validateControllerArgumentParameters,
		} {
			if err := validator(); err != nil {
				return nil, err
			}
		}
	}
	return router, nil
}

// Router represents a router of kocha.
type Router struct {
	forward    forwardRouter
	reverse    reverseRouter
	routeTable RouteTable
}

// newRouter returns a new Router.
func newRouter(rt RouteTable) (*Router, error) {
	router := &Router{routeTable: rt}
	if err := router.buildForward(); err != nil {
		return nil, err
	}
	if err := router.buildReverse(); err != nil {
		return nil, err
	}
	return router, nil
}

func (router *Router) dispatch(req *http.Request) (controller *reflect.Value, method *reflect.Value, args []reflect.Value) {
	methodName := strings.Title(strings.ToLower(req.Method))
	path := util.NormPath(req.URL.Path)
	data, params := router.forward.Lookup(path)
	if data == nil {
		return nil, nil, nil
	}
	route := data.(*Route)
	return route.dispatch(methodName, params)
}

// buildForward builds forward router.
func (router *Router) buildForward() error {
	records := make([]urlrouter.Record, len(router.routeTable))
	for i, route := range router.routeTable {
		records[i] = urlrouter.NewRecord(route.Path, route)
	}
	router.forward = urlrouter.NewURLRouter("doublearray")
	if err := router.forward.Build(records); err != nil {
		return err
	}
	return nil
}

// buildReverse builds reverse router.
func (router *Router) buildReverse() error {
	router.reverse = make(reverseRouter)
	for _, route := range router.routeTable {
		paramNames := urlrouter.ParamNames(route.Path)
		names := make([]string, len(paramNames))
		for i := 0; i < len(paramNames); i++ {
			names[i] = paramNames[i][1:] // truncate the meta character.
		}
		router.reverse[route.Name] = &routeInfo{
			route:         route,
			rawParamNames: paramNames,
			paramNames:    names,
		}
	}
	return nil
}

type forwardRouter urlrouter.URLRouter
type reverseRouter map[string]*routeInfo

// Route represents a route.
type Route struct {
	Name        string
	Path        string
	Controller  interface{}
	MethodTypes map[string]MethodArgs
	paramNames  []string
}

func (route *Route) dispatch(methodName string, params []urlrouter.Param) (controller *reflect.Value, method *reflect.Value, args []reflect.Value) {
	methodArgs := route.MethodTypes[methodName]
	if methodArgs == nil {
		return nil, nil, nil
	}
	for _, param := range params {
		t := methodArgs[param.Name]
		validateParser := typeValidateParsers[t]
		arg, err := validateParser.Parse(param.Value)
		if err != nil {
			return nil, nil, nil
		}
		if !validateParser.Validate(arg) {
			return nil, nil, nil
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

func (route *Route) buildMethodTypes() error {
	controller := reflect.TypeOf(route.Controller)
	cname := controller.Name()
	pkgPath := controller.PkgPath()
	pkgDir, err := findPkgDir(pkgPath)
	if err != nil {
		return err
	}
	if pkgDir == "" {
		return fmt.Errorf("%v: package not found", pkgPath)
	}
	pkgInfo, err := util.ImportDir(pkgDir, 0)
	if err != nil {
		return err
	}
	astFiles := make([]*ast.File, len(pkgInfo.GoFiles))
	for i, goFilePath := range pkgInfo.GoFiles {
		if astFiles[i], err = parser.ParseFile(token.NewFileSet(), filepath.Join(pkgInfo.Dir, goFilePath), nil, 0); err != nil {
			return err
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
	return nil
}

func (route *Route) validateRouteParameters() error {
	var errors []string
	paramNames := make(map[string]bool)
	for _, name := range route.paramNames {
		paramNames[name] = true
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

func (route *Route) validateControllerArgumentParameters() error {
	argNames := make(map[string]bool)
	for _, args := range route.MethodTypes {
		for name := range args {
			argNames[name] = true
		}
	}
	var defNames []string
	for _, name := range route.paramNames {
		if !argNames[name] {
			defNames = append(defNames, name)
		}
	}
	if len(defNames) > 0 {
		var format string
		switch {
		case len(defNames) == 1:
			format = "route parameter `%v` is defined in route `%v`, but argument is not defined in %v of `%v`"
		case len(defNames) > 1:
			format = "route parameters `%v` are defined in route `%v`, but arguments are not defined in %v of `%v`"
		}
		names := strings.Join(defNames, "`, `")
		var methNames []string
		for name := range route.MethodTypes {
			methNames = append(methNames, name)
		}
		methName := strings.Join(methNames, "/")
		controllerName := reflect.TypeOf(route.Controller).Name()
		return fmt.Errorf(format, names, route.Name, methName, controllerName)
	}
	return nil
}

func (route *Route) validateTypeValidateParser() error {
	var errors []string
	for _, args := range route.MethodTypes {
		for _, t := range args {
			if typeValidateParsers[t] == nil {
				errors = append(errors, fmt.Sprintf("TypeValidateParser of %#v is not set", t))
			}
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

type MethodArgs map[string]string

type routeInfo struct {
	route         *Route
	rawParamNames []string
	paramNames    []string
}

// Reverse returns path of route by name and any params.
func Reverse(name string, v ...interface{}) string {
	info := appConfig.router.reverse[name]
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
	switch vlen, nlen := len(v), len(ri.paramNames); {
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
		validateParser := typeValidateParsers[t]
		if !validateParser.Validate(v[i]) {
			panic(fmt.Errorf("parameter type mismatch: %v (controller is %v)", route.Name, reflect.TypeOf(route.Controller).Name()))
		}
	}
	var oldnew []string
	for i := 0; i < len(v); i++ {
		oldnew = append(oldnew, ri.rawParamNames[i], fmt.Sprint(v[i]))
	}
	replacer := strings.NewReplacer(oldnew...)
	path := replacer.Replace(route.Path)
	return util.NormPath(path)
}

// TypeValidateParser is an interface of validator and parser for any type value.
type TypeValidateParser interface {
	// Validate returns whether the valid value as any type.
	Validate(v interface{}) bool

	// Parse returns value that parses v as any type.
	Parse(v string) (value interface{}, err error)
}

// SetTypeValidateParser sets the TypeValidateParser by given type name.
func SetTypeValidateParser(name string, validateParser TypeValidateParser) {
	typeValidateParsers[name] = validateParser
}

// SetTypeValidateParserByValue sets the TypeValidateParser by type of v.
func SetTypeValidateParserByValue(v interface{}, validateParser TypeValidateParser) {
	SetTypeValidateParser(reflect.TypeOf(v).String(), validateParser)
}

// StringTypeValidateParser represents a TypeValidateParser for string.
type StringTypeValidateParser struct {
	stringRegexp *regexp.Regexp
}

// Validate returns whether the valid value as string of path parameter.
func (validateParser *StringTypeValidateParser) Validate(v interface{}) bool {
	if s, ok := v.(string); ok {
		return validateParser.stringRegexp.MatchString(s)
	}
	return false
}

// Parse returns value that parses v as string type.
func (validateParser *StringTypeValidateParser) Parse(v string) (value interface{}, err error) {
	return v, nil
}

// IntTypeValidateParser represents a TypeValidateParser for int.
type IntTypeValidateParser struct {
	intRegexp *regexp.Regexp
}

// Validate returns whether the valid value as int of path parameter.
func (validateParser *IntTypeValidateParser) Validate(v interface{}) bool {
	if i, ok := v.(int); ok {
		return validateParser.intRegexp.MatchString(fmt.Sprint(i))
	}
	return false
}

// Parse returns value that parses v as int type.
func (validateParser *IntTypeValidateParser) Parse(v string) (value interface{}, err error) {
	return strconv.Atoi(v)
}

// URLTypeValidateParser represents a TypeValidateParser for url.URL.
type URLTypeValidateParser struct {
	urlRegexp *regexp.Regexp
}

// Validate returns whether the valid value as url.URL of path parameter.
func (validateParser *URLTypeValidateParser) Validate(v interface{}) bool {
	var s string
	switch t := v.(type) {
	case string:
		s = t
	case *url.URL:
		s = t.Path
	default:
		return false
	}
	return validateParser.urlRegexp.MatchString(s)
}

// Parse returns value that parses v as url.URL type.
func (validateParser *URLTypeValidateParser) Parse(v string) (value interface{}, err error) {
	return url.Parse(v)
}

func findPkgDir(pkgPath string) (string, error) {
	var pkgDir string
	for _, srcDir := range build.Default.SrcDirs() {
		path, err := filepath.Abs(filepath.Join(srcDir, pkgPath))
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(path); err == nil {
			pkgDir = path
			break
		}
	}
	return pkgDir, nil
}

func astTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%v.%v", t.X.(*ast.Ident).Name, t.Sel.Name)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", astTypeName(t.X))
	default:
		panic(fmt.Errorf("sorry, unexpected argument type `%T` found. please report this issue.", t))
	}
}
