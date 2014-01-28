package kocha

import (
	"html/template"
	"net/http"
)

func newTestAppConfig() *AppConfig {
	return &AppConfig{
		AppPath:       "apppath/appname",
		AppName:       "appname",
		DefaultLayout: "app",
		TemplateSet: TemplateSet{
			"appname": {
				"app": {
					"html": {
						"fixture_root_test_ctrl":   template.Must(template.New("tmpl1").Parse(`tmpl1`)),
						"fixture_user_test_ctrl":   template.Must(template.New("tmpl2").Parse(`tmpl2-{{.id}}`)),
						"fixture_date_test_ctrl":   template.Must(template.New("tmpl3").Parse(`tmpl3-{{.name}}-{{.year}}-{{.month}}-{{.day}}`)),
						"fixture_error_test_ctrl":  template.Must(template.New("tmpl4").Parse(`tmpl4`)),
						"fixture_teapot_test_ctrl": template.Must(template.New("tmpl6").Parse(`teapot`)),
					},
					"json": {
						"fixture_json_test_ctrl": template.Must(template.New("tmpl5").Parse(`{"tmpl5":"json"}`)),
					},
				},
			},
		},
		Router: InitRouter(RouteTable{
			{
				Name:       "root",
				Path:       "/",
				Controller: FixtureRootTestCtrl{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{},
				},
			},
			{
				Name:       "user",
				Path:       "/user/:id",
				Controller: FixtureUserTestCtrl{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{
						"id": "int",
					},
				},
			},
			{
				Name:       "date",
				Path:       "/:year/:month/:day/user/:name",
				Controller: FixtureDateTestCtrl{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{
						"year":  "int",
						"month": "int",
						"day":   "int",
						"name":  "string",
					},
				},
			},
			{
				Name:       "error",
				Path:       "/error",
				Controller: FixtureErrorTestCtrl{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{},
				},
			},
			{
				Name:       "json",
				Path:       "/json",
				Controller: FixtureJsonTestCtrl{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{},
				},
			},
			{
				Name:       "teapot",
				Path:       "/teapot",
				Controller: FixtureTeapotTestCtrl{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{},
				},
			},
			{
				Name:       "panic_in_render",
				Path:       "/panic_in_render",
				Controller: FixturePanicInRenderTestCtrl{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{},
				},
			},
			{
				Name:       "static",
				Path:       "/static/*path",
				Controller: StaticServe{},
				MethodTypes: map[string]MethodArgs{
					"Get": MethodArgs{
						"path": "*url.URL",
					},
				},
			},
		}),
		Middlewares: append(DefaultMiddlewares, []Middleware{}...),
		Session: &SessionConfig{
			Name:  "test_session",
			Store: newTestSessionCookieStore(),
		},
	}
}

func newTestSessionCookieStore() *SessionCookieStore {
	return &SessionCookieStore{
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

type FixturePanicInRenderTestCtrl struct{ *Controller }

func (c *FixturePanicInRenderTestCtrl) Get() Result {
	return c.RenderXML(Context{}) // Context is unsupported type in XML.
}

type FixtureUserTestCtrl struct {
	*Controller
}

func (c *FixtureUserTestCtrl) Get(id int) Result {
	return c.Render(Context{
		"id": id,
	})
}

type FixtureDateTestCtrl struct {
	Controller
}

func (c *FixtureDateTestCtrl) Get(year, month int, day int, name string) Result {
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

func (c *FixtureErrorTestCtrl) Get() Result {
	panic("panic test")
	return c.Render()
}

type FixtureJsonTestCtrl struct {
	Controller
}

func (c *FixtureJsonTestCtrl) Get() Result {
	c.Response.ContentType = "application/json"
	return c.Render()
}

type FixtureRootTestCtrl struct {
	*Controller
}

func (c *FixtureRootTestCtrl) Get() Result {
	return c.Render()
}

type FixtureTeapotTestCtrl struct {
	Controller
}

func (c *FixtureTeapotTestCtrl) Get() Result {
	c.Response.StatusCode = http.StatusTeapot
	return c.Render()
}

type FixtureInvalidReturnValueTypeTestCtrl struct {
	*Controller
}

func (c *FixtureInvalidReturnValueTypeTestCtrl) Get() string {
	return ""
}

type FixtureInvalidNumberOfReturnValueTestCtrl struct{ *Controller }

func (c *FixtureInvalidNumberOfReturnValueTestCtrl) Get() (Result, Result) {
	return c.RenderText(""), c.RenderText("")
}

type FixtureTypeUndefinedCtrl struct{ *Controller }

func (c *FixtureTypeUndefinedCtrl) Get(id int32) Result {
	return c.RenderText("")
}
