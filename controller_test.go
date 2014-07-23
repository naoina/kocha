package kocha_test

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func newTestController(name, layout string) *kocha.Controller {
	app, err := kocha.New(&kocha.Config{
		AppPath:       "testdata",
		AppName:       "appname",
		DefaultLayout: "",
		Template: &kocha.Template{
			PathInfo: kocha.TemplatePathInfo{
				Name: "appname",
				Paths: []string{
					filepath.Join("testdata", "app", "views"),
				},
			},
		},
		RouteTable: []*kocha.Route{
			{
				Name:       name,
				Path:       "/",
				Controller: kocha.FixtureRootTestCtrl{},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	return &kocha.Controller{
		Name:     name,
		Layout:   layout,
		Request:  &kocha.Request{Request: req},
		Response: &kocha.Response{ResponseWriter: httptest.NewRecorder()},
		Params:   &kocha.Params{},
		App:      app,
	}
}

func TestMimeTypeFormats(t *testing.T) {
	var actual interface{} = len(kocha.MimeTypeFormats)
	var expected interface{} = 4
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`len(TestMimeTypeFormats) => %#v; want %#v`, actual, expected)
	}
	for k, v := range map[string]string{

		"application/json": "json",
		"application/xml":  "xml",
		"text/html":        "html",
		"text/plain":       "txt",
	} {
		if _, found := kocha.MimeTypeFormats[k]; !found {
			t.Errorf(`MimeTypeFormats["%#v"] => notfound; want %v`, k, v)
		}
	}
}

func TestMimeTypeFormats_Get(t *testing.T) {
	actual := kocha.MimeTypeFormats.Get("application/json")
	expected := "json"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = kocha.MimeTypeFormats.Get("text/plain")
	expected = "txt"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestMimeTypeFormats_Set(t *testing.T) {
	mimeType := "test/mime"
	if kocha.MimeTypeFormats[mimeType] != "" {
		t.Fatalf("Expect none, but %v", kocha.MimeTypeFormats[mimeType])
	}
	expected := "testmimetype"
	kocha.MimeTypeFormats.Set(mimeType, expected)
	defer delete(kocha.MimeTypeFormats, mimeType)
	actual := kocha.MimeTypeFormats[mimeType]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestMimeTypeFormats_Del(t *testing.T) {
	if kocha.MimeTypeFormats["text/html"] == "" {
		t.Fatal("Expect exists, but not exists")
	}
	kocha.MimeTypeFormats.Del("text/html")
	defer func() {
		kocha.MimeTypeFormats["text/html"] = "html"
	}()
	actual := kocha.MimeTypeFormats["text/html"]
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestController_Render_withTooManyContexts(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestController("testctrlr", "")
	c.Render(kocha.Context{}, kocha.Context{})
}

func TestController_Render_withoutContext(t *testing.T) {
	c := newTestController("testctrlr", "")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.Render().Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	var actual interface{} = string(buf)
	var expected interface{} = "tmpl\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}

	actual = c.Context
	expected = kocha.Context{"errors": c.Errors()}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
	}
}

func TestController_Render_WithContext(t *testing.T) {
	func() {
		c := newTestController("testctrlr_ctx", "")
		ctx := kocha.Context{
			"c1": "v1",
			"c2": "v2",
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.Render(ctx).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		ctx["errors"] = c.Errors()
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v\n", ctx)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
		if !reflect.DeepEqual(c.Response.ContentType, "text/html") {
			t.Errorf("Expect %v, but %v", "text/html", c.Response.ContentType)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "")
		c.Context = kocha.Context{
			"c3": "v3",
			"c4": "v4",
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.Render().Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v\n", c.Context)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "")
		c.Context = kocha.Context{
			"c5": "v5",
			"c6": "v6",
		}
		ctx := kocha.Context{
			"c6": "test",
			"c7": "v7",
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.Render(ctx).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v\n", kocha.Context{
			"c5":     "v5",
			"c6":     "test",
			"c7":     "v7",
			"errors": c.Errors(),
		})
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "")
		ctx := "test_ctx"
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.Render(ctx).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := "tmpl_ctx: test_ctx\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "")
		c.Context = kocha.Context{"c1": "v1"}
		ctx := "test_ctx_override"
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't occurred")
			}
		}()
		c.Render(ctx)
	}()

	func() {
		c := newTestController("testctrlr_ctx", "")
		c.Context = kocha.Context{"c1": "v1"}
		c.Render()
		actual := c.Context
		expected := kocha.Context{
			"c1":     "v1",
			"errors": make(map[string][]*kocha.ParamError),
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := newTestController("testctrlr_ctx", "")
		ctx := kocha.Context{"c1": "v1"}
		c.Render(ctx)
		actual := c.Context
		expected := kocha.Context{
			"c1":     "v1",
			"errors": make(map[string][]*kocha.ParamError),
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		origLog := kocha.Log
		defer func() {
			kocha.Log = origLog
		}()
		kocha.Log = &kocha.Logger{
			DEBUG: kocha.Loggers{kocha.NullLogger()},
			INFO:  kocha.Loggers{kocha.NullLogger()},
			WARN:  kocha.Loggers{kocha.NullLogger()},
			ERROR: kocha.Loggers{kocha.NullLogger()},
		}
		c := newTestController("testctrlr_ctx", "")
		c.Context = kocha.Context{"c1": "v1", "errors": "testerr"}
		c.Render()
		actual := c.Context
		expected := kocha.Context{
			"c1":     "v1",
			"errors": "testerr",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Context => %#v, want %#v", actual, expected)
		}
	}()
}

func TestController_Render_withContentType(t *testing.T) {
	c := newTestController("testctrlr", "")
	c.Response.ContentType = "application/json"
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.Render().Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := `{"tmpl2":"content"}` + "\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

func TestController_Render_withMissingTemplateInAppName(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestController("testctrlr", "")
	c.App.Config.AppName = "unknownAppName"
	c.Render()
}

func TestController_Render_withMissingTemplate(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestController("testctrlr", "")
	c.Name = "unknownctrlr"
	c.Render()
}

func TestController_Render_withAnotherLayout(t *testing.T) {
	c := newTestController("testctrlr", "another_layout")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.Render().Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := "Another layout\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

func TestController_RenderJSON(t *testing.T) {
	c := newTestController("testctrlr", "")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.RenderJSON(struct{ A, B string }{"hoge", "foo"}).Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
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

func TestController_RenderXML(t *testing.T) {
	c := newTestController("testctrlr", "")
	ctx := struct {
		XMLName xml.Name `xml:"user"`
		A       string   `xml:"id"`
		B       string   `xml:"name"`
	}{A: "hoge", B: "foo"}
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.RenderXML(ctx).Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
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

func TestController_RenderText(t *testing.T) {
	c := newTestController("testctrlr", "")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.RenderText("test_content_data").Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
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

func TestController_RenderError(t *testing.T) {
	c := newTestController("testctrlr", "")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.RenderError(http.StatusInternalServerError).Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	var actual interface{} = string(buf)
	var expected interface{} = "\nsingle 500 error\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusInternalServerError
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c = newTestController("testctrlr", "")
	w = httptest.NewRecorder()
	res = &kocha.Response{ResponseWriter: w}
	c.RenderError(http.StatusBadRequest).Proc(res)
	buf, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = "400 error\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusBadRequest
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c = newTestController("testctrlr", "")
	c.Response.ContentType = "application/json"
	w = httptest.NewRecorder()
	res = &kocha.Response{ResponseWriter: w}
	c.RenderError(http.StatusInternalServerError).Proc(res)
	buf, err = ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = `{"error":500}` + "\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusInternalServerError
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	func() {
		c = newTestController("testctrlr", "")
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		c.Response.ContentType = "unknown/content-type"
		c.RenderError(http.StatusInternalServerError)
	}()

	func() {
		c = newTestController("testctrlr", "")
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		c.RenderError(http.StatusInternalServerError, nil, nil)
	}()

	c = newTestController("testctrlr", "")
	w = httptest.NewRecorder()
	res = &kocha.Response{ResponseWriter: w}
	c.RenderError(http.StatusTeapot).Proc(res)
	buf, err = ioutil.ReadAll(w.Body)
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

func TestController_SendFile(t *testing.T) {
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
		c := newTestController("testctrlr", "")
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.SendFile(tmpFile.Name()).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
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
		tmpDir := filepath.Join(os.TempDir(), kocha.StaticDir)
		if err := os.Mkdir(tmpDir, 0755); err != nil {
			t.Fatal(err)
		}
		tmpFile, err := ioutil.TempFile(tmpDir, "TestControllerSendFile")
		if err != nil {
			panic(err)
		}
		defer tmpFile.Close()
		defer os.RemoveAll(tmpDir)
		c := newTestController("testctrlr", "")
		c.App.Config.AppPath = filepath.Dir(tmpDir)
		if _, err := tmpFile.WriteString("foobarbaz"); err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.SendFile(filepath.Base(tmpFile.Name())).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
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
		c := newTestController("testctrlr", "")
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.SendFile("unknown/path").Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
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
		c := newTestController("testctrlr", "")
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
		c := newTestController("testctrlr", "")
		c.SendFile(tmpFile.Name())
		actual := c.Response.ContentType
		expected := "application/javascript"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test with included resources
	func() {
		c := newTestController("testctrlr", "")
		c.App.ResourceSet.Add("testrcname", "foobarbaz")
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.SendFile("testrcname").Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
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
		c := newTestController("testctrlr", "")
		c.Response.ContentType = ""
		c.App.ResourceSet.Add("testrcname", "\x89PNG\x0d\x0a\x1a\x0a")
		c.SendFile("testrcname")
		actual := c.Response.ContentType
		expected := "image/png"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()
}

func TestController_Redirect(t *testing.T) {
	c := newTestController("testctrlr", "")
	for _, v := range []struct {
		redirectURL string
		permanent   bool
		expected    int
	}{
		{"/path/to/redirect/permanently", true, 301},
		{"/path/to/redirect", false, 302},
	} {
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		c.Redirect(v.redirectURL, v.permanent).Proc(res)
		actual := []interface{}{w.Code, w.HeaderMap.Get("Location")}
		expected := []interface{}{v.expected, v.redirectURL}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`Controller.Redirect("%#v", %#v) => %#v; want %#v`, v.redirectURL, v.permanent, actual, expected)
		}
	}
}

func TestController_Errors(t *testing.T) {
	func() {
		c := &kocha.Controller{}
		actual := c.Errors()
		expected := make(map[string][]*kocha.ParamError)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Errors() => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := &kocha.Controller{}
		c.Errors()["e1"] = []*kocha.ParamError{&kocha.ParamError{}}
		c.Errors()["e2"] = []*kocha.ParamError{&kocha.ParamError{}, &kocha.ParamError{}}
		actual := c.Errors()
		expected := map[string][]*kocha.ParamError{
			"e1": {&kocha.ParamError{}},
			"e2": {&kocha.ParamError{}, &kocha.ParamError{}},
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.Errors() => %#v, want %#v", actual, expected)
		}
	}()
}

func TestController_HasError(t *testing.T) {
	func() {
		c := &kocha.Controller{}
		actual := c.HasErrors()
		expected := false
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.HasErrors() => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := &kocha.Controller{}
		c.Errors()["e1"] = []*kocha.ParamError{&kocha.ParamError{}}
		actual := c.HasErrors()
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Controller.HasErrors() => %#v, want %#v", actual, expected)
		}
	}()
}

func TestStaticServe_Get(t *testing.T) {
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
	app := kocha.NewTestApp()
	c := &kocha.StaticServe{Controller: &kocha.Controller{App: app}}
	c.Controller.Request = &kocha.Request{Request: req}
	c.Controller.Response = &kocha.Response{ResponseWriter: httptest.NewRecorder()}
	u, err := url.Parse(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.Get(u).Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := []interface{}{w.Code, string(buf)}
	expected := []interface{}{200, ""}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`StaticServe.Get(%#v) => %#v; want %#v`, u, actual, expected)
	}
}

func TestNewErrorController(t *testing.T) {
	for _, v := range []int{
		http.StatusInternalServerError,
		http.StatusTeapot,
	} {
		actual := kocha.NewErrorController(v)
		expected := &kocha.ErrorController{StatusCode: v}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Value %v, expect %v, but %v", v, expected, actual)
		}
	}
}

func TestErrorController_Get(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	app := kocha.NewTestApp()
	c := &kocha.ErrorController{
		Controller: &kocha.Controller{App: app},
		StatusCode: http.StatusTeapot,
	}
	c.Controller.Request = &kocha.Request{Request: req}
	c.Controller.Response = &kocha.Response{ResponseWriter: httptest.NewRecorder()}
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	c.Get().Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := http.StatusText(http.StatusTeapot)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
