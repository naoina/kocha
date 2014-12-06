package kocha_test

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
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
	tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{in . "a"}}`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, []string{"b", "a", "c"}); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := "true"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}

	buf.Reset()
	if err := tmpl.Execute(&buf, []string{"ab", "b", "c"}); err != nil {
		panic(err)
	}
	actual = buf.String()
	expected = "false"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
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
		unit := &testUnit{"test1", true, 0}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "test_tmpl1: \n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that if ActiveIf returns false.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		unit := &testUnit{"test2", false, 0}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "def_tmpl: \n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that unknown template.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		unit := &testUnit{"test3", true, 0}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template . "unknown_tmpl" "def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "def_tmpl: \n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that unknown templates.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		unit := &testUnit{"test4", true, 0}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template . "unknown_tmpl" "unknown_def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err == nil {
			t.Errorf("no error returned by unknown template")
		}
	}()

	// test that unknown default template.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		unit := &testUnit{"test5", true, 0}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template . "test_tmpl1" "unknown"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Errorf("no error returned by unknown default template")
		}
	}()

	// test that single context.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		unit := &testUnit{"test6", true, 0}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl" "ctx"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Error(err)
		}
		actual := buf.String()
		expected := "test_tmpl1: ctx\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()

	// test that too many contexts.
	func() {
		app := kocha.NewTestApp()
		funcMap := template.FuncMap(app.Template.FuncMap)
		unit := &testUnit{"test7", true, 0}
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl" "ctx" "over"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err == nil {
			t.Errorf("no error returned by too many number of context")
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
