package kocha_test

import (
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestRouter_Reverse(t *testing.T) {
	app := kocha.NewTestApp()
	actual := app.Router.Reverse("root")
	expected := "/"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = app.Router.Reverse("user", 77)
	expected = "/user/77"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = app.Router.Reverse("date", 2013, 10, 26, "naoina")
	expected = "/2013/10/26/user/naoina"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	for _, v := range []string{"/hoge.png", "hoge.png"} {
		actual = app.Router.Reverse("static", v)
		expected = "/static/hoge.png"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}

func TestRouter_Reverse_withUnknownRouteName(t *testing.T) {
	app := kocha.NewTestApp()
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	app.Router.Reverse("unknown")
}

func TestRouter_Reverse_withFewArguments(t *testing.T) {
	app := kocha.NewTestApp()
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	app.Router.Reverse("user")
}

func TestRouter_Reverse_withManyArguments(t *testing.T) {
	app := kocha.NewTestApp()
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	app.Router.Reverse("user", 77, 100)
}

func TestRouter_Reverse_withTypeMismatch(t *testing.T) {
	app := kocha.NewTestApp()
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	app.Router.Reverse("user", "naoina")
}

type TestTypeValidateParser struct {
	id string
}

func (validateParser *TestTypeValidateParser) Validate(v interface{}) bool {
	return true
}

func (validateParser *TestTypeValidateParser) Parse(s string) (value interface{}, err error) {
	return nil, nil
}

func Test_SetTypeValidateParser(t *testing.T) {
	t.Skipf("TODO")
	// oldTypeValidateParsers := typeValidateParsers
	// typeValidateParsers = make(map[string]TypeValidateParser)
	// for k, v := range oldTypeValidateParsers {
	// typeValidateParsers[k] = v
	// }
	// defer func() {
	// typeValidateParsers = oldTypeValidateParsers
	// }()

	// name := "hoge"
	// if typeValidateParsers[name] != nil {
	// t.Fatal("%v TypeValidateParser has already been set", name)
	// }
	// var actual, expected TypeValidateParser
	// expected = &TestTypeValidateParser{"test1"}
	// SetTypeValidateParser(name, expected)
	// actual = typeValidateParsers[name]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }

	// expected = &TestTypeValidateParser{"test2"}
	// SetTypeValidateParser(name, expected)
	// actual = typeValidateParsers[name]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }

	// expected = &TestTypeValidateParser{"test3"}
	// SetTypeValidateParser("string", expected)
	// actual = typeValidateParsers["string"]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }
}

func Test_SetTypeValidateParserByValue(t *testing.T) {
	t.Skipf("TODO")
	// oldTypeValidateParsers := typeValidateParsers
	// typeValidateParsers = make(map[string]TypeValidateParser)
	// for k, v := range oldTypeValidateParsers {
	// typeValidateParsers[k] = v
	// }
	// defer func() {
	// typeValidateParsers = oldTypeValidateParsers
	// }()

	// var actual, expected TypeValidateParser
	// expected = &TestTypeValidateParser{"test1"}
	// SetTypeValidateParserByValue("hoge", expected)
	// if typeValidateParsers["hoge"] != nil {
	// t.Errorf("Expect %#v is not set, but set")
	// }
	// actual = typeValidateParsers["string"]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }

	// expected = &TestTypeValidateParser{"test2"}
	// SetTypeValidateParserByValue(1, expected)
	// actual = typeValidateParsers["int"]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }

	// expected = &TestTypeValidateParser{"test3"}
	// SetTypeValidateParserByValue(int32(1), expected)
	// actual = typeValidateParsers["int32"]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }

	// expected = &TestTypeValidateParser{"test4"}
	// SetTypeValidateParserByValue([]string{}, expected)
	// actual = typeValidateParsers["[]string"]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }

	// expected = &TestTypeValidateParser{"test5"}
	// value := 1
	// SetTypeValidateParserByValue(&value, expected)
	// actual = typeValidateParsers["*int"]
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("Expect %#v, but %#v", expected, actual)
	// }
}

func Test_StringTypeValidateParser_Validate(t *testing.T) {
	t.Skipf("TODO")
	// type String string
	// validateParser := typeValidateParsers["string"]
	// if validateParser == nil {
	// t.Fatalf("TypeValidateParser of type %#v is not set", "string")
	// }
	// for v, expected := range map[interface{}]bool{
	// "hoge":          true,
	// "a":             true,
	// "-":             true,
	// "a-b":           true,
	// "/":             false,
	// "path/to/route": false,
	// "":              false,
	// 1:               false,
	// String("a"):     false,
	// } {
	// actual := validateParser.Validate(v)
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
	// }
	// }
}

func Test_StringTypeValidateParser_Parse(t *testing.T) {
	t.Skipf("TODO")
	// validateParser := typeValidateParsers["string"]
	// if validateParser == nil {
	// t.Fatalf("TypeValidateParser of type %#v is not set", "string")
	// }
	// for _, v := range []string{
	// "", "hoge", "foo", "a", "---", "/", "/path/to/route",
	// } {
	// actual, err := validateParser.Parse(v)
	// if err != nil {
	// t.Fatal(err)
	// }
	// expected := v
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
	// }
	// }
}

func Test_IntTypeValidateParser_Validate(t *testing.T) {
	t.Skipf("TODO")
	// validateParser := typeValidateParsers["int"]
	// if validateParser == nil {
	// t.Fatalf("TypeValidateParser of type %#v is not set", "int")
	// }
	// for v, expected := range map[interface{}]bool{
	// 0:        true,
	// 1:        true,
	// 9:        true,
	// 10:       true,
	// 1.1:      false,
	// 1.0:      false,
	// "1":      false,
	// int32(1): false,
	// } {
	// actual := validateParser.Validate(v)
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %v, but %v", v, expected, actual)
	// }
	// }
}

func Test_IntTypeValidateParser_Parse(t *testing.T) {
	t.Skipf("TODO")
	// validateParser := typeValidateParsers["int"]
	// if validateParser == nil {
	// t.Fatalf("TypeValidateParser of type %#v is not set", "int")
	// }
	// for v, expected := range map[string]int{
	// "0":   0,
	// "1":   1,
	// "10":  10,
	// "777": 777,
	// } {
	// actual, err := validateParser.Parse(v)
	// if err != nil {
	// t.Fatal(err)
	// }
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
	// }
	// }

	// for _, v := range []string{
	// "", "a", "0x01", "100a", "a100", "1a0",
	// } {
	// _, err := validateParser.Parse(v)
	// if err == nil {
	// t.Errorf("%#v is no error return", v)
	// }
	// }
}

func Test_URLTypeValidateParser_Validate(t *testing.T) {
	t.Skipf("TODO")
	// validateParser := typeValidateParsers["*url.URL"]
	// if validateParser == nil {
	// t.Fatalf("TypeValidateParser of type %#v is not set", "*url.URL")
	// }
	// for v, expected := range map[interface{}]bool{
	// "/":                  true,
	// "/path":              true,
	// "/path/-/":           true,
	// "/path/-/route.html": true,
	// "/\x00":              false,
	// "/^":                 false,
	// "/$$$":               false,
	// "":                   false,
	// } {
	// actual := validateParser.Validate(v)
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
	// }
	// }

	// for _, v := range []string{
	// "/", "/path", "/path/-/", "/path/-/route.html",
	// } {
	// u, err := url.Parse(v)
	// if err != nil {
	// t.Fatal(err)
	// }
	// actual := validateParser.Validate(u)
	// expected := true
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
	// }
	// }

	// type String string
	// for _, v := range []interface{}{
	// 1, 1.0, int32(1), []string(nil), String("/"),
	// } {
	// actual := validateParser.Validate(v)
	// expected := false
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
	// }
	// }
}

func Test_URLTypeValidateParser_Parse(t *testing.T) {
	t.Skipf("TODO")
	// validateParser := typeValidateParsers["*url.URL"]
	// if validateParser == nil {
	// t.Fatalf("TypeValidateParser of type %#v is not set", "*url.URL")
	// }
	// for _, v := range []string{
	// "/", "/path", "/path/to/route", "/path/to/route.html",
	// } {
	// actual, err := validateParser.Parse(v)
	// if err != nil {
	// t.Fatal(err)
	// }
	// expected, err := url.Parse(v)
	// if err != nil {
	// t.Fatal(err)
	// }
	// if !reflect.DeepEqual(actual, expected) {
	// t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
	// }
	// }
}
