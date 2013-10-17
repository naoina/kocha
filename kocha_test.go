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
		RouteTable: []*Route{
			&Route{
				Name:        "route1",
				Path:        "route_path1",
				Controller:  nil,
				MethodTypes: nil,
				RegexpPath:  nil,
			},
			&Route{
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
}
