package kocha_test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/naoina/kocha"
)

func newTestApp() *kocha.Application {
	config := &kocha.Config{
		AppPath:       "testdata",
		AppName:       "appname",
		DefaultLayout: "app",
		Template: &kocha.Template{
			PathInfo: kocha.TemplatePathInfo{
				Name: "appname",
				Paths: []string{
					filepath.Join("testdata", "app", "views"),
				},
			},
		},
		RouteTable: kocha.RouteTable{
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
		Middlewares: append(kocha.DefaultMiddlewares, []kocha.Middleware{}...),
		Session: &kocha.SessionConfig{
			Name:  "test_session",
			Store: newTestSessionCookieStore(),
		},
		MaxClientBodySize: kocha.DefaultMaxClientBodySize,
	}
	app, err := kocha.New(config)
	if err != nil {
		panic(err)
	}
	return app
}

func newTestSessionCookieStore() *kocha.SessionCookieStore {
	return &kocha.SessionCookieStore{
		SecretKey:  "abcdefghijklmnopqrstuvwxyzABCDEF",
		SigningKey: "abcdefghijklmn",
	}
}

func testInvokeWrapper(f func()) {
	defer func() {
		failedUnits = make(map[string]bool)
	}()
	f()
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

type FixturePanicInRenderTestCtrl struct{ *kocha.Controller }

func (c *FixturePanicInRenderTestCtrl) Get() kocha.Result {
	return c.RenderXML(Context{}) // Context is unsupported type in XML.
}

type FixtureUserTestCtrl struct {
	*kocha.Controller
}

func (c *FixtureUserTestCtrl) Get(id int) kocha.Result {
	return c.Render(kocha.Context{
		"id": id,
	})
}

type FixtureDateTestCtrl struct {
	kocha.Controller
}

func (c *FixtureDateTestCtrl) Get(year, month int, day int, name string) kocha.Result {
	return c.Render(kocha.Context{
		"year":  year,
		"month": month,
		"day":   day,
		"name":  name,
	})
}

type FixtureErrorTestCtrl struct {
	kocha.Controller
}

func (c *FixtureErrorTestCtrl) Get() kocha.Result {
	panic("panic test")
}

type FixtureJsonTestCtrl struct {
	kocha.Controller
}

func (c *FixtureJsonTestCtrl) Get() kocha.Result {
	c.Response.ContentType = "application/json"
	return c.Render()
}

type FixtureRootTestCtrl struct {
	*kocha.Controller
}

func (c *FixtureRootTestCtrl) Get() kocha.Result {
	return c.Render()
}

type FixtureTeapotTestCtrl struct {
	kocha.Controller
}

func (c *FixtureTeapotTestCtrl) Get() kocha.Result {
	c.Response.StatusCode = http.StatusTeapot
	return c.Render()
}

type FixtureInvalidReturnValueTypeTestCtrl struct {
	*kocha.Controller
}

func (c *FixtureInvalidReturnValueTypeTestCtrl) Get() string {
	return ""
}

type FixtureInvalidNumberOfReturnValueTestCtrl struct{ *kocha.Controller }

func (c *FixtureInvalidNumberOfReturnValueTestCtrl) Get() (kocha.Result, kocha.Result) {
	return c.RenderText(""), c.RenderText("")
}

type FixtureTypeUndefinedCtrl struct{ *kocha.Controller }

func (c *FixtureTypeUndefinedCtrl) Get(id int32) kocha.Result {
	return c.RenderText("")
}

type FixturePostTestCtrl struct {
	*kocha.Controller
}

func (c *FixturePostTestCtrl) Post() kocha.Result {
	m := orderedOutputMap{}
	for k, v := range c.Params.Values {
		m[k] = v
	}
	return c.Render(kocha.Context{"params": m})
}
