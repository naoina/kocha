package kocha_test

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/naoina/kocha"
)

func TestTemplate_FuncMap_in_withInvalidType(t *testing.T) {
	app := kocha.NewTestApp()
	funcMap := template.FuncMap(app.Template.FuncMap)
	tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{in 1 1}}`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err == nil {
		t.Errorf("Expect errors, but no errors")
	}
}

func TestTemplate_FuncMap_in(t *testing.T) {
	app := kocha.NewTestApp()
	funcMap := template.FuncMap(app.Template.FuncMap)
	var buf bytes.Buffer
	for _, v := range []struct {
		Arr    interface{}
		Sep    interface{}
		expect string
		err    error
	}{
		{[]string{"b", "a", "c"}, "a", "true", nil},
		{[]string{"ab", "b", "c"}, "a", "false", nil},
		{nil, "a", "", fmt.Errorf("valid types are slice, array and string, got `invalid'")},
	} {
		buf.Reset()
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{in .Arr .Sep}}`))
		err := tmpl.Execute(&buf, v)
		if !strings.HasSuffix(fmt.Sprint(err), fmt.Sprint(v.err)) {
			t.Errorf(`{{in %#v %#v}}; error has "%v"; want "%v"`, v.Arr, v.Sep, err, v.err)
		}
		actual := buf.String()
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`{{in %#v %#v}} => %#v; want %#v`, v.Arr, v.Sep, actual, expect)
		}
	}
}

func TestTemplate_FuncMap_url(t *testing.T) {
	app := kocha.NewTestApp()
	funcMap := template.FuncMap(app.Template.FuncMap)

	func() {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{url "root"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			panic(err)
		}
		actual := buf.String()
		expected := "/"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	func() {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{url "user" 713}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			panic(err)
		}
		actual := buf.String()
		expected := "/user/713"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()
}

func TestTemplate_FuncMap_nl2br(t *testing.T) {
	app := kocha.NewTestApp()
	funcMap := template.FuncMap(app.Template.FuncMap)
	tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{nl2br "a\nb\nc\n"}}`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := "a<br>b<br>c<br>"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

func TestTemplate_FuncMap_raw(t *testing.T) {
	app := kocha.NewTestApp()
	funcMap := template.FuncMap(app.Template.FuncMap)
	tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{raw "\n<br>"}}`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := "\n<br>"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

func TestTemplate_FuncMap_invokeTemplate(t *testing.T) {
	// test that if ActiveIf returns true.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		c := &kocha.Context{
			Data: map[interface{}]interface{}{
				"unit": &testUnit{"test1", true, 0},
				"ctx":  "testctx1",
			},
		}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template .Data.unit "test_tmpl1" "def_tmpl" $}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "test_tmpl1: testctx1\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that if ActiveIf returns false.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		c := &kocha.Context{
			Data: map[interface{}]interface{}{
				"unit": &testUnit{"test2", false, 0},
				"ctx":  "testctx2",
			},
		}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template .Data.unit "test_tmpl1" "def_tmpl" $}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "def_tmpl: testctx2\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that unknown template.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		c := &kocha.Context{
			Data: map[interface{}]interface{}{
				"unit": &testUnit{"test3", true, 0},
				"ctx":  "testctx3",
			},
		}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template .Data.unit "unknown_tmpl" "def_tmpl" $}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "def_tmpl: testctx3\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that unknown templates.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		c := &kocha.Context{
			Data: map[interface{}]interface{}{
				"unit": &testUnit{"test4", true, 0},
				"ctx":  "testctx4",
			},
		}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template .Data.unit "unknown_tmpl" "unknown_def_tmpl" $}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); !strings.HasSuffix(err.Error(), "template not found: appname:/unknown_def_tmpl.html") {
			t.Error(err)
		}
	}()

	// test that unknown default template.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		c := &kocha.Context{
			Data: map[interface{}]interface{}{
				"unit": &testUnit{"test5", true, 0},
				"ctx":  "testctx5",
			},
		}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template .Data.unit "test_tmpl1" "unknown" $}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); err != nil {
			t.Error(err)
		}
	}()

	// test that single context.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		c := &kocha.Context{
			Data: map[interface{}]interface{}{
				"unit": &testUnit{"test6", true, 0},
				"ctx":  "testctx6",
			},
		}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template .Data.unit "test_tmpl1" "def_tmpl" $}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "test_tmpl1: testctx6\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that too many contexts.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		c := &kocha.Context{
			Data: map[interface{}]interface{}{
				"unit": &testUnit{"test7", true, 0},
				"ctx":  "testctx7",
			},
		}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template .Data.unit "test_tmpl1" "def_tmpl" $ $}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); !strings.HasSuffix(err.Error(), "number of context must be 0 or 1") {
			t.Error(err)
		}
	}()
}

func TestTemplateFuncMap_flash(t *testing.T) {
	c := newTestContext("testctrlr", "")
	funcMap := template.FuncMap(c.App.Template.FuncMap)
	for _, v := range []struct {
		key    string
		expect string
	}{
		{"", ""},
		{"success", "test succeeded"},
		{"success", "test successful"},
		{"error", "test failed"},
		{"error", "test failure"},
	} {
		c.Flash = kocha.Flash{}
		c.Flash.Set(v.key, v.expect)
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(fmt.Sprintf(`{{flash . "unknown"}}{{flash . "%s"}}`, v.key)))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, c); err != nil {
			t.Error(err)
			continue
		}
		actual := buf.String()
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`{{flash . %#v}} => %#v; want %#v`, v.key, actual, expect)
		}
	}
}

func TestTemplate_Get(t *testing.T) {
	app := kocha.NewTestApp()
	func() {
		for _, v := range []struct {
			appName   string
			layout    string
			ctrlrName string
			format    string
		}{
			{"appname", "application", "testctrlr", "html"},
			{"appname", "", "testctrlr", "js"},
			{"appname", "another_layout", "testctrlr", "html"},
		} {
			tmpl, err := app.Template.Get(v.appName, v.layout, v.ctrlrName, v.format)
			var actual interface{} = err
			var expect interface{} = nil
			if !reflect.DeepEqual(actual, expect) {
				t.Fatalf(`Template.Get(%#v, %#v, %#v, %#v) => %T, %#v, want *template.Template, %#v`, v.appName, v.layout, v.ctrlrName, v.format, tmpl, actual, expect)
			}
		}
	}()

	func() {
		for _, v := range []struct {
			appName   string
			layout    string
			ctrlrName string
			format    string
			expectErr error
		}{
			{"unknownAppName", "app", "test_tmpl1", "html", fmt.Errorf("kocha: template not found: unknownAppName:app/test_tmpl1.html")},
			{"testAppName", "app", "unknown_tmpl1", "html", fmt.Errorf("kocha: template not found: testAppName:app/unknown_tmpl1.html")},
			{"testAppName", "app", "test_tmpl1", "xml", fmt.Errorf("kocha: template not found: testAppName:app/test_tmpl1.xml")},
			{"testAppName", "", "test_tmpl1", "xml", fmt.Errorf("kocha: template not found: testAppName:/test_tmpl1.xml")},
		} {
			tmpl, err := app.Template.Get(v.appName, v.layout, v.ctrlrName, v.format)
			actual := tmpl
			expect := (*template.Template)(nil)
			actualErr := err
			expectErr := v.expectErr
			if !reflect.DeepEqual(actual, expect) || !reflect.DeepEqual(actualErr, expectErr) {
				t.Errorf(`Template.Get(%#v, %#v, %#v, %#v) => %#v, %#v, ; want %#v, %#v`, v.appName, v.layout, v.ctrlrName, v.format, actual, actualErr, expect, expectErr)
			}
		}
	}()
}

func TestTemplateDelims(t *testing.T) {
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
			LeftDelim:  "{%",
			RightDelim: "%}",
		},
		RouteTable: []*kocha.Route{
			{
				Name: "root",
				Path: "/",
				Controller: &kocha.FixtureAnotherDelimsTestCtrl{
					Ctx: "test_other_delims_ctx",
				},
			},
		},
		Middlewares: []kocha.Middleware{
			&kocha.DispatchMiddleware{},
		},
		Logger: &kocha.LoggerConfig{
			Writer: ioutil.Discard,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	var actual interface{} = w.Code
	var expect interface{} = 200
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`GET / status => %#v; want %#v`, actual, expect)
	}
	actual = w.Body.String()
	expect = "This is other delims: test_other_delims_ctx\n"
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`GET / => %#v; want %#v`, actual, expect)
	}
}
