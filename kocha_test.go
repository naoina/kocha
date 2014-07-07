package kocha_test

import (
	"os"
	"reflect"
	"testing"
)

func TestConst(t *testing.T) {
	for actual, expected := range map[interface{}]interface{}{
		DefaultHttpAddr:          "127.0.0.1:9100",
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
		RouteTable: RouteTable{
			{
				Name:       "route1",
				Path:       "route_path1",
				Controller: &FixtureRootTestCtrl{},
			},
			{
				Name:       "route2",
				Path:       "route_path2",
				Controller: &FixtureRootTestCtrl{},
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

func TestSettingEnv(t *testing.T) {
	func() {
		actual := SettingEnv("TEST_KOCHA_ENV", "default value")
		expected := "default value"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("SettingEnv(%q, %q) => %q, want %q", "TEST_KOCHA_ENV", "default value", actual, expected)
		}

		actual = os.Getenv("TEST_KOCHA_ENV")
		expected = "default value"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("os.Getenv(%q) => %q, want %q", "TEST_KOCHA_ENV", actual, expected)
		}
	}()

	func() {
		os.Setenv("TEST_KOCHA_ENV", "set kocha env")
		defer os.Clearenv()
		actual := SettingEnv("TEST_KOCHA_ENV", "default value")
		expected := "set kocha env"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("SettingEnv(%q, %q) => %q, want %q", "TEST_KOCHA_ENV", "default value", actual, expected)
		}

		actual = os.Getenv("TEST_KOCHA_ENV")
		expected = "set kocha env"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("os.Getenv(%q) => %q, want %q", "TEST_KOCHA_ENV", actual, expected)
		}
	}()
}
