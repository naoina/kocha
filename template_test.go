package kocha_test

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"sort"
	"testing"
	"time"

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

func TestTemplate_FuncMap_date(t *testing.T) {
	app := kocha.NewTestApp()
	funcMap := template.FuncMap(app.Template.FuncMap)
	base := `{{date . "%v"}}`
	now := time.Now()
	tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(fmt.Sprintf(base, "2006/01/02 15:04:05.999999999")))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, now); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := now.Format("2006/01/02 15:04:05.999999999")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}

	tmpl = template.Must(template.New("test").Funcs(funcMap).Parse(fmt.Sprintf(base, "Jan 02 2006 03:04.999999999")))
	buf.Reset()
	if err := tmpl.Execute(&buf, now); err != nil {
		panic(err)
	}
	actual = buf.String()
	expected = now.Format("Jan 02 2006 03:04.999999999")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

func TestTemplate_Get(t *testing.T) {
	app := kocha.NewTestApp()
	func() {
		appname, a, ctrl, typ := "appname", "application", "testctrlr", "html"
		tmpl := app.Template.Get(appname, a, ctrl, typ)
		if tmpl == nil {
			t.Fatalf(`Template.Get(%#v, %#v, %#v, %#v) => nil, want *template.Template`, appname, a, ctrl, typ)
		}
		var actual []string
		for _, v := range tmpl.Templates() {
			actual = append(actual, v.Name())
		}
		expected := []string{"layout", "testctrlr.html"}
		sort.Strings(actual)
		sort.Strings(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	func() {
		appname, a, ctrl, typ := "appname", "", "testctrlr", "js"
		tmpl := app.Template.Get(appname, a, ctrl, typ)
		if tmpl == nil {
			t.Fatalf(`Template.Get(%#v, %#v, %#v, %#v) => nil, want *template.Template`, appname, a, ctrl, typ)
		}
		var actual []string
		for _, v := range tmpl.Templates() {
			actual = append(actual, v.Name())
		}
		expected := []string{"testctrlr"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	func() {
		appname, a, ctrl, typ := "appname", "another_layout", "testctrlr", "html"
		tmpl := app.Template.Get(appname, a, ctrl, typ)
		if tmpl == nil {
			t.Fatalf(`Template.Get(%#v, %#v, %#v, %#v) => nil, want *template.Template`, appname, a, ctrl, typ)
		}
		var actual []string
		for _, v := range tmpl.Templates() {
			actual = append(actual, v.Name())
		}
		expected := []string{"layout", "testctrlr.html"}
		sort.Strings(actual)
		sort.Strings(expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	func() {
		actual := app.Template.Get("unknownAppName", "app", "test_tmpl1", "html")
		if actual != nil {
			t.Errorf("Expect %v, but %v", nil, actual)
		}
	}()

	func() {
		actual := app.Template.Get("testAppName", "app", "unknown_tmpl1", "html")
		if actual != nil {
			t.Errorf("Expect %v, but %v", nil, actual)
		}
	}()

	func() {
		actual := app.Template.Get("testAppName", "app", "test_tmpl1", "xml")
		if actual != nil {
			t.Errorf("Expect %v, but %v", nil, actual)
		}
	}()
}

func TestTemplate_Ident(t *testing.T) {
	app := kocha.NewTestApp()
	for expected, args := range map[string][]string{
		"a:b c.html":   []string{"a", "b", "c", "html"},
		"b:a c.html":   []string{"b", "a", "c", "html"},
		"a:b c.js":     []string{"a", "b", "c", "js"},
		"a:b c_d.html": []string{"a", "b", "cD", "html"},
		"a:b c_d_e.js": []string{"a", "b", "CDE", "js"},
	} {
		actual := app.Template.Ident(args[0], args[1], args[2], args[3])
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}
