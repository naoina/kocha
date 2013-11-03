package kocha

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

func TestTemplateFuncs_eq(t *testing.T) {
	base := `{{eq "%v" "%v"}}`
	if err := quick.Check(func(x string) bool {
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(fmt.Sprintf(base, x, x)))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			panic(err)
		}
		return buf.String() == "true"
	}, nil); err != nil {
		t.Error(err)
	}

	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(fmt.Sprintf(base, "a", "b")))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := "false"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

func TestTemplateFuncs_ne(t *testing.T) {
	base := `{{ne "%v" "%v"}}`
	if err := quick.Check(func(x string) bool {
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(fmt.Sprintf(base, x, x)))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			panic(err)
		}
		return buf.String() == "false"
	}, nil); err != nil {
		t.Error(err)
	}

	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(fmt.Sprintf(base, "a", "b")))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := "true"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

func TestTemplateFuncs_in_with_invalid_type(t *testing.T) {
	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{in 1 1}}`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err == nil {
		t.Errorf("Expect errors, but no errors")
	}
}

func TestTemplateFuncs_in(t *testing.T) {
	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{in . "a"}}`))
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

func TestTemplateFuncs_url(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{url "root"}}`))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := "/"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}

	tmpl = template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{url "user" 713}}`))
	buf.Reset()
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic(err)
	}
	actual = buf.String()
	expected = "/user/713"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestTemplateFuncs_nl2br(t *testing.T) {
	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{nl2br "a\nb\nc\n"}}`))
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

func TestTemplateFuncs_raw(t *testing.T) {
	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{raw "\n<br>"}}`))
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

func TestTemplateFuncs_date(t *testing.T) {
	base := `{{date . "%v"}}`
	now := time.Now()
	tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(fmt.Sprintf(base, "2006/01/02 15:04:05.999999999")))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, now); err != nil {
		panic(err)
	}
	actual := buf.String()
	expected := now.Format("2006/01/02 15:04:05.999999999")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}

	tmpl = template.Must(template.New("test").Funcs(TemplateFuncs).Parse(fmt.Sprintf(base, "Jan 02 2006 03:04.999999999")))
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

func TestTemplateSet_Get(t *testing.T) {
	templateSet := TemplateSet{
		"testAppName": map[string]*template.Template{
			"test_tmpl1.html": template.Must(template.New("test_tmpl1").Parse(``)),
			"test_tmpl1.js":   template.Must(template.New("test_tmpl1").Parse(``)),
		},
	}
	actual := templateSet.Get("testAppName", "test_tmpl1", "html")
	expected := templateSet["testAppName"]["test_tmpl1.html"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateSet.Get("testAppName", "test_tmpl1", "js")
	expected = templateSet["testAppName"]["test_tmpl1.js"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateSet.Get("unknownAppName", "test_tmpl1", "html")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}

	actual = templateSet.Get("testAppName", "unknown_tmpl1", "html")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}

	actual = templateSet.Get("testAppName", "test_tmpl1", "xml")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}
}

func TestTemplateSet_Ident(t *testing.T) {
	templateSet := TemplateSet{}
	for expected, args := range map[string][]string{
		"a:b.html":   []string{"a", "b", "html"},
		"b:a.html":   []string{"b", "a", "html"},
		"a:b.js":     []string{"a", "b", "js"},
		"a:b_c.html": []string{"a", "bC", "html"},
		"a:b_c_d.js": []string{"a", "BCD", "js"},
	} {
		actual := templateSet.Ident(args[0], args[1], args[2])
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}

func TestTemplateSetFromPaths(t *testing.T) {
	actual := TemplateSetFromPaths(map[string][]string{
		"appName": []string{
			filepath.Join("testdata", "app", "views"),
		},
	})
	expected := map[string]map[string]string{
		"appName": map[string]string{
			"fixture_date_test_ctrl.html":   "This is layout\n\nThis is date\n\n",
			"fixture_root_test_ctrl.html":   "This is layout\n\nThis is root\n\n",
			"fixture_user_test_ctrl.html":   "This is layout\n\nThis is user\n\n",
			"fixture_teapot_test_ctrl.html": "This is layout\n\nI'm a tea pot\n\n",
			"errors/500.html":               "This is layout\n\n500 error\n\n",
		},
	}
	for appName, actualMap := range actual {
		tmpls, ok := expected[appName]
		if !ok {
			t.Errorf("%v is unexpected", appName)
			continue
		}
		for name, actualTemplate := range actualMap {
			expectedContent, ok := tmpls[name]
			if !ok {
				t.Errorf("%v is unexpected", name)
				continue
			}
			var actualContent bytes.Buffer
			if err := actualTemplate.Execute(&actualContent, nil); err != nil {
				t.Error(err)
				continue
			}
			if !reflect.DeepEqual(actualContent.String(), expectedContent) {
				t.Errorf("Expect %q, but %q", expectedContent, actualContent.String())
			}
		}
	}
}
