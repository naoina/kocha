package kocha

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newControllerTestAppConfig() *AppConfig {
	config := &AppConfig{
		AppPath:     "testAppPath",
		AppName:     "testAppName",
		TemplateSet: TemplateSet{},
	}
	config.templateMap = TemplateMap{
		"testAppName": {
			"app": {
				"html": {
					"testctrlr":     template.Must(template.New("tmpl1").Parse(`tmpl1`)),
					"testctrlr_ctx": template.Must(template.New("tmpl1_ctx").Parse(`tmpl_ctx: {{.}}`)),
					"errors/500":    template.Must(template.New("tmpl3").Parse(`500 error`)),
					"errors/400":    template.Must(template.New("tmpl4").Parse(`400 error`)),
				},
				"json": {

					"testctrlr":     template.Must(template.New("tmpl2").Parse(`{"tmpl2":"content"}`)),
					"testctrlr_ctx": template.Must(template.New("tmpl2_ctx").Parse("tmpl2_ctx: {{.}}")),
					"errors/500":    template.Must(template.New("tmpl5").Parse(`{"error":500}`)),
				},
			},
			"anotherLayout": {
				"html": {
					"testctrlr": template.Must(template.New("a_tmpl1").Parse(`<b>a_tmpl1</b>`)),
				},
			},
		},
	}
	return config
}

func newTestController(name, layout string) *Controller {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	return &Controller{
		Name:     name,
		Layout:   layout,
		Request:  newRequest(req),
		Response: newResponse(w),
		Params:   &Params{},
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
	c := newTestController("testctrlr", "app")
	c.Render(Context{}, Context{})
}

func TestControllerRender_without_Context(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	c := newTestController("testctrlr", "app")
	buf, err := ioutil.ReadAll(c.Render().(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	var actual interface{} = string(buf)
	var expected interface{} = "tmpl1"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = c.Context
	expected = Context{"errors": c.errors}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
	}
}

func TestControllerRender_with_Context(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()

	func() {
		c := newTestController("testctrlr_ctx", "app")
		ctx := Context{
			"c1": "v1",
			"c2": "v2",
		}
		buf, err := ioutil.ReadAll(c.Render(ctx).(*ResultContent).Body)
		if err != nil {
			t.Fatal(err)
		}
		ctx["errors"] = c.errors
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v", ctx)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
		if !reflect.DeepEqual(c.Response.ContentType, "text/html") {
			t.Errorf("Expect %v, but %v", "text/html", c.Response.ContentType)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "app")
		c.Context = Context{
			"c3": "v3",
			"c4": "v4",
		}
		buf, err := ioutil.ReadAll(c.Render().(*ResultContent).Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v", c.Context)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "app")
		c.Context = Context{
			"c5": "v5",
			"c6": "v6",
		}
		ctx := Context{
			"c6": "test",
			"c7": "v7",
		}
		buf, err := ioutil.ReadAll(c.Render(ctx).(*ResultContent).Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v", Context{
			"c5":     "v5",
			"c6":     "test",
			"c7":     "v7",
			"errors": c.errors,
		})
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "app")
		ctx := "test_ctx"
		buf, err := ioutil.ReadAll(c.Render(ctx).(*ResultContent).Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := "tmpl_ctx: test_ctx"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "app")
		c.Context = Context{"c1": "v1"}
		ctx := "test_ctx_override"
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't occurred")
			}
		}()
		c.Render(ctx)
	}()

	func() {
		c := newTestController("testctrlr_ctx", "app")
		c.Context = Context{"c1": "v1"}
		c.Render()
		actual := c.Context
		expected := Context{
			"c1":     "v1",
			"errors": make(map[string][]*ParamError),
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "app")
		ctx := Context{"c1": "v1"}
		c.Render(ctx)
		actual := c.Context
		expected := Context{
			"c1":     "v1",
			"errors": make(map[string][]*ParamError),
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		origLog := Log
		defer func() {
			Log = origLog
		}()
		Log = initLogger(nil)
		c := newTestController("testctrlr_ctx", "app")
		c.Context = Context{"c1": "v1", "errors": "testerr"}
		c.Render()
		actual := c.Context
		expected := Context{
			"c1":     "v1",
			"errors": "testerr",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
		}
	}()
}

func TestControllerRender_with_ContentType(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	c := newTestController("testctrlr", "app")
	c.Response.ContentType = "application/json"
	buf, err := ioutil.ReadAll(c.Render().(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := `{"tmpl2":"content"}`
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
	c := newTestController("testctrlr", "app")
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
	c := newTestController("testctrlr", "app")
	c.Name = "unknownctrlr"
	c.Render()
}

func TestControllerRender_with_another_layout(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	c := newTestController("testctrlr", "anotherLayout")
	buf, err := ioutil.ReadAll(c.Render().(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := "<b>a_tmpl1</b>"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestControllerRenderJSON(t *testing.T) {
	c := newTestController("testctrlr", "app")
	buf, err := ioutil.ReadAll(c.RenderJSON(struct{ A, B string }{"hoge", "foo"}).(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := `{"A":"hoge","B":"foo"}`
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if !reflect.DeepEqual(c.Response.ContentType, "application/json") {
		t.Errorf("Expect %v, but %v", "application/json", c.Response.ContentType)
	}
}

func TestControllerRenderXML(t *testing.T) {
	c := newTestController("testctrlr", "app")
	ctx := struct {
		XMLName xml.Name `xml:"user"`
		A       string   `xml:"id"`
		B       string   `xml:"name"`
	}{A: "hoge", B: "foo"}
	buf, err := ioutil.ReadAll(c.RenderXML(ctx).(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := "<user><id>hoge</id><name>foo</name></user>"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if !reflect.DeepEqual(c.Response.ContentType, "application/xml") {
		t.Errorf("Expect %v, but %v", "application/xml", c.Response.ContentType)
	}
}

func TestControllerRenderText(t *testing.T) {
	c := newTestController("testctrlr", "app")
	buf, err := ioutil.ReadAll(c.RenderText("test_content_data").(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := "test_content_data"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if !reflect.DeepEqual(c.Response.ContentType, "text/plain") {
		t.Errorf("Expect %v, but %v", "text/plain", c.Response.ContentType)
	}
}

func TestControllerRenderError(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	c := newTestController("testctrlr", "app")
	buf, err := ioutil.ReadAll(c.RenderError(http.StatusInternalServerError).(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	var actual interface{} = string(buf)
	var expected interface{} = "500 error"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusInternalServerError
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c = newTestController("testctrlr", "app")
	buf, err = ioutil.ReadAll(c.RenderError(http.StatusBadRequest).(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = "400 error"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusBadRequest
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c = newTestController("testctrlr", "app")
	c.Response.ContentType = "application/json"
	buf, err = ioutil.ReadAll(c.RenderError(http.StatusInternalServerError).(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = `{"error":500}`
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusInternalServerError
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	func() {
		c = newTestController("testctrlr", "app")
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		c.Response.ContentType = "unknown/content-type"
		c.RenderError(http.StatusInternalServerError)
	}()

	func() {
		c = newTestController("testctrlr", "app")
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		c.RenderError(http.StatusInternalServerError, nil, nil)
	}()

	c = newTestController("testctrlr", "app")
	buf, err = ioutil.ReadAll(c.RenderError(http.StatusTeapot).(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = http.StatusText(http.StatusTeapot)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusTeapot
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestControllerSendFile(t *testing.T) {
	// general test
	func() {
		tmpFile, err := ioutil.TempFile("", "TestControllerSendFile")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.WriteString("foobarbaz"); err != nil {
			t.Fatal(err)
		}
		c := newTestController("testctrlr", "app")
		result, ok := c.SendFile(tmpFile.Name()).(*ResultContent)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultContent{}, result)
		}

		buf, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := "foobarbaz"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test default static path
	func() {
		tmpDir := filepath.Join(os.TempDir(), StaticDir)
		if err := os.Mkdir(tmpDir, 0755); err != nil {
			t.Fatal(err)
		}
		tmpFile, err := ioutil.TempFile(tmpDir, "TestControllerSendFile")
		if err != nil {
			panic(err)
		}
		defer tmpFile.Close()
		defer os.RemoveAll(tmpDir)
		oldAppConfig := appConfig
		appConfig = newControllerTestAppConfig()
		appConfig.AppPath = filepath.Dir(tmpDir)
		defer func() {
			appConfig = oldAppConfig
		}()
		if _, err := tmpFile.WriteString("foobarbaz"); err != nil {
			t.Fatal(err)
		}
		c := newTestController("testctrlr", "app")
		result, ok := c.SendFile(filepath.Base(tmpFile.Name())).(*ResultContent)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultContent{}, result)
		}

		buf, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := "foobarbaz"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test file not found
	func() {
		oldAppConfig := appConfig
		appConfig = newControllerTestAppConfig()
		defer func() {
			appConfig = oldAppConfig
		}()
		c := newTestController("testctrlr", "app")
		result, ok := c.SendFile("unknown/path").(*ResultContent)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultContent{}, result)
		}
		buf, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := http.StatusText(http.StatusNotFound)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test detect content type by body
	func() {
		tmpFile, err := ioutil.TempFile("", "TestControllerSendFile")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.WriteString("foobarbaz"); err != nil {
			t.Fatal(err)
		}
		c := newTestController("testctrlr", "app")
		c.SendFile(tmpFile.Name())
		actual := c.Response.ContentType
		expected := "text/plain; charset=utf-8"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}

		c.Response.ContentType = ""
		if _, err := tmpFile.Seek(0, os.SEEK_SET); err != nil {
			t.Fatal(err)
		}
		if _, err := tmpFile.Write([]byte("\x89PNG\x0d\x0a\x1a\x0a")); err != nil {
			t.Fatal(err)
		}
		c.SendFile(tmpFile.Name())
		actual = c.Response.ContentType
		expected = "image/png"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test detect content type by ext
	func() {
		currentPath, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		tmpFile, err := os.Open(filepath.Join(currentPath, "testdata", "public", "test.js"))
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()
		mime.AddExtensionType(".js", "application/javascript") // To avoid differences between environments.
		c := newTestController("testctrlr", "app")
		c.SendFile(tmpFile.Name())
		actual := c.Response.ContentType
		expected := "application/javascript"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test with included resources
	func() {
		defer func() {
			includedResources = make(map[string]*resource)
		}()
		includedResources["testrcname"] = &resource{[]byte("foobarbaz")}
		c := newTestController("testctrlr", "app")
		result, ok := c.SendFile("testrcname").(*ResultContent)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultContent{}, result)
		}

		buf, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := "foobarbaz"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test detect content type with included resources
	func() {
		defer func() {
			includedResources = make(map[string]*resource)
		}()
		c := newTestController("testctrlr", "app")
		c.Response.ContentType = ""
		includedResources["testrcname"] = &resource{[]byte("\x89PNG\x0d\x0a\x1a\x0a")}
		c.SendFile("testrcname")
		actual := c.Response.ContentType
		expected := "image/png"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()
}

func TestControllerRedirect(t *testing.T) {
	c := newTestController("testctrlr", "app")
	actual := c.Redirect("/path/to/redirect/permanently", true)
	expected := &ResultRedirect{
		Request:     c.Request,
		URL:         "/path/to/redirect/permanently",
		Permanently: true,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = c.Redirect("/path/to/redirect", false)
	expected = &ResultRedirect{
		Request:     c.Request,
		URL:         "/path/to/redirect",
		Permanently: false,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestControllerErrors(t *testing.T) {
	func() {
		c := &Controller{}
		actual := c.Errors()
		expected := make(map[string][]*ParamError)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Errors() => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := &Controller{}
		c.errors = map[string][]*ParamError{
			"e1": {&ParamError{}},
			"e2": {&ParamError{}, &ParamError{}},
		}
		actual := c.Errors()
		expected := map[string][]*ParamError{
			"e1": {&ParamError{}},
			"e2": {&ParamError{}, &ParamError{}},
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Errors() => %#v, want %#v", actual, expected)
		}
	}()
}

func TestControllerHasError(t *testing.T) {
	func() {
		c := &Controller{}
		actual := c.HasErrors()
		expected := false
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.HasErrors() => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := &Controller{}
		c.errors = map[string][]*ParamError{
			"e1": {&ParamError{}},
		}
		actual := c.HasErrors()
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.HasErrors() => %#v, want %#v", actual, expected)
		}
	}()
}

func TestStaticServeGet(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	tmpFile, err := ioutil.TempFile("", "TestStaticServeGet")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()
	defer os.RemoveAll(tmpFile.Name())
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	c := &StaticServe{Controller: &Controller{}}
	c.Controller.Request = newRequest(req)
	c.Controller.Response = newResponse(w)
	u, err := url.Parse(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	result, ok := c.Get(u).(*ResultContent)
	if !ok {
		t.Errorf("Expect %T, but %T", &ResultContent{}, result)
	}
}

func TestNewErrorController(t *testing.T) {
	for _, v := range []int{
		http.StatusInternalServerError,
		http.StatusTeapot,
	} {
		actual := NewErrorController(v)
		expected := &ErrorController{StatusCode: v}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Value %v, expect %v, but %v", v, expected, actual)
		}
	}
}

func TestErrorControllerGet(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newControllerTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	w := httptest.NewRecorder()
	c := &ErrorController{
		Controller: &Controller{},
		StatusCode: http.StatusTeapot,
	}
	c.Controller.Request = newRequest(req)
	c.Controller.Response = newResponse(w)
	buf, err := ioutil.ReadAll(c.Get().(*ResultContent).Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := http.StatusText(http.StatusTeapot)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
