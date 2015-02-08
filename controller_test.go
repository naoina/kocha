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
		data := kocha.OrderedOutputMap{
			"c1": "v1",
			"c2": "v2",
		}
		w := httptest.NewRecorder()
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.Render(data); err != nil {
			t.Fatal(err)
		}
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expect1 := "tmpl_ctx: map[c1:v1 c2:v2]\n"
		expect2 := "tmpl_ctx: map[c2:v2 c1:v1]\n"
		if !reflect.DeepEqual(actual, expect1) && !reflect.DeepEqual(actual, expect2) {
			t.Errorf(`c.Render(%#v) => %#v; want %#v or %#v`, data, actual, expect1, expect2)
		}
		if !reflect.DeepEqual(c.Response.ContentType, "text/html") {
			t.Errorf("Expect %v, but %v", "text/html", c.Response.ContentType)
		}
	}()

	func() {
		c := newTestContext("testctrlr_ctx", "")
		c.Data = kocha.OrderedOutputMap{
			"c3": "v3",
			"c4": "v4",
		}
		w := httptest.NewRecorder()
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.Render(nil); err != nil {
			t.Fatal(err)
		}
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
		c.Data = kocha.OrderedOutputMap{
			"c5": "v5",
			"c6": "v6",
		}
		ctx := kocha.OrderedOutputMap{
			"c6": "test",
			"c7": "v7",
		}
		w := httptest.NewRecorder()
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.Render(ctx); err != nil {
			t.Fatal(err)
		}
		buf, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		actual := string(buf)
		expected := fmt.Sprintf("tmpl_ctx: %v\n", kocha.OrderedOutputMap{
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
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.Render(ctx); err != nil {
			t.Fatal(err)
		}
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
		c.Data = map[interface{}]interface{}{"c1": "v1"}
		ctx := "test_ctx_override"
		w := httptest.NewRecorder()
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.Render(ctx); err != nil {
			t.Fatal(err)
		}
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
		c.Data = map[interface{}]interface{}{"c1": "v1"}
		if err := c.Render(nil); err != nil {
			t.Fatal(err)
		}
		actual := c.Data
		expected := map[interface{}]interface{}{
			"c1": "v1",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Context.Data => %#v, want %#v", actual, expected)
		}
	}()

	func() {
		c := newTestContext("testctrlr_ctx", "")
		ctx := map[interface{}]interface{}{"c1": "v1"}
		if err := c.Render(ctx); err != nil {
			t.Fatal(err)
		}
		actual := c.Data
		expected := map[interface{}]interface{}{
			"c1": "v1",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Context.Data => %#v, want %#v", actual, expected)
		}
	}()
}

func TestContext_Render_withDifferentKeyType(t *testing.T) {
	for _, v := range []struct {
		data   interface{}
		ctx    interface{}
		expect error
	}{
		{map[interface{}]interface{}{"c1": "v1"}, map[string]interface{}{"c2": "v2"}, nil},
		{map[string]interface{}{"c1": "v1"}, map[interface{}]interface{}{"c2": "v2"}, fmt.Errorf("kocha: context: key of type interface {} is not assignable to type string")},
		{map[int]interface{}{1: "v1"}, map[string]interface{}{"2": "v2"}, fmt.Errorf("kocha: context: key of type string is not assignable to type int")},
		{map[string]string{"c1": "v1"}, map[string]interface{}{"c2": "v2"}, fmt.Errorf("kocha: context: value of type interface {} is not assignable to type string")},
		{map[string]int{"c1": 1}, map[string]string{"c2": "v2"}, fmt.Errorf("kocha: context: value of type string is not assignable to type int")},
		{map[string]string{"c1": "v1"}, map[string][]byte{"c2": []byte("v2")}, nil},
		{map[string]interface{}{"c1": "v1"}, map[string]int{"c2": 2}, nil},
	} {
		c := newTestContext("testctrlr_ctx", "")
		c.Data = v.data
		ctx := v.ctx
		actual := c.Render(ctx)
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`c.Render(%#v) => %#v; want %#v`, ctx, actual, expect)
		}
	}
}

func TestContext_Render_withContentType(t *testing.T) {
	c := newTestContext("testctrlr", "")
	w := httptest.NewRecorder()
	c.Response.ResponseWriter = w
	c.Response.ContentType = "application/json"
	if err := c.Render(nil); err != nil {
		t.Fatal(err)
	}
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
	c := newTestContext("testctrlr", "")
	c.App.Config.AppName = "unknownAppName"
	actual := c.Render(nil)
	expect := fmt.Errorf("kocha: template not found: unknownAppName:/testctrlr.html")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`kocha.Render(%#v, %#v) => %#v; want %#v`, c, nil, actual, expect)
	}
}

func TestContext_Render_withMissingTemplate(t *testing.T) {
	c := newTestContext("testctrlr", "")
	c.Name = "unknownctrlr"
	actual := c.Render(nil)
	expect := fmt.Errorf("kocha: template not found: appname:/unknownctrlr.html")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`kocha.Render(%#v, %#v) => %#v; want %#v`, c, nil, actual, expect)
	}
}

func TestContext_Render_withAnotherLayout(t *testing.T) {
	c := newTestContext("testctrlr", "another_layout")
	w := httptest.NewRecorder()
	c.Response = &kocha.Response{ResponseWriter: w}
	if err := c.Render(nil); err != nil {
		t.Fatal(err)
	}
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
	c.Response = &kocha.Response{ResponseWriter: w}
	if err := c.RenderJSON(struct{ A, B string }{"hoge", "foo"}); err != nil {
		t.Fatal(err)
	}
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
	c.Response = &kocha.Response{ResponseWriter: w}
	if err := c.RenderXML(ctx); err != nil {
		t.Fatal(err)
	}
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
	c.Response = &kocha.Response{ResponseWriter: w}
	if err := c.RenderText("test_content_data"); err != nil {
		t.Fatal(err)
	}
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
	c.Response = &kocha.Response{ResponseWriter: w}
	if err := c.RenderError(http.StatusInternalServerError, nil); err != nil {
		t.Fatal(err)
	}
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
	c.Response = &kocha.Response{ResponseWriter: w}
	if err := c.RenderError(http.StatusBadRequest, nil); err != nil {
		t.Fatal(err)
	}
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
	w = httptest.NewRecorder()
	c.Response.ResponseWriter = w
	c.Response.ContentType = "application/json"
	if err := c.RenderError(http.StatusInternalServerError, nil); err != nil {
		t.Fatal(err)
	}
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
		c.Response.ContentType = "unknown/content-type"
		actual := c.RenderError(http.StatusInternalServerError, nil)
		expect := fmt.Errorf("kocha: unknown Content-Type: unknown/content-type")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`kocha.RenderError(%#v, %#v, %#v) => %#v; want %#v`, c, http.StatusInternalServerError, nil, actual, expect)
		}
	}()

	c = newTestContext("testctrlr", "")
	w = httptest.NewRecorder()
	c.Response = &kocha.Response{ResponseWriter: w}
	if err := c.RenderError(http.StatusTeapot, nil); err != nil {
		t.Fatal(err)
	}
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
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.SendFile(tmpFile.Name()); err != nil {
			t.Fatal(err)
		}
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
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.SendFile(filepath.Base(tmpFile.Name())); err != nil {
			t.Fatal(err)
		}
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
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.SendFile("unknown/path"); err != nil {
			t.Fatal(err)
		}
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
		if err := c.SendFile(tmpFile.Name()); err != nil {
			t.Fatal(err)
		}
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
		if err := c.SendFile(tmpFile.Name()); err != nil {
			t.Fatal(err)
		}
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
		if err := c.SendFile(tmpFile.Name()); err != nil {
			t.Fatal(err)
		}
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
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.SendFile("testrcname"); err != nil {
			t.Fatal(err)
		}
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
		if err := c.SendFile("testrcname"); err != nil {
			t.Fatal(err)
		}
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
		c.Response = &kocha.Response{ResponseWriter: w}
		if err := c.Redirect(v.redirectURL, v.permanent); err != nil {
			t.Fatal(err)
		}
		actual := []interface{}{w.Code, w.HeaderMap.Get("Location")}
		expected := []interface{}{v.expected, v.redirectURL}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`Controller.Redirect("%#v", %#v) => %#v; want %#v`, v.redirectURL, v.permanent, actual, expected)
		}
	}
}
