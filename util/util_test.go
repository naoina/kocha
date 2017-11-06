package util

import (
	"encoding/base64"
	"fmt"
	"go/build"
	"go/format"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"testing/quick"
)

type orderedOutputMap map[string]interface{}

func (m orderedOutputMap) String() string {
	keys := make([]string, 0, len(m))
	for key, _ := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	outputs := make([]string, 0, len(keys))
	for _, key := range keys {
		outputs = append(outputs, fmt.Sprintf("%s:%s", key, m[key]))
	}
	return fmt.Sprintf("map[%v]", strings.Join(outputs, " "))
}

func (m orderedOutputMap) GoString() string {
	keys := make([]string, 0, len(m))
	for key, _ := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i, key := range keys {
		keys[i] = fmt.Sprintf("%#v:%#v", key, m[key])
	}
	return fmt.Sprintf("map[string]interface{}{%v}", strings.Join(keys, ", "))
}

func Test_NormPath(t *testing.T) {
	for v, expected := range map[string]string{
		"/":           "/",
		"/path":       "/path",
		"/path/":      "/path/",
		"//path//":    "/path/",
		"/path/to":    "/path/to",
		"/path/to///": "/path/to/",
	} {
		actual := NormPath(v)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%v: Expect %v, but %v", v, expected, actual)
		}
	}
}

func TestGoString(t *testing.T) {
	re := regexp.MustCompile(`^/path/to/([^/]+)(?:\.html)?$`)
	actual := GoString(re)
	expected := `regexp.MustCompile("^/path/to/([^/]+)(?:\\.html)?$")`
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	tmpl := template.Must(template.New("test").Parse(`foo{{.name}}bar`))
	actual = GoString(tmpl)
	expected = `template.Must(template.New("test").Funcs(kocha.TemplateFuncs).Parse(util.Gunzip("\x1f\x8b\b\x00\x00\x00\x00\x00\x02\xffJ\xcbϯ\xae\xd6\xcbK\xccM\xad\xadMJ,\x02\x04\x00\x00\xff\xff4%\x83\xb6\x0f\x00\x00\x00")))`
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = GoString(testGoString{})
	expected = "gostring"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = GoString(nil)
	expected = "nil"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	var ptr *int
	actual = GoString(ptr)
	expected = "nil"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	n := 10
	nptr := &n
	actual = GoString(nptr)
	expected = "10"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	aBuf, err := format.Source([]byte(GoString(struct {
		Name, path string
		Route      orderedOutputMap
		G          *testGoString
	}{
		Name: "foo",
		path: "path",
		Route: orderedOutputMap{
			"first":  "Tokyo",
			"second": "Kyoto",
			"third":  []int{10, 11, 20},
		},
		G: &testGoString{},
	})))
	if err != nil {
		t.Fatal(err)
	}
	eBuf, err := format.Source([]byte(`
struct {
	Name string
	path string
	Route util.orderedOutputMap
	G *util.testGoString
}{

	G: gostring,

	Name: "foo",

	Route: map[string]interface{}{"first": "Tokyo", "second": "Kyoto", "third": []int{10, 11, 20}},
}`))
	if err != nil {
		t.Fatal(err)
	}
	actual = string(aBuf)
	expected = string(eBuf)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
	}
}

type testGoString struct{}

func (g testGoString) GoString() string { return "gostring" }

func Test_Gzip(t *testing.T) {
	actual := base64.StdEncoding.EncodeToString([]byte(Gzip("kocha")))
	expected := "H4sIAAAAAAAC/8rOT85IBAQAAP//DJOFlgUAAAA="
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	// reversibility test
	gz := Gzip("kocha")
	actual = Gunzip(gz)
	expected = "kocha"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestGunzip(t *testing.T) {
	actual := Gunzip("\x1f\x8b\b\x00\x00\tn\x88\x02\xff\xca\xceO\xceH\x04\x04\x00\x00\xff\xff\f\x93\x85\x96\x05\x00\x00\x00")
	expected := "kocha"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	// reversibility test
	raw := Gunzip("\x1f\x8b\b\x00\x00\tn\x88\x02\xff\xca\xceO\xceH\x04\x04\x00\x00\xff\xff\f\x93\x85\x96\x05\x00\x00\x00")
	actual = base64.StdEncoding.EncodeToString([]byte(Gzip(raw)))
	expected = "H4sIAAAAAAAC/8rOT85IBAQAAP//DJOFlgUAAAA="
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestFindAppDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "TestFindAppDir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	origGOPATH := build.Default.GOPATH
	defer func() {
		build.Default.GOPATH = origGOPATH
		os.Setenv("GOPATH", origGOPATH)
	}()
	build.Default.GOPATH = tempDir + string(filepath.ListSeparator) + build.Default.GOPATH
	os.Setenv("GOPATH", build.Default.GOPATH)
	myappPath := filepath.Join(tempDir, "src", "path", "to", "myapp")
	if err := os.MkdirAll(myappPath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(myappPath); err != nil {
		t.Fatal(err)
	}
	actual, err := FindAppDir()
	if err != nil {
		t.Fatal(err)
	}
	expected := "path/to/myapp"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("FindAppDir() => %q, want %q", actual, expected)
	}
}

func TestIsUnexportedField(t *testing.T) {
	// test for bug case older than Go1.3.
	func() {
		type b struct{}
		type C struct {
			b
		}
		v := reflect.TypeOf(C{}).Field(0)
		actual := IsUnexportedField(v)
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("IsUnexportedField(%#v) => %v, want %v", v, actual, expected)
		}
	}()

	// test for correct case.
	func() {
		type B struct{}
		type C struct {
			B
		}
		v := reflect.TypeOf(C{}).Field(0)
		actual := IsUnexportedField(v)
		expected := false
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("IsUnexportedField(%#v) => %v, want %v", v, actual, expected)
		}
	}()
}

func TestGenerateRandomKey(t *testing.T) {
	if err := quick.Check(func(length uint16) bool {
		already := make([][]byte, 0, 100)
		for i := 0; i < 100; i++ {
			buf := GenerateRandomKey(int(length))
			for _, v := range already {
				if !reflect.DeepEqual(buf, v) {
					return false
				}
			}
		}
		return true
	}, &quick.Config{MaxCount: 10}); err != nil {
		t.Error(err)
	}
}
