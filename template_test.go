package kocha

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

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

func TestTemplateFuncs_invoke_template(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	appConfig.templateMap = TemplateMap{
		appConfig.AppName: {
			"": {
				"html": {
					"def_tmpl":   template.Must(template.New("def_tmpl").Parse(`<div>def_tmpl:{{.}}</div>`)),
					"test_tmpl1": template.Must(template.New("test_tmpl1").Parse(`<div>test_tmpl1:{{.}}</div>`)),
					"test_tmpl2": template.Must(template.New("test_tmpl2").Parse(`<div>test_tmpl2:{{.}}</div>`)),
				},
			},
		},
	}

	// test that if ActiveIf returns true.
	testInvokeWrapper(func() {
		unit := &testUnit{"test1", true, 0}
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Fatal(err)
		}
		actual := buf.String()
		expected := "<div>test_tmpl1:</div>"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	// test that if ActiveIf returns false.
	testInvokeWrapper(func() {
		unit := &testUnit{"test2", false, 0}
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Fatal(err)
		}
		actual := buf.String()
		expected := "<div>def_tmpl:</div>"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	// test that unknown template.
	testInvokeWrapper(func() {
		unit := &testUnit{"test3", true, 0}
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{invoke_template . "unknown_tmpl" "def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Fatal(err)
		}
		actual := buf.String()
		expected := "<div>def_tmpl:</div>"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	// test that unknown templates.
	testInvokeWrapper(func() {
		unit := &testUnit{"test4", true, 0}
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{invoke_template . "unknown_tmpl" "unknown_def_tmpl"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err == nil {
			t.Errorf("no error returned by unknown template")
		}
	})

	// test that unknown default template.
	testInvokeWrapper(func() {
		unit := &testUnit{"test5", true, 0}
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{invoke_template . "test_tmpl1" "unknown"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Errorf("no error returned by unknown default template")
		}
	})

	// test that single context.
	testInvokeWrapper(func() {
		unit := &testUnit{"test6", true, 0}
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl" "ctx"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err != nil {
			t.Fatal(err)
		}
		actual := buf.String()
		expected := "<div>test_tmpl1:ctx</div>"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	// test that too many contexts.
	testInvokeWrapper(func() {
		unit := &testUnit{"test7", true, 0}
		tmpl := template.Must(template.New("test").Funcs(TemplateFuncs).Parse(`{{invoke_template . "test_tmpl1" "def_tmpl" "ctx" "over"}}`))
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, unit); err == nil {
			t.Errorf("no error returned by too many number of context")
		}
	})
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

func TestTemplateMap_Get(t *testing.T) {
	templateMap := TemplateMap{
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
	actual := templateMap.Get("testAppName", "app", "test_tmpl1", "html")
	expected := templateMap["testAppName"]["app"]["html"]["test_tmpl1"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateMap.Get("testAppName", "app", "test_tmpl1", "js")
	expected = templateMap["testAppName"]["app"]["js"]["test_tmpl1"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateMap.Get("testAppName", "anotherLayout", "test_tmpl1", "html")
	expected = templateMap["testAppName"]["anotherLayout"]["html"]["test_tmpl1"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = templateMap.Get("unknownAppName", "app", "test_tmpl1", "html")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}

	actual = templateMap.Get("testAppName", "app", "unknown_tmpl1", "html")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}

	actual = templateMap.Get("testAppName", "app", "test_tmpl1", "xml")
	if actual != nil {
		t.Errorf("Expect %v, but %v", nil, actual)
	}
}

func TestTemplateMap_Ident(t *testing.T) {
	templateMap := TemplateMap{}
	for expected, args := range map[string][]string{
		"a:b c.html":   []string{"a", "b", "c", "html"},
		"b:a c.html":   []string{"b", "a", "c", "html"},
		"a:b c.js":     []string{"a", "b", "c", "js"},
		"a:b c_d.html": []string{"a", "b", "cD", "html"},
		"a:b c_d_e.js": []string{"a", "b", "CDE", "js"},
	} {
		actual := templateMap.Ident(args[0], args[1], args[2], args[3])
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}

func TestTemplateSet_buildTemplateMap(t *testing.T) {
	ts := TemplateSet{
		{
			Name: "appName",
			Paths: []string{
				filepath.Join("testdata", "app", "views"),
			},
		},
	}
	actual, err := ts.buildTemplateMap()
	if err != nil {
		t.Fatal(err)
	}
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
