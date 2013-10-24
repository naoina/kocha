package kocha

import (
	"html/template"
	"reflect"
	"testing"
)

func newControllerTestAppConfig() *AppConfig {
	return &AppConfig{
		AppPath: "testAppPath",
		AppName: "testAppName",
		TemplateSet: TemplateSet{
			"testAppName": map[string]*template.Template{
				"testctrlr.html": template.Must(template.New("tmpl1").Parse(`tmpl1`)),
			},
		},
	}
}

func newTestController() *Controller {
	return &Controller{
		Name: "testctrlr",
	}
}

func TestControllerRender_with_too_many_contexts(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestController()
	c.Render(Context{}, Context{})
}

func TestControllerRender_without_Context(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	c := newTestController()
	actual := c.Render()
	expected := &ResultTemplate{
		Template: appConfig.TemplateSet["testAppName"]["testctrlr.html"],
		Context:  nil,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestControllerRender_with_Context(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	c := newTestController()
	ctx := Context{
		"c1": "v1",
		"c2": "v2",
	}
	actual := c.Render(ctx)
	expected := &ResultTemplate{
		Template: appConfig.TemplateSet["testAppName"]["testctrlr.html"],
		Context:  ctx,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestControllerRender_with_missing_Template_in_AppName(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestController()
	appConfig.AppName = "unknownAppName"
	c.Render()
}

func TestControllerRender_with_missing_Template(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestController()
	c.Name = "unknownctrlr"
	c.Render()
}

func TestControllerRenderJSON(t *testing.T) {
	c := newTestController()
	actual := c.RenderJSON(struct{ A, B string }{"hoge", "foo"})
	expected := &ResultJSON{
		Context: struct{ A, B string }{"hoge", "foo"},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
