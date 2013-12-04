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
		"testAppName": {
			"app": {
				"html": {
					"test_tmpl1": template.Must(template.New("test_tmpl1").Parse(``)),
				},
				"js": {
					"test_tmpl1": template.Must(template.New("test_tmpl1").Parse(``)),
				},
			},
			"anotherLayout": {
				"html": {
					"test_tmpl1": template.Must(template.New("test_tmpl1").Parse(``)),
				},
			},
		},
	}
	actual := templateSet.Get("testAppName", "app", "test_tmpl1", "html")
	expected := templateSet["testAppName"]["app"]["html"]["test_tmpl1"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateSet.Get("testAppName", "app", "test_tmpl1", "js")
	expected = templateSet["testAppName"]["app"]["js"]["test_tmpl1"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateSet.Get("testAppName", "anotherLayout", "test_tmpl1", "html")
	expected = templateSet["testAppName"]["anotherLayout"]["html"]["test_tmpl1"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateSet.Get("unknownAppName", "app", "test_tmpl1", "html")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}

	actual = templateSet.Get("testAppName", "app", "unknown_tmpl1", "html")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}

	actual = templateSet.Get("testAppName", "app", "test_tmpl1", "xml")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}
}

func TestTemplateSet_Ident(t *testing.T) {
	templateSet := TemplateSet{}
	for expected, args := range map[string][]string{
		"a:b c.html":   []string{"a", "b", "c", "html"},
		"b:a c.html":   []string{"b", "a", "c", "html"},
		"a:b c.js":     []string{"a", "b", "c", "js"},
		"a:b c_d.html": []string{"a", "b", "cD", "html"},
		"a:b c_d_e.js": []string{"a", "b", "CDE", "js"},
	} {
		actual := templateSet.Ident(args[0], args[1], args[2], args[3])
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
	expected := map[string]map[string]map[string]map[string]string{
		"appName": {
			"": {
				"html": {
					"fixture_date_test_ctrl":   "\nsingle date\n",
					"fixture_root_test_ctrl":   "\nsingle root\n",
					"fixture_user_test_ctrl":   "\nsingle user\n",
					"fixture_teapot_test_ctrl": "\nsingle tea pot\n",
					"errors/500":               "\nsingle 500 error\n",
				},
				"json": {
					"fixture_teapot_test_ctrl": "\n{\"single\":\"tea pot\"}\n",
				},
			},
			"application": {
				"html": {
					"fixture_date_test_ctrl":   "This is layout\n\nThis is date\n\n",
					"fixture_root_test_ctrl":   "This is layout\n\nThis is root\n\n",
					"fixture_user_test_ctrl":   "This is layout\n\nThis is user\n\n",
					"fixture_teapot_test_ctrl": "This is layout\n\nI'm a tea pot\n\n",
					"errors/500":               "This is layout\n\n500 error\n\n",
				},
				"json": {
					"fixture_teapot_test_ctrl": "{\n  \"layout\": \"application\",\n  \n{\"status\":418, \"text\":\"I'm a tea pot\"}\n\n}\n",
				},
			},
			"sub": {
				"html": {
					"fixture_date_test_ctrl":   "This is sub\n\nThis is date\n\n",
					"fixture_root_test_ctrl":   "This is sub\n\nThis is root\n\n",
					"fixture_user_test_ctrl":   "This is sub\n\nThis is user\n\n",
					"fixture_teapot_test_ctrl": "This is sub\n\nI'm a tea pot\n\n",
					"errors/500":               "This is sub\n\n500 error\n\n",
				},
			},
		},
	}
	for appName, actualAppTemplateSet := range actual {
		appTemplateSet, ok := expected[appName]
		if !ok {
			t.Errorf("appName %v is unexpected", appName)
			continue
		}
		for layoutName, actualLayoutTemplateSet := range actualAppTemplateSet {
			layoutTemplates, ok := appTemplateSet[layoutName]
			if !ok {
				t.Errorf("layout %v is unexpected", layoutName)
				continue
			}
			for ext, actualFileExtTemplateSet := range actualLayoutTemplateSet {
				templates, ok := layoutTemplates[ext]
				if !ok {
					t.Errorf("ext %v is unexpected", ext)
					continue
				}
				for name, actualTemplate := range actualFileExtTemplateSet {
					expectedContent, ok := templates[name]
					if !ok {
						t.Errorf("name %v is unexpected", name)
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
	}
}
