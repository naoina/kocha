package kocha

import (
	"reflect"
	"testing"
)

func TestConst(t *testing.T) {
	if !reflect.DeepEqual(DefaultHttpAddr, "0.0.0.0") {
		t.Errorf(`Expect %v, but %v`, "0.0.0.0", DefaultHttpAddr)
	}

	if !reflect.DeepEqual(DefaultHttpPort, 80) {
		t.Errorf("Expect %v, but %v", 80, DefaultHttpPort)
	}

	if !reflect.DeepEqual(DefaultMaxClientBodySize, 1024*1024*10) {
		t.Errorf("Expect %v, but %v", 1024*1024*10, DefaultMaxClientBodySize)
	}
}

func TestInit(t *testing.T) {
	initialized = false
	defer func() {
		initialized = false
	}()
	config := &AppConfig{
		AppPath:     "testpath",
		AppName:     "testappname",
		TemplateSet: nil,
		RouteTable: RouteTable{
			{
				Name:        "route1",
				Path:        "route_path1",
				Controller:  nil,
				MethodTypes: nil,
				RegexpPath:  nil,
			},
			{
				Name:        "route2",
				Path:        "route_path2",
				Controller:  nil,
				MethodTypes: nil,
				RegexpPath:  nil,
			},
		},
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
