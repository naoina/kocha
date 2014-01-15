package kocha

import (
	"reflect"
	"testing"
)

func TestConst(t *testing.T) {
	for actual, expected := range map[interface{}]interface{}{
		DefaultHttpAddr:          "0.0.0.0",
		DefaultHttpPort:          80,
		DefaultMaxClientBodySize: 1024 * 1024 * 10,
		StaticDir:                "public",
	} {
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}

func TestInit(t *testing.T) {
	initialized = false
	defer func() {
		initialized = false
	}()
	config := &AppConfig{
		AppPath:       "testpath",
		AppName:       "testappname",
		DefaultLayout: "testapp",
		TemplateSet:   nil,
		Router: NewRouter(RouteTable{
			{
				Name:        "route1",
				Path:        "route_path1",
				Controller:  nil,
				MethodTypes: nil,
			},
			{
				Name:        "route2",
				Path:        "route_path2",
				Controller:  nil,
				MethodTypes: nil,
			},
		}),
		Logger: &Logger{},
	}
	Init(config)
	actual := appConfig
	expected := config
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if config.MaxClientBodySize != DefaultMaxClientBodySize {
		t.Errorf("Expect %v, but %v", DefaultMaxClientBodySize, config.MaxClientBodySize)
	}

	if !initialized {
		t.Errorf("Expect %v, but %v", true, initialized)
	}

	config.MaxClientBodySize = -1
	Init(config)
	actual = appConfig
	expected = config
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if config.MaxClientBodySize != DefaultMaxClientBodySize {
		t.Errorf("Expect %v, but %v", DefaultMaxClientBodySize, config.MaxClientBodySize)
	}

	config.MaxClientBodySize = 20131108
	Init(config)
	actual = appConfig
	expected = config
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	if config.MaxClientBodySize != 20131108 {
		t.Errorf("Expect %v, but %v", 20131108, config.MaxClientBodySize)
	}
}
