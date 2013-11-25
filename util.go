package kocha

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var (
	// Now returns current time. This is for mock in tests.
	Now = func() time.Time { return time.Now() }
)

// ToCamelCase returns a copy of the string s with all Unicode letters mapped to their camel case.
// It will convert to upper case previous letter of '_' and first letter, and remove letter of '_'.
func ToCamelCase(s string) string {
	result := make([]rune, 0, len(s))
	upper := false
	for _, r := range s {
		if r == '_' {
			upper = true
			continue
		}
		if upper {
			result = append(result, unicode.ToUpper(r))
			upper = false
			continue
		}
		result = append(result, r)
	}
	result[0] = unicode.ToUpper(result[0])
	return string(result)
}

// ToSnakeCase returns a copy of the string s with all Unicode letters mapped to their snake case.
// It will insert letter of '_' at position of previous letter of uppercase and all
// letters convert to lower case.
func ToSnakeCase(s string) string {
	var result bytes.Buffer
	result.WriteRune(unicode.ToLower(rune(s[0])))
	for _, c := range s[1:] {
		if unicode.IsUpper(c) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(c))
	}
	return result.String()
}

func normPath(p string) string {
	result := path.Clean(p)
	// path.Clean() truncate the trailing slash but add it.
	if p[len(p)-1] == '/' && result != "/" {
		result += "/"
	}
	return result
}

type Error struct {
	Usager  usager
	Message string
}

func (e Error) Error() string {
	return e.Message
}

type usager interface {
	Usage() string
}

type fileStatus uint8

const (
	fileStatusConflict fileStatus = iota
	fileStatusNoConflict
	fileStatusIdentical
)

func PanicOnError(usager usager, format string, a ...interface{}) {
	panic(Error{usager, fmt.Sprintf(format, a...)})
}

func CopyTemplate(u usager, srcPath, dstPath string, data map[string]interface{}) {
	tmpl, err := template.ParseFiles(srcPath)
	if err != nil {
		PanicOnError(u, "abort: failed to parse template: %v", err)
	}
	var bufFrom bytes.Buffer
	if err := tmpl.Execute(&bufFrom, data); err != nil {
		PanicOnError(u, "abort: failed to process template: %v", err)
	}
	printFunc := PrintCreate
	switch detectConflict(u, bufFrom.Bytes(), dstPath) {
	case fileStatusConflict:
		PrintConflict(dstPath)
		if !confirmOverwrite(dstPath) {
			PrintSkip(dstPath)
			return
		}
		printFunc = PrintOverwrite
	case fileStatusIdentical:
		PrintIdentical(dstPath)
		return
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		PanicOnError(u, "abort: failed to create file: %v", err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, &bufFrom); err != nil {
		PanicOnError(u, "abort: failed to output file: %v", err)
	}
	printFunc(dstPath)
}

func detectConflict(u usager, src []byte, dstPath string) fileStatus {
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		return fileStatusNoConflict
	}
	dstBuf, err := ioutil.ReadFile(dstPath)
	if err != nil {
		PanicOnError(u, "abort: failed to read file: %v", err)
	}
	if bytes.Equal(src, dstBuf) {
		return fileStatusIdentical
	}
	return fileStatusConflict
}

func confirmOverwrite(dstPath string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Overwrite %v? [Yn] ", dstPath)
		yesno, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		switch strings.ToUpper(strings.TrimSpace(yesno)) {
		case "", "YES", "Y":
			return true
		case "NO", "N":
			return false
		}
	}
}

type colorfunc func(string, ...interface{}) string

func Red(s string, a ...interface{}) string {
	return color(31, s, a...)
}

func Green(s string, a ...interface{}) string {
	return color(32, s, a...)
}

func Yellow(s string, a ...interface{}) string {
	return color(33, s, a...)
}

func Blue(s string, a ...interface{}) string {
	return color(34, s, a...)
}

func Magenta(s string, a ...interface{}) string {
	return color(35, s, a...)
}

func Cyan(s string, a ...interface{}) string {
	return color(36, s, a...)
}

func color(colorCode int, s string, a ...interface{}) string {
	switch length := len(a); {
	case length == 0:
		a = append(a, "%s")
	case length > 1:
		panic(errors.New("too many arguments"))
	}
	return fmt.Sprintf(fmt.Sprintf("\x1b[%d;1m%s\x1b[0m", colorCode, a[0]), s)
}

func PrintIdentical(path string) {
	printPathStatus(Blue, "identical", path)
}

func PrintConflict(path string) {
	printPathStatus(Red, "conflict", path)
}

func PrintSkip(path string) {
	printPathStatus(Cyan, "skip", path)
}

func PrintOverwrite(path string) {
	printPathStatus(Cyan, "overwrite", path)
}

func PrintCreate(path string) {
	printPathStatus(Green, "create", path)
}

func PrintExist(path string) {
	printPathStatus(Blue, "exist", path)
}

func PrintCreateDirectory(path string) {
	printPathStatus(Green, "create directory", path)
}

func printPathStatus(f colorfunc, message, s string) {
	fmt.Println(f(message, "%20s"), "", s)
}

// GoString returns Go-syntax representation of the value.
// It returns compilable Go-syntax that different with "%#v" format for fmt package.
func GoString(i interface{}) string {
	switch t := i.(type) {
	case *regexp.Regexp:
		return fmt.Sprintf(`regexp.MustCompile(%q)`, t)
	case *htmltemplate.Template:
		var buf bytes.Buffer
		for _, t := range t.Templates() {
			if t.Name() == "content" {
				continue
			}
			if _, err := buf.WriteString(reflect.ValueOf(t).Elem().FieldByName("text").Elem().FieldByName("text").String()); err != nil {
				panic(err)
			}
		}
		return fmt.Sprintf(`template.Must(template.New(%q).Funcs(kocha.TemplateFuncs).Parse(kocha.Gunzip(%q)))`, t.Name(), Gzip(buf.String()))
	case fmt.GoStringer:
		return t.GoString()
	case nil:
		return "nil"
	}
	v := reflect.ValueOf(i)
	var name string
	if v.Kind() == reflect.Ptr {
		if v = v.Elem(); !v.IsValid() {
			return "nil"
		}
		name = "&"
	}
	name += v.Type().String()
	var (
		tmplStr string
		fields  interface{}
	)
	switch v.Kind() {
	case reflect.Struct:
		f := make(map[string]interface{})
		for i := 0; i < v.NumField(); i++ {
			if tf := v.Type().Field(i); !tf.Anonymous && v.Field(i).CanInterface() {
				f[tf.Name] = GoString(v.Field(i).Interface())
			}
		}
		tmplStr = `
{{.name}}{
	{{range $name, $value := .fields}}
	{{$name}}: {{$value}},
	{{end}}
}`
		fields = f
	case reflect.Slice:
		f := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			f[i] = GoString(v.Index(i).Interface())
		}
		tmplStr = `
{{.name}}{
	{{range $value := .fields}}
	{{$value}},
	{{end}}
}`
		fields = f
	case reflect.Map:
		f := make(map[interface{}]interface{})
		for _, k := range v.MapKeys() {
			f[k.Interface()] = GoString(v.MapIndex(k).Interface())
		}
		tmplStr = `
{{.name}}{
	{{range $name, $value := .fields}}
	{{$name|printf "%q"}}: {{$value}},
	{{end}}
}`
		fields = f
	default:
		return fmt.Sprintf("%#v", v.Interface())
	}
	t := template.Must(template.New(name).Parse(tmplStr))
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]interface{}{
		"name":   name,
		"fields": fields,
	}); err != nil {
		panic(err)
	}
	return buf.String()
}

// Gzip returns gzipped string.
func Gzip(raw string) string {
	var gzipped bytes.Buffer
	w, err := gzip.NewWriterLevel(&gzipped, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	if _, err := w.Write([]byte(raw)); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return gzipped.String()
}

// Gunzip returns unzipped string.
func Gunzip(gz string) string {
	r, err := gzip.NewReader(bytes.NewReader([]byte(gz)))
	if err != nil {
		panic(err)
	}
	result, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func detectContentType(r io.Reader) (contentType string) {
	buf := make([]byte, 512)
	if n, err := io.ReadFull(r, buf); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			panic(err)
		}
		buf = buf[:n]
	}
	if rs, ok := r.(io.Seeker); ok {
		if _, err := rs.Seek(0, os.SEEK_SET); err != nil {
			panic(err)
		}
	}
	return http.DetectContentType(buf)
}
