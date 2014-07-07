package kocha_test

import (
	"go/build"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/naoina/kocha/util"
)

func Test_buildRouter(t *testing.T) {
	func() {
		util.ImportDir = build.ImportDir
	}()
	util.ImportDir = func(dir string, mode build.ImportMode) (*build.Package, error) {
		pkg, err := build.ImportDir(dir, mode)
		if err != nil {
			return nil, err
		}
		pkg.GoFiles = []string{"testfixtures_test.go"}
		return pkg, err
	}
	for _, v := range []interface{}{
		nil,
		"",
		"hoge",
		struct{ Controller interface{} }{},
		struct{ Controller interface{} }{Controller: ""},
		struct{ Controller interface{} }{Controller: "hoge"},
		struct{ Controller interface{} }{Controller: 1},
	} {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/",
				Controller: v,
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("%#v: RouteTable.buildRouter() => _, nil, want => error", v)
		}
	}

	// test for validate the single argument mismatch between controller method and route parameter.
	func() {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/",
				Controller: FixtureUserTestCtrl{},
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("RouteTable.buildRouter() => _, nil, want => error")
		}
	}()

	// test for validate the argument names mismatch between controller method and route parameter.
	func() {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/:name",
				Controller: FixtureUserTestCtrl{},
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("RouteTable.buildRouter() => _, nil, want => error")
		}
	}()
	func() {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/:id",
				Controller: FixtureRootTestCtrl{},
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("RouteTable.buildRouter() => _, nil, want => error")
		}
	}()

	// test for validate the duplicated route parameters.
	func() {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/:id/:id",
				Controller: FixtureUserTestCtrl{},
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("RouteTable.buildRouter() => _, nil, want => error")
		}
	}()

	// test for validate the multiple arguments mismatch between controller method and route parameters.
	func() {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/",
				Controller: FixtureDateTestCtrl{},
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("RouteTable.buildRouter() => _, nil, want => error")
		}
	}()

	// test for validate the invalid return value.
	for _, v := range []interface{}{
		FixtureInvalidReturnValueTypeTestCtrl{},
		FixtureInvalidNumberOfReturnValueTestCtrl{},
	} {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/",
				Controller: v,
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("%#v: RouteTable.buildRouter() => _, nil, want => error", v)
		}
	}

	// test for validate the TypeValidateParsers.
	func() {
		rt := RouteTable{
			{
				Name:       "testroute",
				Path:       "/:id",
				Controller: FixtureTypeUndefinedCtrl{},
			},
		}
		if _, err := rt.buildRouter(); err == nil {
			t.Errorf("RouteTable.buildRouter() => _, nil, want => error")
		}
	}()
}

func Test_router_dispatch_with_route_missing(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	req, err := http.NewRequest("GET", "/missing", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller, method, args := appConfig.router.dispatch(req)
	if controller != nil {
		t.Errorf("Expect %v, but %v", nil, controller)
	}
	if method != nil {
		t.Errorf("Expect %v, but %v", nil, method)
	}
	if args != nil {
		t.Errorf("Expect %v, but %v", nil, args)
	}
}

func Test_router_dispatch(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller, method, args := appConfig.router.dispatch(req)
	if _, ok := controller.Interface().(*FixtureRootTestCtrl); !ok {
		t.Errorf("Expect %v, but %v", reflect.ValueOf(&FixtureRootTestCtrl{}), controller)
	}
	actual := method.Type().String()
	methodExpected := "func() kocha.Result"
	if !reflect.DeepEqual(actual, methodExpected) {
		t.Errorf("Expect %v, but %v", methodExpected, actual)
	}
	if len(args) != 0 {
		t.Errorf("Expect length is %v, but %v", 0, len(args))
	}

	req, err = http.NewRequest("GET", "/user/777", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller, method, args = appConfig.router.dispatch(req)
	if _, ok := controller.Interface().(*FixtureUserTestCtrl); !ok {
		t.Errorf("Expect %v, but %v", reflect.ValueOf(&FixtureUserTestCtrl{}), controller)
	}
	actual = method.Type().String()
	methodExpected = "func(int) kocha.Result"
	if !reflect.DeepEqual(actual, methodExpected) {
		t.Errorf("Expect %v, but %v", methodExpected, actual)
	}
	argsExpected := []interface{}{777}
	for i, arg := range args {
		if !reflect.DeepEqual(arg.Interface(), argsExpected[i]) {
			t.Errorf("Expect %v, but %v", argsExpected[i], arg)
		}
	}

	// test for invalid path parameter.
	for _, v := range []string{
		"0x16", "1.0", "-1", "10a1", "100a",
	} {
		req, err = http.NewRequest("GET", "/user/"+v, nil)
		if err != nil {
			t.Fatal(err)
		}
		controller, method, args = appConfig.router.dispatch(req)
		if controller != nil {
			t.Errorf("%#v expect nil, but returns instance of %T", v, controller.Interface())
		}
	}

	req, err = http.NewRequest("GET", "/2013/10/19/user/naoina", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller, method, args = appConfig.router.dispatch(req)
	if _, ok := controller.Interface().(*FixtureDateTestCtrl); !ok {
		t.Errorf("Expect %v, but %v", reflect.ValueOf(&FixtureDateTestCtrl{}), controller)
	}
	actual = method.Type().String()
	methodExpected = "func(int, int, int, string) kocha.Result"
	if !reflect.DeepEqual(actual, methodExpected) {
		t.Errorf("Expect %v, but %v", methodExpected, actual)
	}
	argsExpected = []interface{}{2013, 10, 19, "naoina"}
	for i, arg := range args {
		if !reflect.DeepEqual(arg.Interface(), argsExpected[i]) {
			t.Errorf("Expect %v, but %v", argsExpected[i], arg)
		}
	}
}

func TestReverse(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()

	actual := Reverse("root")
	expected := "/"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = Reverse("user", 77)
	expected = "/user/77"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = Reverse("date", 2013, 10, 26, "naoina")
	expected = "/2013/10/26/user/naoina"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	for _, v := range []string{"/hoge.png", "hoge.png"} {
		actual = Reverse("static", v)
		expected = "/static/hoge.png"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}

func TestReverse_with_unknown_route_name(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	Reverse("unknown")
}

func TestReverse_with_few_arguments(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	Reverse("user")
}

func TestReverse_with_many_arguments(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	Reverse("user", 77, 100)
}

func TestReverse_with_type_mismatch(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()

	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	Reverse("user", "naoina")
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
	oldTypeValidateParsers := typeValidateParsers
	typeValidateParsers = make(map[string]TypeValidateParser)
	for k, v := range oldTypeValidateParsers {
		typeValidateParsers[k] = v
	}
	defer func() {
		typeValidateParsers = oldTypeValidateParsers
	}()

	name := "hoge"
	if typeValidateParsers[name] != nil {
		t.Fatal("%v TypeValidateParser has already been set", name)
	}
	var actual, expected TypeValidateParser
	expected = &TestTypeValidateParser{"test1"}
	SetTypeValidateParser(name, expected)
	actual = typeValidateParsers[name]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}

	expected = &TestTypeValidateParser{"test2"}
	SetTypeValidateParser(name, expected)
	actual = typeValidateParsers[name]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}

	expected = &TestTypeValidateParser{"test3"}
	SetTypeValidateParser("string", expected)
	actual = typeValidateParsers["string"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}
}

func Test_SetTypeValidateParserByValue(t *testing.T) {
	oldTypeValidateParsers := typeValidateParsers
	typeValidateParsers = make(map[string]TypeValidateParser)
	for k, v := range oldTypeValidateParsers {
		typeValidateParsers[k] = v
	}
	defer func() {
		typeValidateParsers = oldTypeValidateParsers
	}()

	var actual, expected TypeValidateParser
	expected = &TestTypeValidateParser{"test1"}
	SetTypeValidateParserByValue("hoge", expected)
	if typeValidateParsers["hoge"] != nil {
		t.Errorf("Expect %#v is not set, but set")
	}
	actual = typeValidateParsers["string"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}

	expected = &TestTypeValidateParser{"test2"}
	SetTypeValidateParserByValue(1, expected)
	actual = typeValidateParsers["int"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}

	expected = &TestTypeValidateParser{"test3"}
	SetTypeValidateParserByValue(int32(1), expected)
	actual = typeValidateParsers["int32"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}

	expected = &TestTypeValidateParser{"test4"}
	SetTypeValidateParserByValue([]string{}, expected)
	actual = typeValidateParsers["[]string"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}

	expected = &TestTypeValidateParser{"test5"}
	value := 1
	SetTypeValidateParserByValue(&value, expected)
	actual = typeValidateParsers["*int"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}
}

func Test_StringTypeValidateParser_Validate(t *testing.T) {
	type String string
	validateParser := typeValidateParsers["string"]
	if validateParser == nil {
		t.Fatalf("TypeValidateParser of type %#v is not set", "string")
	}
	for v, expected := range map[interface{}]bool{
		"hoge":          true,
		"a":             true,
		"-":             true,
		"a-b":           true,
		"/":             false,
		"path/to/route": false,
		"":              false,
		1:               false,
		String("a"):     false,
	} {
		actual := validateParser.Validate(v)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
		}
	}
}

func Test_StringTypeValidateParser_Parse(t *testing.T) {
	validateParser := typeValidateParsers["string"]
	if validateParser == nil {
		t.Fatalf("TypeValidateParser of type %#v is not set", "string")
	}
	for _, v := range []string{
		"", "hoge", "foo", "a", "---", "/", "/path/to/route",
	} {
		actual, err := validateParser.Parse(v)
		if err != nil {
			t.Fatal(err)
		}
		expected := v
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
		}
	}
}

func Test_IntTypeValidateParser_Validate(t *testing.T) {
	validateParser := typeValidateParsers["int"]
	if validateParser == nil {
		t.Fatalf("TypeValidateParser of type %#v is not set", "int")
	}
	for v, expected := range map[interface{}]bool{
		0:        true,
		1:        true,
		9:        true,
		10:       true,
		1.1:      false,
		1.0:      false,
		"1":      false,
		int32(1): false,
	} {
		actual := validateParser.Validate(v)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %v, but %v", v, expected, actual)
		}
	}
}

func Test_IntTypeValidateParser_Parse(t *testing.T) {
	validateParser := typeValidateParsers["int"]
	if validateParser == nil {
		t.Fatalf("TypeValidateParser of type %#v is not set", "int")
	}
	for v, expected := range map[string]int{
		"0":   0,
		"1":   1,
		"10":  10,
		"777": 777,
	} {
		actual, err := validateParser.Parse(v)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
		}
	}

	for _, v := range []string{
		"", "a", "0x01", "100a", "a100", "1a0",
	} {
		_, err := validateParser.Parse(v)
		if err == nil {
			t.Errorf("%#v is no error return", v)
		}
	}
}

func Test_URLTypeValidateParser_Validate(t *testing.T) {
	validateParser := typeValidateParsers["*url.URL"]
	if validateParser == nil {
		t.Fatalf("TypeValidateParser of type %#v is not set", "*url.URL")
	}
	for v, expected := range map[interface{}]bool{
		"/":                  true,
		"/path":              true,
		"/path/-/":           true,
		"/path/-/route.html": true,
		"/\x00":              false,
		"/^":                 false,
		"/$$$":               false,
		"":                   false,
	} {
		actual := validateParser.Validate(v)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
		}
	}

	for _, v := range []string{
		"/", "/path", "/path/-/", "/path/-/route.html",
	} {
		u, err := url.Parse(v)
		if err != nil {
			t.Fatal(err)
		}
		actual := validateParser.Validate(u)
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
		}
	}

	type String string
	for _, v := range []interface{}{
		1, 1.0, int32(1), []string(nil), String("/"),
	} {
		actual := validateParser.Validate(v)
		expected := false
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
		}
	}
}

func Test_URLTypeValidateParser_Parse(t *testing.T) {
	validateParser := typeValidateParsers["*url.URL"]
	if validateParser == nil {
		t.Fatalf("TypeValidateParser of type %#v is not set", "*url.URL")
	}
	for _, v := range []string{
		"/", "/path", "/path/to/route", "/path/to/route.html",
	} {
		actual, err := validateParser.Parse(v)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := url.Parse(v)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%#v expect %#v, but %#v", v, expected, actual)
		}
	}
}
