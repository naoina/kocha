package util

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"go/build"
	"go/format"

	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/naoina/go-stringutil"
)

const (
	TemplateSuffix = ".tmpl"
)

var (
	// Now returns current time. This is for mock in tests.
	Now = func() time.Time { return time.Now() }

	// for test.
	ImportDir = build.ImportDir

	printColor = func(_, format string, a ...interface{}) { fmt.Printf(format, a...) }
)

func ToCamelCase(s string) string {
	return stringutil.ToUpperCamelCase(s)
}

func ToSnakeCase(s string) string {
	return stringutil.ToSnakeCase(s)
}

func NormPath(p string) string {
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
	fileStatusConflict fileStatus = iota + 1
	fileStatusNoConflict
	fileStatusIdentical
)

func CopyTemplate(srcPath, dstPath string, data map[string]interface{}) error {
	tmpl, err := template.ParseFiles(srcPath)
	if err != nil {
		return fmt.Errorf("kocha: failed to parse template: %v: %v", srcPath, err)
	}
	var bufFrom bytes.Buffer
	if err := tmpl.Execute(&bufFrom, data); err != nil {
		return fmt.Errorf("kocha: failed to process template: %v: %v", srcPath, err)
	}
	buf := bufFrom.Bytes()
	if strings.HasSuffix(srcPath, ".go"+TemplateSuffix) {
		if buf, err = format.Source(buf); err != nil {
			return fmt.Errorf("kocha: failed to gofmt: %v: %v", srcPath, err)
		}
	}
	dstDir := filepath.Dir(dstPath)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		PrintCreateDirectory(dstDir)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("kocha: failed to create directory: %v: %v", dstDir, err)
		}
	}
	printFunc := PrintCreate
	status, err := detectConflict(buf, dstPath)
	if err != nil {
		return err
	}
	switch status {
	case fileStatusConflict:
		PrintConflict(dstPath)
		if !confirmOverwrite(dstPath) {
			PrintSkip(dstPath)
			return nil
		}
		printFunc = PrintOverwrite
	case fileStatusIdentical:
		PrintIdentical(dstPath)
		return nil
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("kocha: failed to create file: %v: %v", dstPath, err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, bytes.NewBuffer(buf)); err != nil {
		return fmt.Errorf("kocha: failed to output file: %v: %v", dstPath, err)
	}
	printFunc(dstPath)
	return nil
}

func detectConflict(src []byte, dstPath string) (fileStatus, error) {
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		return fileStatusNoConflict, nil
	}
	dstBuf, err := ioutil.ReadFile(dstPath)
	if err != nil {
		return 0, fmt.Errorf("kocha: failed to read file: %v", err)
	}
	if bytes.Equal(src, dstBuf) {
		return fileStatusIdentical, nil
	}
	return fileStatusConflict, nil
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

func makePrintColor(w io.Writer, color, format string, a ...interface{}) {
	fmt.Fprintf(w, "\x1b[%s;1m", color)
	fmt.Fprintf(w, format, a...)
	fmt.Fprint(w, "\x1b[0m")
}

func PrintGreen(s string, a ...interface{}) {
	printColor("32", s, a...)
}

func PrintIdentical(path string) {
	printPathStatus("34", "identical", path) // Blue.
}

func PrintConflict(path string) {
	printPathStatus("31", "conflict", path) // Red.
}

func PrintSkip(path string) {
	printPathStatus("36", "skip", path) // Cyan.
}

func PrintOverwrite(path string) {
	printPathStatus("36", "overwrite", path) // Cyan.
}

func PrintCreate(path string) {
	printPathStatus("32", "create", path) // Green.
}

func PrintCreateDirectory(path string) {
	printPathStatus("32", "create directory", path) // Green.
}

func printPathStatus(color, message, s string) {
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
		return fmt.Sprintf(`template.Must(template.New(%q).Funcs(kocha.TemplateFuncs).Parse(util.Gunzip(%q)))`, t.Name(), Gzip(buf.String()))
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

var settingEnvRegexp = regexp.MustCompile(`\bkocha\.Getenv\(\s*(.+?)\s*,\s*(.+?)\s*\)`)

// FindEnv returns map of environment variables.
// Key of map is key of environment variable, Value of map is value of
// environment variable.
func FindEnv(basedir string) (map[string]string, error) {
	if basedir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		basedir = pwd
	}
	env := make(map[string]string)
	if err := filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch info.Name()[0] {
		case '.', '_':
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		matches := settingEnvRegexp.FindAllStringSubmatch(string(body), -1)
		if matches == nil {
			return nil
		}
		for _, m := range matches {
			key, err := strconv.Unquote(m[1])
			if err != nil {
				continue
			}
			value, err := strconv.Unquote(m[2])
			if err != nil {
				value = "WILL BE SET IN RUNTIME"
			}
			env[key] = value
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return env, nil
}

// FindAppDir returns application directory. (aka import path)
// An application directory retrieves from current working directory.
// For example, if current working directory is
// "/path/to/gopath/src/github.com/naoina/myapp", FindAppDir returns
// "github.com/naoina/myapp".
func FindAppDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	bp, err := filepath.EvalSymlinks(filepath.Join(filepath.SplitList(build.Default.GOPATH)[0], "src"))
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(dir)[len(bp)+1:], nil
}

// FindAbsDir returns an absolute path of importPath in GOPATH.
// For example, if importPath is "github.com/naoina/myapp",
// and GOPATH is "/path/to/gopath", FindAbsDir returns
// "/path/to/gopath/src/github.com/naoina/myapp".
func FindAbsDir(importPath string) (string, error) {
	if importPath == "" {
		return os.Getwd()
	}
	dir := filepath.FromSlash(importPath)
	for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
		candidate := filepath.Join(gopath, "src", dir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("package `%s' not found in GOPATH", importPath)
}

// IsUnexportedField returns whether the field is unexported.
// This function is to avoid the bug in versions older than Go1.3.
// See following links:
//     https://code.google.com/p/go/issues/detail?id=7247
//     http://golang.org/ref/spec#Exported_identifiers
func IsUnexportedField(field reflect.StructField) bool {
	return !(field.PkgPath == "" && unicode.IsUpper(rune(field.Name[0])))
}

// Generate a random bytes.
func GenerateRandomKey(length int) []byte {
	result := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, result); err != nil {
		panic(err)
	}
	return result
}

func PrintEnv(basedir string) error {
	envMap, err := FindEnv(basedir)
	if err != nil {
		return err
	}
	envKeys := make([]string, 0, len(envMap))
	for k := range envMap {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "kocha: you can be setting for your app by the following environment variables at the time of launching the app:\n\n")
	for _, k := range envKeys {
		v := os.Getenv(k)
		if v == "" {
			v = envMap[k]
		}
		fmt.Fprintf(&buf, "%4s%v=%v\n", "", k, strconv.Quote(v))
	}
	fmt.Println(buf.String())
	return nil
}

type Commander interface {
	Run(args []string) error
	Name() string
	Usage() string
	Option() interface{}
}

func RunCommand(cmd Commander) {
	parser := flags.NewNamedParser(cmd.Name(), flags.PrintErrors|flags.PassDoubleDash|flags.PassAfterNonOption)
	if _, err := parser.AddGroup("", "", cmd.Option()); err != nil {
		panic(err)
	}
	args, err := parser.Parse()
	if err != nil {
		fmt.Fprint(os.Stderr, cmd.Usage())
		os.Exit(1)
	}
	opt := reflect.ValueOf(cmd.Option())
	for opt.Kind() == reflect.Ptr {
		opt = opt.Elem()
	}
	h := opt.FieldByName("Help")
	if h.IsValid() && h.Kind() == reflect.Bool && h.Bool() {
		fmt.Fprint(os.Stderr, cmd.Usage())
		os.Exit(0)
	}
	if err := cmd.Run(args); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintf(os.Stderr, "%s: %v\n", cmd.Name(), err)
			fmt.Fprint(os.Stderr, cmd.Usage())
		}
		os.Exit(1)
	}
}

func init() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		w := colorable.NewColorableStdout()
		printColor = func(color, format string, a ...interface{}) {
			fmt.Fprintf(w, "\x1b[%s;1m", color)
			fmt.Fprintf(w, format, a...)
			fmt.Fprint(w, "\x1b[0m")
		}
	}
}
