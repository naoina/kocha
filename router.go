package kocha

import (
	"fmt"
	"go/ast"
	"go/build"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	"strings"

	"github.com/naoina/denco"
	"github.com/naoina/kocha/util"
)

// The routing table.
type RouteTable []*Route

func (rt RouteTable) buildRouter() (*Router, error) {
	router, err := newRouter(rt)
	if err != nil {
		return nil, err
	}
	for _, route := range rt {
		info := router.reverse[route.Name]
		route.paramNames = info.paramNames
	}
	return router, nil
}

// Router represents a router of kocha.
type Router struct {
	forward    *denco.Router
	reverse    map[string]*routeInfo
	routeTable RouteTable
	app        *Application
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

func (router *Router) dispatch(req *http.Request) (controller Controller, handler requestHandler, params denco.Params, found bool) {
	path := util.NormPath(req.URL.Path)
	data, params, found := router.forward.Lookup(path)
	if !found {
		return nil, nil, nil, false
	}
	route := data.(*Route)
	handler, found = route.dispatch(req.Method)
	return route.Controller, handler, params, found
}

// buildForward builds forward router.
func (router *Router) buildForward() error {
	records := make([]denco.Record, len(router.routeTable))
	for i, route := range router.routeTable {
		records[i] = denco.NewRecord(route.Path, route)
	}
	router.forward = denco.New()
	if err := router.forward.Build(records); err != nil {
		return err
	}
	return nil
}

// buildReverse builds reverse router.
func (router *Router) buildReverse() error {
	router.reverse = make(map[string]*routeInfo)
	for _, route := range router.routeTable {
		paramNames := route.ParamNames()
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

// Reverse returns path of route by name and any params.
func (router *Router) Reverse(name string, v ...interface{}) string {
	info := router.reverse[name]
	if info == nil {
		types := make([]string, len(v))
		for i, value := range v {
			types[i] = reflect.TypeOf(value).Name()
		}
		panic(fmt.Errorf("no match route found: %v (%v)", name, strings.Join(types, ", ")))
	}
	return info.reverse(v...)
}

// Route represents a route.
type Route struct {
	Name       string
	Path       string
	Controller Controller

	paramNames []string
}

func (route *Route) dispatch(method string) (handler requestHandler, found bool) {
	switch strings.ToUpper(method) {
	case "GET":
		if h, ok := route.Controller.(Getter); ok {
			return h.GET, true
		}
	case "POST":
		if h, ok := route.Controller.(Poster); ok {
			return h.POST, true
		}
	case "PUT":
		if h, ok := route.Controller.(Putter); ok {
			return h.PUT, true
		}
	case "DELETE":
		if h, ok := route.Controller.(Deleter); ok {
			return h.DELETE, true
		}
	case "HEAD":
		if h, ok := route.Controller.(Header); ok {
			return h.HEAD, true
		}
	case "PATCH":
		if h, ok := route.Controller.(Patcher); ok {
			return h.PATCH, true
		}
	}
	return nil, false
}

// ParamNames returns names of the path parameters.
func (route *Route) ParamNames() (names []string) {
	path := route.Path
	for i := 0; i < len(route.Path); i++ {
		if c := path[i]; c == denco.ParamCharacter || c == denco.WildcardCharacter {
			next := denco.NextSeparator(path, i+1)
			names = append(names, path[i:next])
			i = next
		}
	}
	return names
}

type routeInfo struct {
	route         *Route
	rawParamNames []string
	paramNames    []string
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
	var oldnew []string
	for i := 0; i < len(v); i++ {
		oldnew = append(oldnew, ri.rawParamNames[i], fmt.Sprint(v[i]))
	}
	replacer := strings.NewReplacer(oldnew...)
	path := replacer.Replace(route.Path)
	return util.NormPath(path)
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
