package kocha

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/naoina/kocha/util"
)

var (
	ErrInvalidFormat        = errors.New("invalid format")
	ErrUnsupportedFieldType = errors.New("unsupported field type")
)

// ParamError indicates that a field has error.
type ParamError struct {
	Name string
	Err  error
}

// NewParamError returns a new ParamError.
func NewParamError(name string, err error) *ParamError {
	return &ParamError{
		Name: name,
		Err:  err,
	}
}

func (e *ParamError) Error() string {
	return fmt.Sprintf("%v is %v", e.Name, e.Err)
}

var formTimeFormats = []string{
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006/01/02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
	"2006/01/02",
	"20060102150405",
	"200601021504",
	"20060102",
}

// Params represents a form values.
type Params struct {
	c *Context
	url.Values
	prefix string
}

func newParams(c *Context, values url.Values, prefix string) *Params {
	return &Params{
		c:      c,
		Values: values,
		prefix: prefix,
	}
}

// From returns a new Params that has prefix made from given name and children.
func (params *Params) From(name string, children ...string) *Params {
	return newParams(params.c, params.Values, params.prefixedName(name, children...))
}

// Bind binds form values of fieldNames to obj.
// obj must be a pointer of struct. If obj isn't a pointer of struct, it returns error.
// Note that it in the case of errors due to a form value binding error, no error is returned.
// Binding errors will set to map of returned from Controller.Errors().
func (params *Params) Bind(obj interface{}, fieldNames ...string) error {
	rvalue := reflect.ValueOf(obj)
	if rvalue.Kind() != reflect.Ptr {
		return fmt.Errorf("kocha: Bind: first argument must be a pointer, but %v", rvalue.Type().Kind())
	}
	for rvalue.Kind() == reflect.Ptr {
		rvalue = rvalue.Elem()
	}
	if rvalue.Kind() != reflect.Struct {
		return fmt.Errorf("kocha: Bind: first argument must be a pointer of struct, but %T", obj)
	}
	rtype := rvalue.Type()
	for _, name := range fieldNames {
		index := params.findFieldIndex(rtype, name, nil)
		if len(index) < 1 {
			_, filename, line, _ := runtime.Caller(1)
			params.c.App.Logger.Warnf(
				"kocha: Bind: %s:%s: field name `%s' given, but %s.%s is undefined",
				filepath.Base(filename), line, name, rtype.Name(), util.ToCamelCase(name))
			continue
		}
		fname := params.prefixedName(params.prefix, name)
		values, found := params.Values[fname]
		if !found {
			continue
		}
		field := rvalue.FieldByIndex(index)
		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}
		value, err := params.parse(field.Interface(), values[0])
		if err != nil {
			params.c.Errors()[name] = append(params.c.Errors()[name], NewParamError(name, err))
		}
		field.Set(reflect.ValueOf(value))
	}
	return nil
}

func (params *Params) prefixedName(prefix string, names ...string) string {
	if prefix != "" {
		names = append([]string{prefix}, names...)
	}
	return strings.Join(names, ".")
}

type embeddefFieldInfo struct {
	field reflect.StructField
	name  string
	index []int
}

func (params *Params) findFieldIndex(rtype reflect.Type, name string, index []int) []int {
	var embeddedFieldInfos []*embeddefFieldInfo
	for i := 0; i < rtype.NumField(); i++ {
		field := rtype.Field(i)
		if util.IsUnexportedField(field) {
			continue
		}
		if field.Anonymous {
			embeddedFieldInfos = append(embeddedFieldInfos, &embeddefFieldInfo{field, name, append(index, i)})
			continue
		}
		if field.Name == util.ToCamelCase(name) {
			return append(index, i)
		}
	}
	for _, fi := range embeddedFieldInfos {
		if index := params.findFieldIndex(fi.field.Type, fi.name, fi.index); len(index) > 0 {
			return index
		}
	}
	return nil
}

func (params *Params) parse(fv interface{}, vStr string) (value interface{}, err error) {
	switch t := fv.(type) {
	case sql.Scanner:
		err = t.Scan(vStr)
	case time.Time:
		for _, format := range formTimeFormats {
			if value, err = time.Parse(format, vStr); err == nil {
				break
			}
		}
	case string:
		value = vStr
	case bool:
		value, err = strconv.ParseBool(vStr)
	case int, int8, int16, int32, int64:
		if value, err = strconv.ParseInt(vStr, 10, 0); err == nil {
			value = reflect.ValueOf(value).Convert(reflect.TypeOf(t)).Interface()
		}
	case uint, uint8, uint16, uint32, uint64:
		if value, err = strconv.ParseUint(vStr, 10, 0); err == nil {
			value = reflect.ValueOf(value).Convert(reflect.TypeOf(t)).Interface()
		}
	case float32, float64:
		if value, err = strconv.ParseFloat(vStr, 0); err == nil {
			value = reflect.ValueOf(value).Convert(reflect.TypeOf(t)).Interface()
		}
	default:
		params.c.App.Logger.Warnf("kocha: Bind: unsupported field type: %T", t)
		err = ErrUnsupportedFieldType
	}
	if err != nil {
		if err != ErrUnsupportedFieldType {
			params.c.App.Logger.Warnf("kocha: Bind: %v", err)
			err = ErrInvalidFormat
		}
		return nil, err
	}
	return value, nil
}
