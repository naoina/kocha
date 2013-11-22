package kocha

import (
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newControllerTestAppConfig() *AppConfig {
	return &AppConfig{
		AppPath: "testAppPath",
		AppName: "testAppName",
		TemplateSet: TemplateSet{
			"testAppName": map[string]*template.Template{
				"testctrlr.html":  template.Must(template.New("tmpl1").Parse(`tmpl1`)),
				"testctrlr.json":  template.Must(template.New("tmpl2").Parse(`{"tmpl2":"content"}`)),
				"errors/500.html": template.Must(template.New("tmpl3").Parse(`500 error`)),
				"errors/400.html": template.Must(template.New("tmpl4").Parse(`400 error`)),
				"errors/500.json": template.Must(template.New("tmpl5").Parse(`{"error":500}`)),
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

func TestControllerRenderText(t *testing.T) {
	c := newTestController()
	actual := c.RenderText("test_content_data")
	expected := &ResultText{"test_content_data"}
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
	c := newTestController()
	var actual interface{} = c.RenderError(http.StatusInternalServerError)
	var expected interface{} = &ResultTemplate{
		Template: appConfig.TemplateSet["testAppName"]["errors/500.html"],
		Context:  nil,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusInternalServerError
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c = newTestController()
	actual = c.RenderError(http.StatusBadRequest)
	expected = &ResultTemplate{
		Template: appConfig.TemplateSet["testAppName"]["errors/400.html"],
		Context:  nil,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusBadRequest
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c = newTestController()
	c.Response.ContentType = "application/json"
	actual = c.RenderError(http.StatusInternalServerError)
	expected = &ResultTemplate{
		Template: appConfig.TemplateSet["testAppName"]["errors/500.json"],
		Context:  nil,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusInternalServerError
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	func() {
		c = newTestController()
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		c.Response.ContentType = "unknown/content-type"
		c.RenderError(http.StatusInternalServerError)
	}()

	func() {
		c = newTestController()
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		c.RenderError(http.StatusInternalServerError, nil, nil)
	}()

	c = newTestController()
	actual = c.RenderError(http.StatusTeapot)
	expected = &ResultText{
		Content: http.StatusText(http.StatusTeapot),
	}
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
		c := newTestController()
		result, ok := c.SendFile(tmpFile.Name()).(*ResultContent)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultContent{}, result)
		}

		actual, err := ioutil.ReadAll(result.Reader)
		if err != nil {
			t.Fatal(err)
		}
		expected := []byte("foobarbaz")
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
		c := newTestController()
		result, ok := c.SendFile(filepath.Base(tmpFile.Name())).(*ResultContent)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultContent{}, result)
		}

		actual, err := ioutil.ReadAll(result.Reader)
		if err != nil {
			t.Fatal(err)
		}
		expected := []byte("foobarbaz")
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
		c := newTestController()
		result, ok := c.SendFile("unknown/path").(*ResultText)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultText{}, result)
		}
		actual := result.Content
		expected := http.StatusText(http.StatusNotFound)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test detect content type
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
		c := newTestController()
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

	// test with included resources
	func() {
		defer func() {
			includedResources = make(map[string]*resource)
		}()
		includedResources["testrcname"] = &resource{[]byte("foobarbaz")}
		c := newTestController()
		result, ok := c.SendFile("testrcname").(*ResultContent)
		if !ok {
			t.Errorf("Expect %T, but %T", &ResultContent{}, result)
		}

		actual, err := ioutil.ReadAll(result.Reader)
		if err != nil {
			t.Fatal(err)
		}
		expected := []byte("foobarbaz")
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test detect content type with included resources
	func() {
		defer func() {
			includedResources = make(map[string]*resource)
		}()
		c := newTestController()
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
	c := newTestController()
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
	c.Controller.Request = NewRequest(req)
	c.Controller.Response = NewResponse(w)
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
	c.Controller.Request = NewRequest(req)
	c.Controller.Response = NewResponse(w)
	actual := c.Get()
	expected := &ResultText{
		Content: http.StatusText(http.StatusTeapot),
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
