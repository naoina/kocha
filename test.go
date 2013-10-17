package kocha

import (
	"html/template"
	"regexp"
)

func newTestAppConfig() *AppConfig {
	return &AppConfig{
		AppPath: "apppath/appname",
		AppName: "appname",
		TemplateSet: TemplateSet{
			"appname": map[string]*template.Template{
				"fixture_root_test_ctrl.html": template.Must(template.New("tmpl1").Parse(`tmpl1`)),
				"fixture_user_test_ctrl.html": template.Must(template.New("tmpl2").Parse(`tmpl2-{{.id}}`)),
				"fixture_date_test_ctrl.html": template.Must(template.New("tmpl2").Parse(`tmpl3-{{.name}}-{{.year}}-{{.month}}-{{.day}}`)),
			},
		},
		RouteTable: []*Route{
			&Route{
				Name:       "root",
				Path:       "/",
				Controller: FixtureRootTestCtrl{},
				MethodTypes: map[string]methodArgs{
					"Get": methodArgs{},
				},
				RegexpPath: regexp.MustCompile(`^/$`),
			},
			&Route{
				Name:       "user",
				Path:       "/user/:id",
				Controller: FixtureUserTestCtrl{},
				MethodTypes: map[string]methodArgs{
					"Get": methodArgs{
						"id": "int",
					},
				},
				RegexpPath: regexp.MustCompile(`^/user/(?P<id>\d+)$`),
			},
			&Route{
				Name:       "date",
				Path:       "/:year/:month/:day/user/:name",
				Controller: FixtureDateTestCtrl{},
				MethodTypes: map[string]methodArgs{
					"Get": methodArgs{
						"year":  "int",
						"month": "int",
						"day":   "int",
						"name":  "string",
					},
				},
				RegexpPath: regexp.MustCompile(`^/(?P<year>\d+)/(?P<month>\d+)/(?P<day>\d+)/user/(?P<name>[\w-]+)$`),
			},
		},
	}
}
