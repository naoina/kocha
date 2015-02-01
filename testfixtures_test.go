package kocha

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

func NewTestApp() *Application {
	config := &Config{
		AppPath:       "testdata",
		AppName:       "appname",
		DefaultLayout: "application",
		Template: &Template{
			PathInfo: TemplatePathInfo{
				Name: "appname",
				Paths: []string{
					filepath.Join("testdata", "app", "view"),
				},
			},
		},
		RouteTable: RouteTable{
			{
				Name:       "root",
				Path:       "/",
				Controller: &FixtureRootTestCtrl{},
			},
			{
				Name:       "user",
				Path:       "/user/:id",
				Controller: &FixtureUserTestCtrl{},
			},
			{
				Name:       "date",
				Path:       "/:year/:month/:day/user/:name",
				Controller: &FixtureDateTestCtrl{},
			},
			{
				Name:       "error",
				Path:       "/error",
				Controller: &FixtureErrorTestCtrl{},
			},
			{
				Name:       "json",
				Path:       "/json",
				Controller: &FixtureJsonTestCtrl{},
			},
			{
				Name:       "teapot",
				Path:       "/teapot",
				Controller: &FixtureTeapotTestCtrl{},
			},
			{
				Name:       "panic_in_render",
				Path:       "/panic_in_render",
				Controller: &FixturePanicInRenderTestCtrl{},
			},
			{
				Name:       "static",
				Path:       "/static/*path",
				Controller: &StaticServe{},
			},
			{
				Name:       "post_test",
				Path:       "/post_test",
				Controller: &FixturePostTestCtrl{},
			},
			{
				Name: "error_controller_test",
				Path: "/error_controller_test",
				Controller: &ErrorController{
					StatusCode: http.StatusBadGateway,
				},
			},
		},
		Logger: &LoggerConfig{
			Writer: ioutil.Discard,
		},
		Middlewares: []Middleware{},
		Session: &SessionConfig{
			Name:  "test_session",
			Store: NewTestSessionCookieStore(),
		},
		MaxClientBodySize: DefaultMaxClientBodySize,
	}
	app, err := New(config)
	if err != nil {
		panic(err)
	}
	return app
}

func NewTestSessionCookieStore() *SessionCookieStore {
	return &SessionCookieStore{
		SecretKey:  "abcdefghijklmnopqrstuvwxyzABCDEF",
		SigningKey: "abcdefghijklmn",
	}
}

type orderedOutputMap map[string]interface{}

func (m orderedOutputMap) String() string {
	keys := make([]string, 0, len(m))
	for key, _ := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	outputs := make([]string, 0, len(keys))
	for _, key := range keys {
		outputs = append(outputs, fmt.Sprintf("%s:%s", key, m[key]))
	}
	return fmt.Sprintf("map[%v]", strings.Join(outputs, " "))
}

func (m orderedOutputMap) GoString() string {
	keys := make([]string, 0, len(m))
	for key, _ := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		keys[i] = fmt.Sprintf("%#v:%#v", key, m[key])
	}
	return fmt.Sprintf("map[string]interface{}{%v}", strings.Join(keys, ", "))
}

type FixturePanicInRenderTestCtrl struct {
	*DefaultController
}

func (ctrl *FixturePanicInRenderTestCtrl) GET(c *Context) Result {
	return RenderXML(c, Data{}) // Context is unsupported type in XML.
}

type FixtureUserTestCtrl struct {
	*DefaultController
}

func (ctrl *FixtureUserTestCtrl) GET(c *Context) Result {
	return Render(c, Data{
		"id": c.Params.Get("id"),
	})
}

type FixtureDateTestCtrl struct {
	DefaultController
}

func (ctrl *FixtureDateTestCtrl) GET(c *Context) Result {
	return Render(c, Data{
		"year":  c.Params.Get("year"),
		"month": c.Params.Get("month"),
		"day":   c.Params.Get("day"),
		"name":  c.Params.Get("name"),
	})
}

type FixtureErrorTestCtrl struct {
	DefaultController
}

func (ctrl *FixtureErrorTestCtrl) GET(c *Context) Result {
	panic("panic test")
}

type FixtureJsonTestCtrl struct {
	DefaultController
}

func (ctrl *FixtureJsonTestCtrl) GET(c *Context) Result {
	c.Response.ContentType = "application/json"
	return Render(c, nil)
}

type FixtureRootTestCtrl struct {
	*DefaultController
}

func (ctrl *FixtureRootTestCtrl) GET(c *Context) Result {
	return Render(c, nil)
}

type FixtureTeapotTestCtrl struct {
	DefaultController
}

func (ctrl *FixtureTeapotTestCtrl) GET(c *Context) Result {
	c.Response.StatusCode = http.StatusTeapot
	return Render(c, nil)
}

type FixtureInvalidReturnValueTypeTestCtrl struct {
	*DefaultController
}

func (ctrl *FixtureInvalidReturnValueTypeTestCtrl) GET(c *Context) string {
	return ""
}

type FixturePostTestCtrl struct {
	*DefaultController
}

func (ctrl *FixturePostTestCtrl) POST(c *Context) Result {
	m := orderedOutputMap{}
	for k, v := range c.Params.Values {
		m[k] = v
	}
	return Render(c, Data{"params": m})
}

type FixtureAnotherDelimsTestCtrl struct {
	*DefaultController
	Ctx string
}

func (ctrl *FixtureAnotherDelimsTestCtrl) GET(c *Context) Result {
	return Render(c, ctrl.Ctx)
}
