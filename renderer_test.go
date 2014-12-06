package kocha_test

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func newTestContext(name, layout string) *kocha.Context {
	app, err := kocha.New(&kocha.Config{
		AppPath:       "testdata",
		AppName:       "appname",
		DefaultLayout: "",
		Template: &kocha.Template{
			PathInfo: kocha.TemplatePathInfo{
				Name: "appname",
				Paths: []string{
					filepath.Join("testdata", "app", "view"),
				},
			},
		},
		RouteTable: []*kocha.Route{
			{
				Name:       name,
				Path:       "/",
				Controller: &kocha.FixtureRootTestCtrl{},
			},
		},
		Logger: &kocha.LoggerConfig{
			Writer: ioutil.Discard,
		},
	})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		panic(err)
	}
	return &kocha.Context{
		Name:     name,
		Layout:   layout,
		Request:  &kocha.Request{Request: req},
		Response: &kocha.Response{ResponseWriter: httptest.NewRecorder()},
		Params:   &kocha.Params{},
		App:      app,
	}
}

func TestContext_Render(t *testing.T) {
	func() {
		c := newTestContext("testctrlr_ctx", "")
		data := kocha.Data{
			"c1": "v1",
			"c2": "v2",
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.Render(c, data).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v\n", data)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
		if !reflect.DeepEqual(c.Response.ContentType, "text/html") {
			t.Errorf("Expect %v, but %v", "text/html", c.Response.ContentType)
		}
	}()

	func() {
		c := newTestContext("testctrlr_ctx", "")
		c.Data = kocha.Data{
			"c3": "v3",
			"c4": "v4",
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.Render(c, nil).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v\n", c.Data)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestContext("testctrlr_ctx", "")
		c.Data = kocha.Data{
			"c5": "v5",
			"c6": "v6",
		}
		ctx := kocha.Data{
			"c6": "test",
			"c7": "v7",
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.Render(c, ctx).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v\n", kocha.Data{
			"c5": "v5",
			"c6": "test",
			"c7": "v7",
		})
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestContext("testctrlr_ctx", "")
		ctx := "test_ctx"
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.Render(c, ctx).Proc(res)
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
		c := newTestContext("testctrlr_ctx", "")
		c.Data = kocha.Data{"c1": "v1"}
		ctx := "test_ctx_override"
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.Render(c, ctx).Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := "tmpl_ctx: test_ctx_override\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		c := newTestContext("testctrlr_ctx", "")
		c.Data = kocha.Data{"c1": "v1"}
		kocha.Render(c, nil)
		actual := c.Data
		expected := kocha.Data{
			"c1": "v1",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Context.Data => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := newTestContext("testctrlr_ctx", "")
		ctx := kocha.Data{"c1": "v1"}
		kocha.Render(c, ctx)
		actual := c.Data
		expected := kocha.Data{
			"c1": "v1",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Context.Data => %#v, want %#v", actual, expected)
		}
	}()
}

func TestContext_Render_withContentType(t *testing.T) {
	c := newTestContext("testctrlr", "")
	c.Response.ContentType = "application/json"
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	kocha.Render(c, nil).Proc(res)
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

func TestContext_Render_withMissingTemplateInAppName(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestContext("testctrlr", "")
	c.App.Config.AppName = "unknownAppName"
	kocha.Render(c, nil)
}

func TestContext_Render_withMissingTemplate(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("panic doesn't happened")
		}
	}()
	c := newTestContext("testctrlr", "")
	c.Name = "unknownctrlr"
	kocha.Render(c, nil)
}

func TestContext_Render_withAnotherLayout(t *testing.T) {
	c := newTestContext("testctrlr", "another_layout")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	kocha.Render(c, nil).Proc(res)
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

func TestContext_RenderJSON(t *testing.T) {
	c := newTestContext("testctrlr", "")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	kocha.RenderJSON(c, struct{ A, B string }{"hoge", "foo"}).Proc(res)
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

func TestContext_RenderXML(t *testing.T) {
	c := newTestContext("testctrlr", "")
	ctx := struct {
		XMLName xml.Name `xml:"user"`
		A       string   `xml:"id"`
		B       string   `xml:"name"`
	}{A: "hoge", B: "foo"}
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	kocha.RenderXML(c, ctx).Proc(res)
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

func TestContext_RenderText(t *testing.T) {
	c := newTestContext("testctrlr", "")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	kocha.RenderText(c, "test_content_data").Proc(res)
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

func TestContext_RenderError(t *testing.T) {
	c := newTestContext("testctrlr", "")
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	kocha.RenderError(c, http.StatusInternalServerError, nil).Proc(res)
	buf, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	var actual interface{} = string(buf)
	var expected interface{} = "500 error\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
	actual = c.Response.StatusCode
	expected = http.StatusInternalServerError
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c = newTestContext("testctrlr", "")
	w = httptest.NewRecorder()
	res = &kocha.Response{ResponseWriter: w}
	kocha.RenderError(c, http.StatusBadRequest, nil).Proc(res)
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

	c = newTestContext("testctrlr", "")
	c.Response.ContentType = "application/json"
	w = httptest.NewRecorder()
	res = &kocha.Response{ResponseWriter: w}
	kocha.RenderError(c, http.StatusInternalServerError, nil).Proc(res)
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
		c = newTestContext("testctrlr", "")
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		c.Response.ContentType = "unknown/content-type"
		kocha.RenderError(c, http.StatusInternalServerError, nil)
	}()

	c = newTestContext("testctrlr", "")
	w = httptest.NewRecorder()
	res = &kocha.Response{ResponseWriter: w}
	kocha.RenderError(c, http.StatusTeapot, nil).Proc(res)
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

func TestContext_SendFile(t *testing.T) {
	// general test
	func() {
		tmpFile, err := ioutil.TempFile("", "TestContextSendFile")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.WriteString("foobarbaz"); err != nil {
			t.Fatal(err)
		}
		c := newTestContext("testctrlr", "")
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.SendFile(c, tmpFile.Name()).Proc(res)
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
		tmpFile, err := ioutil.TempFile(tmpDir, "TestContextSendFile")
		if err != nil {
			panic(err)
		}
		defer tmpFile.Close()
		defer os.RemoveAll(tmpDir)
		c := newTestContext("testctrlr", "")
		c.App.Config.AppPath = filepath.Dir(tmpDir)
		if _, err := tmpFile.WriteString("foobarbaz"); err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.SendFile(c, filepath.Base(tmpFile.Name())).Proc(res)
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
		c := newTestContext("testctrlr", "")
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.SendFile(c, "unknown/path").Proc(res)
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		file, err := ioutil.ReadFile(filepath.Join(c.App.Config.AppPath, "app", "view", "error", "404.html"))
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expect := string(file)
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`kocha.SendFile(c, "unknown/path").Proc(res); body => %#v; want %#v`, actual, expect)
		}
	}()

	// test detect content type by body
	func() {
		tmpFile, err := ioutil.TempFile("", "TestContextSendFile")
		if err != nil {
			t.Fatal(err)
		}
		defer tmpFile.Close()
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.WriteString("foobarbaz"); err != nil {
			t.Fatal(err)
		}
		c := newTestContext("testctrlr", "")
		kocha.SendFile(c, tmpFile.Name())
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
		kocha.SendFile(c, tmpFile.Name())
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
		c := newTestContext("testctrlr", "")
		kocha.SendFile(c, tmpFile.Name())
		actual := c.Response.ContentType
		expected := "application/javascript"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test with included resources
	func() {
		c := newTestContext("testctrlr", "")
		c.App.ResourceSet.Add("testrcname", "foobarbaz")
		w := httptest.NewRecorder()
		res := &kocha.Response{ResponseWriter: w}
		kocha.SendFile(c, "testrcname").Proc(res)
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
		c := newTestContext("testctrlr", "")
		c.Response.ContentType = ""
		c.App.ResourceSet.Add("testrcname", "\x89PNG\x0d\x0a\x1a\x0a")
		kocha.SendFile(c, "testrcname")
		actual := c.Response.ContentType
		expected := "image/png"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()
}

func TestContext_Redirect(t *testing.T) {
	c := newTestContext("testctrlr", "")
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
		kocha.Redirect(c, v.redirectURL, v.permanent).Proc(res)
		actual := []interface{}{w.Code, w.HeaderMap.Get("Location")}
		expected := []interface{}{v.expected, v.redirectURL}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`Controller.Redirect("%#v", %#v) => %#v; want %#v`, v.redirectURL, v.permanent, actual, expected)
		}
	}
}
