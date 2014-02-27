package kocha

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"
	"unicode"

	"go/format"

	"github.com/daviddengcn/go-colortext"
)

var (
	// Now returns current time. This is for mock in tests.
	Now = func() time.Time { return time.Now() }
)

// ToCamelCase returns a copy of the string s with all Unicode letters mapped to their camel case.
// It will convert to upper case previous letter of '_' and first letter, and remove letter of '_'.
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}
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
	if s == "" {
		return ""
	}
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

// SplitExt returns pair of file name and extension.
//
// It will truncated a dot of extension.
// e.g. When the given path is "path/to/image.png", SplitExt returns ("path/to/image", "png").
func SplitExt(path string) (name, ext string) {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[:i], path[i+1:]
		}
	}
	return path, ""
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
	buf := bufFrom.Bytes()
	if strings.HasSuffix(srcPath, ".go.template") {
		if buf, err = format.Source(buf); err != nil {
			PanicOnError(u, "abort: failed to gofmt: %v", err)
		}
	}
	dstDir := filepath.Dir(dstPath)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		PrintCreateDirectory(dstDir)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			PanicOnError(u, "abort: failed to create directory: %v", err)
		}
	}
	printFunc := PrintCreate
	switch detectConflict(u, buf, dstPath) {
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
	if _, err := io.Copy(dstFile, bytes.NewBuffer(buf)); err != nil {
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

func PrintRed(s string, a ...interface{}) {
	printColor(ct.Red, s, a...)
}

func PrintGreen(s string, a ...interface{}) {
	printColor(ct.Green, s, a...)
}

func PrintYellow(s string, a ...interface{}) {
	printColor(ct.Yellow, s, a...)
}

func PrintBlue(s string, a ...interface{}) {
	printColor(ct.Blue, s, a...)
}

func PrintMagenta(s string, a ...interface{}) {
	printColor(ct.Magenta, s, a...)
}

func PrintCyan(s string, a ...interface{}) {
	printColor(ct.Cyan, s, a...)
}

func printColor(color ct.Color, format string, a ...interface{}) {
	ct.ChangeColor(color, true, ct.None, false)
	fmt.Printf(format, a...)
	ct.ResetColor()
}

func PrintIdentical(path string) {
	printPathStatus(ct.Blue, "identical", path)
}

func PrintConflict(path string) {
	printPathStatus(ct.Red, "conflict", path)
}

func PrintSkip(path string) {
	printPathStatus(ct.Cyan, "skip", path)
}

func PrintOverwrite(path string) {
	printPathStatus(ct.Cyan, "overwrite", path)
}

func PrintCreate(path string) {
	printPathStatus(ct.Green, "create", path)
}

func PrintExist(path string) {
	printPathStatus(ct.Blue, "exist", path)
}

func PrintCreateDirectory(path string) {
	printPathStatus(ct.Green, "create directory", path)
}

func printPathStatus(color ct.Color, message, s string) {
	printColor(color, "%20s", message)
	fmt.Println("", s)
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
