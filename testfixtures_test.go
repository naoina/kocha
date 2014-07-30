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
					filepath.Join("testdata", "app", "views"),
				},
			},
		},
		RouteTable: RouteTable{
			{
				Name:       "root",
				Path:       "/",
				Controller: FixtureRootTestCtrl{},
			},
			{
				Name:       "user",
				Path:       "/user/:id",
				Controller: FixtureUserTestCtrl{},
			},
			{
				Name:       "date",
				Path:       "/:year/:month/:day/user/:name",
				Controller: FixtureDateTestCtrl{},
			},
			{
				Name:       "error",
				Path:       "/error",
				Controller: FixtureErrorTestCtrl{},
			},
			{
				Name:       "json",
				Path:       "/json",
				Controller: FixtureJsonTestCtrl{},
			},
			{
				Name:       "teapot",
				Path:       "/teapot",
				Controller: FixtureTeapotTestCtrl{},
			},
			{
				Name:       "panic_in_render",
				Path:       "/panic_in_render",
				Controller: FixturePanicInRenderTestCtrl{},
			},
			{
				Name:       "static",
				Path:       "/static/*path",
				Controller: StaticServe{},
			},
			{
				Name:       "post_test",
				Path:       "/post_test",
				Controller: FixturePostTestCtrl{},
			},
		},
		Logger: &LoggerConfig{
			Writer: ioutil.Discard,
		},
		Middlewares: append(DefaultMiddlewares, []Middleware{}...),
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

type FixturePanicInRenderTestCtrl struct{ *Controller }

func (c *FixturePanicInRenderTestCtrl) GET() Result {
	return c.RenderXML(Context{}) // Context is unsupported type in XML.
}

type FixtureUserTestCtrl struct {
	*Controller
}

func (c *FixtureUserTestCtrl) GET(id int) Result {
	return c.Render(Context{
		"id": id,
	})
}

type FixtureDateTestCtrl struct {
	Controller
}

func (c *FixtureDateTestCtrl) GET(year, month int, day int, name string) Result {
	return c.Render(Context{
		"year":  year,
		"month": month,
		"day":   day,
		"name":  name,
	})
}

type FixtureErrorTestCtrl struct {
	Controller
}

func (c *FixtureErrorTestCtrl) GET() Result {
	panic("panic test")
}

type FixtureJsonTestCtrl struct {
	Controller
}

func (c *FixtureJsonTestCtrl) GET() Result {
	c.Response.ContentType = "application/json"
	return c.Render()
}

type FixtureRootTestCtrl struct {
	*Controller
}

func (c *FixtureRootTestCtrl) GET() Result {
	return c.Render()
}

type FixtureTeapotTestCtrl struct {
	Controller
}

func (c *FixtureTeapotTestCtrl) GET() Result {
	c.Response.StatusCode = http.StatusTeapot
	return c.Render()
}

type FixtureInvalidReturnValueTypeTestCtrl struct {
	*Controller
}

func (c *FixtureInvalidReturnValueTypeTestCtrl) GET() string {
	return ""
}

type FixtureInvalidNumberOfReturnValueTestCtrl struct{ *Controller }

func (c *FixtureInvalidNumberOfReturnValueTestCtrl) GET() (Result, Result) {
	return c.RenderText(""), c.RenderText("")
}

type FixtureTypeUndefinedCtrl struct{ *Controller }

func (c *FixtureTypeUndefinedCtrl) GET(id int32) Result {
	return c.RenderText("")
}

type FixturePostTestCtrl struct {
	*Controller
}

func (c *FixturePostTestCtrl) POST() Result {
	m := orderedOutputMap{}
	for k, v := range c.Params.Values {
		m[k] = v
	}
	return c.Render(Context{"params": m})
}
