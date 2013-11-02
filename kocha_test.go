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
	expectedConfig := &AppConfig{
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
	}
	Init(expectedConfig)
	if !reflect.DeepEqual(appConfig, expectedConfig) {
		t.Errorf("Expect %v, but %v", expectedConfig, appConfig)
	}
	if !initialized {
		t.Errorf("Expect %v, but %v", true, initialized)
	}
	if maxClientBodySize != DefaultMaxClientBodySize {
		t.Errorf("Expect %v, but %v", DefaultMaxClientBodySize, maxClientBodySize)
	}

	configs["testappname"]["MaxClientBodySize"] = 100
	Init(expectedConfig)
	if maxClientBodySize != 100 {
		t.Errorf("Expect %v, but %v", 100, maxClientBodySize)
	}
}

func TestInit_with_invalid_MaxClientBodySize(t *testing.T) {
	initialized = false
	defer func() {
		initialized = false
	}()
	ap := &AppConfig{
		AppName: "testappname",
	}
	for _, v := range []interface{}{
		"100", 1.1, nil, uint(100),
	} {
		func() {
			configs["testappname"]["MaxClientBodySize"] = v
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("Value is %v, Expect panic, but not occurred", v)
				}
			}()
			Init(ap)
		}()
	}
}
