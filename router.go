package kocha

import (
	"fmt"
	"reflect"

	"strings"

	"github.com/naoina/denco"
	"github.com/naoina/kocha/util"
)

// The routing table.
type RouteTable []*Route

func (rt RouteTable) buildRouter() (*Router, error) {
	router := &Router{routeTable: rt}
	if err := router.buildForward(); err != nil {
		return nil, err
	}
	if err := router.buildReverse(); err != nil {
		return nil, err
	}
	return router, nil
}

// Router represents a router of kocha.
type Router struct {
	forward    *denco.Router
	reverse    map[string]*Route
	routeTable RouteTable
}

func (router *Router) dispatch(req *Request) (name string, handler requestHandler, params denco.Params, found bool) {
	path := util.NormPath(req.URL.Path)
	data, params, found := router.forward.Lookup(path)
	if !found {
		return "", nil, nil, false
	}
	route := data.(*Route)
	handler, found = route.dispatch(req.Method)
	return route.Name, handler, params, found
}

// buildForward builds forward router.
func (router *Router) buildForward() error {
	records := make([]denco.Record, len(router.routeTable))
	for i, route := range router.routeTable {
		records[i] = denco.NewRecord(route.Path, route)
	}
	router.forward = denco.New()
	return router.forward.Build(records)
}

// buildReverse builds reverse router.
func (router *Router) buildReverse() error {
	router.reverse = make(map[string]*Route)
	for _, route := range router.routeTable {
		router.reverse[route.Name] = route
		for i := 0; i < len(route.Path); i++ {
			if c := route.Path[i]; c == denco.ParamCharacter || c == denco.WildcardCharacter {
				next := denco.NextSeparator(route.Path, i+1)
				route.paramNames = append(route.paramNames, route.Path[i:next])
				i = next
			}
		}
	}
	return nil
}

// Reverse returns path of route by name and any params.
func (router *Router) Reverse(name string, v ...interface{}) (string, error) {
	route := router.reverse[name]
	if route == nil {
		types := make([]string, len(v))
		for i, value := range v {
			types[i] = reflect.TypeOf(value).Name()
		}
		return "", fmt.Errorf("kocha: no match route found: %v (%v)", name, strings.Join(types, ", "))
	}
	return route.reverse(v...)
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
func (route *Route) ParamNames() []string {
	return route.paramNames
}

func (r *Route) reverse(v ...interface{}) (string, error) {
	switch vlen, nlen := len(v), len(r.paramNames); {
	case vlen < nlen:
		return "", fmt.Errorf("kocha: too few arguments: %v (controller is %T)", r.Name, r.Controller)
	case vlen > nlen:
		return "", fmt.Errorf("kocha: too many arguments: %v (controller is %T)", r.Name, r.Controller)
	case vlen+nlen == 0:
		return r.Path, nil
	}
	var oldnew []string
	for i := 0; i < len(v); i++ {
		oldnew = append(oldnew, r.paramNames[i], fmt.Sprint(v[i]))
	}
	replacer := strings.NewReplacer(oldnew...)
	path := replacer.Replace(r.Path)
	return util.NormPath(path), nil
}
