package kocha

import (
	"encoding/xml"
	"html/template"
	"net/http"
	"net/http/httptest"
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
				"testctrlr.json": template.Must(template.New("tmpl2").Parse(`{"tmpl2":"content"}`)),
			},
		},
	}
}

func newTestController() *Controller {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	return &Controller{
		Name:     "testctrlr",
		Request:  NewRequest(req),
		Response: NewResponse(w),
		Params:   Params{},
	}
}

func TestMimeTypeFormats(t *testing.T) {
	actual := MimeTypeFormats
	expected := mimeTypeFormats{
		"application/json": "json",
		"application/xml":  "xml",
		"text/html":        "html",
		"text/plain":       "txt",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestMimeTypeFormatsGet(t *testing.T) {
	actual := MimeTypeFormats.Get("application/json")
	expected := "json"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = MimeTypeFormats.Get("text/plain")
	expected = "txt"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestMimeTypeFormatsSet(t *testing.T) {
	mimeType := "test/mime"
	if MimeTypeFormats[mimeType] != "" {
		t.Fatalf("Expect none, but %v", MimeTypeFormats[mimeType])
	}
	expected := "testmimetype"
	MimeTypeFormats.Set(mimeType, expected)
	defer delete(MimeTypeFormats, mimeType)
	actual := MimeTypeFormats[mimeType]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestMimeTypeFormatsDel(t *testing.T) {
	if MimeTypeFormats["text/html"] == "" {
		t.Fatal("Expect exists, but not exists")
	}
	MimeTypeFormats.Del("text/html")
	defer func() {
		MimeTypeFormats["text/html"] = "html"
	}()
	actual := MimeTypeFormats["text/html"]
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
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
	if !reflect.DeepEqual(c.Response.ContentType, "text/html") {
		t.Errorf("Expect %v, but %v", "text/html", c.Response.ContentType)
	}
}

func TestControllerRender_with_ContentType(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	c := newTestController()
	c.Response.ContentType = "application/json"
	actual := c.Render()
	expected := &ResultTemplate{
		Template: appConfig.TemplateSet["testAppName"]["testctrlr.json"],
		Context:  nil,
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
	if !reflect.DeepEqual(c.Response.ContentType, "application/json") {
		t.Errorf("Expect %v, but %v", "application/json", c.Response.ContentType)
	}
}

func TestControllerRenderXML(t *testing.T) {
	c := newTestController()
	ctx := struct {
		XMLName xml.Name `xml:"user"`
		A       string   `xml:"id"`
		B       string   `xml:"name"`
	}{A: "hoge", B: "foo"}
	actual := c.RenderXML(ctx)
	expected := &ResultXML{ctx}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if !reflect.DeepEqual(c.Response.ContentType, "application/xml") {
		t.Errorf("Expect %v, but %v", "application/xml", c.Response.ContentType)
	}
}

func TestControllerRenderPlainText(t *testing.T) {
	c := newTestController()
	actual := c.RenderPlainText("test_content_data")
	expected := &ResultPlainText{"test_content_data"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if !reflect.DeepEqual(c.Response.ContentType, "text/plain") {
		t.Errorf("Expect %v, but %v", "text/plain", c.Response.ContentType)
	}
}
