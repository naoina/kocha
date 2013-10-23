package kocha

import (
	"bytes"
	"html/template"
	"path/filepath"
	"reflect"
	"testing"
)

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
			"fixture_date_test_ctrl.html": "This is layout\n\nThis is date\n\n",
			"fixture_root_test_ctrl.html": "This is layout\n\nThis is root\n\n",
			"fixture_user_test_ctrl.html": "This is layout\n\nThis is user\n\n",
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
