package kocha

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestInitRouteTable(t *testing.T) {
	actual := InitRouteTable(RouteTable{
		{
			Name:       "root",
			Path:       "/",
			Controller: FixtureRootTestCtrl{},
		},
		{
			Name:       "root_indirect",
			Path:       "/indirect",
			Controller: &FixtureRootTestCtrl{},
		},
		{
			Name:       "user",
			Path:       "/user/:id",
			Controller: FixtureUserTestCtrl{},
		},
		{
			Name:       "date",
			Path:       "/:year/:month/:day/user/:name",
			Controller: FixtureDateTestCtrl{},
		},
		{
			Name:       "static",
			Path:       "/static/*path",
			Controller: StaticServe{},
		},
	})
	expected := RouteTable{
		{
			Name:       "root",
			Path:       "/",
			Controller: FixtureRootTestCtrl{},
			MethodTypes: map[string]MethodArgs{
				"Get": MethodArgs{},
			},
		},
		{
			Name:       "root_indirect",
			Path:       "/indirect",
			Controller: FixtureRootTestCtrl{},
			MethodTypes: map[string]MethodArgs{
				"Get": MethodArgs{},
			},
		},
		{
			Name:       "user",
			Path:       "/user/:id",
			Controller: FixtureUserTestCtrl{},
			MethodTypes: map[string]MethodArgs{
				"Get": MethodArgs{
					"id": "int",
				},
			},
		},
		{
			Name:       "date",
			Path:       "/:year/:month/:day/user/:name",
			Controller: FixtureDateTestCtrl{},
			MethodTypes: map[string]MethodArgs{
				"Get": MethodArgs{
					"year":  "int",
					"month": "int",
					"day":   "int",
					"name":  "string",
				},
			},
		},
		{
			Name:       "static",
			Path:       "/static/*path",
			Controller: StaticServe{},
			MethodTypes: map[string]MethodArgs{
				"Get": MethodArgs{
					"path": "url.URL",
				},
			},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %q, but %q", expected, actual)
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
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("panic doesn't happened by %v", v)
				}
			}()
			InitRouteTable(RouteTable{
				{
					Name:       "testroute",
					Path:       "/",
					Controller: v,
				},
			})
		}()
	}

	// test for validate the single argument mismatch between controller method and route parameter.
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		InitRouteTable(RouteTable{
			{
				Name:       "testroute",
				Path:       "/",
				Controller: FixtureUserTestCtrl{},
			},
		})
	}()

	// test for validate the argument names mismatch between controller method and route parameter.
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		InitRouteTable(RouteTable{
			{
				Name:       "testroute",
				Path:       "/:name",
				Controller: FixtureUserTestCtrl{},
			},
		})
	}()

	// test for validate the duplicated route parameters.
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		InitRouteTable(RouteTable{
			{
				Name:       "testroute",
				Path:       "/:id/:id",
				Controller: FixtureUserTestCtrl{},
			},
		})
	}()

	// test for validate the multiple arguments mismatch between controller method and route parameters.
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't happened")
			}
		}()
		InitRouteTable(RouteTable{
			{
				Name:       "testroute",
				Path:       "/",
				Controller: FixtureDateTestCtrl{},
			},
		})
	}()

	// test for validate the invalid return value.
	for _, v := range []interface{}{
		FixtureInvalidReturnValueTypeTestCtrl{},
		FixtureInvalidNumberOfReturnValueTestCtrl{},
	} {
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("panic doesn't happened")
				}
			}()
			InitRouteTable(RouteTable{
				{
					Name:       "testroute",
					Path:       "/",
					Controller: v,
				},
			})
		}()
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

func TestReverse_with_type_macher_is_not_defined(t *testing.T) {
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
	Reverse("type_undefined", 1)
}

func Test_dispatch_with_route_missing(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	req, err := http.NewRequest("GET", "/missing", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller, method, args := dispatch(req)
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

func Test_dispatch(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller, method, args := dispatch(req)
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
	controller, method, args = dispatch(req)
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
		controller, method, args = dispatch(req)
		if controller != nil {
			t.Errorf("%#v expect nil, but returns instance of %T", v, controller.Interface())
		}
	}

	req, err = http.NewRequest("GET", "/2013/10/19/user/naoina", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller, method, args = dispatch(req)
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

func Test_StringTypeValidateParser_Validate(t *testing.T) {
	type String string
	validateParser := typeValidateParsers["string"]
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
	validateParser := typeValidateParsers["url.URL"]
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
	validateParser := typeValidateParsers["url.URL"]
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
