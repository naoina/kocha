package kocha

import (
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	if len(configs) != 0 {
		t.Fatalf("configs expect empty, but not empty")
	}
	actual := Config("testAppName")
	if len(configs) != 1 {
		t.Fatalf("configs expect one, but not one")
	}
	defer func() {
		delete(configs, "testAppName")
	}()
	expected := configs["testAppName"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = Config("testAppName")
	if len(configs) != 1 {
		t.Fatalf("configs expect one, but not one")
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_configGet(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	actual, ok := c.Get("key1")
	if ok {
		t.Errorf("Expect false, but %v", ok)
	}
	if !reflect.DeepEqual(actual, nil) {
		t.Errorf("Expect %v, but %v", nil, actual)
	}

	c["key1"] = "testValue1"
	actual, ok = c.Get("key1")
	if !ok {
		t.Errorf("Expect true, but %v", ok)
	}
	expected := "testValue1"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", nil, actual)
	}
}

func Test_configSet(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	if _, ok := c["testKey"]; ok {
		t.Fatalf("Expect false, but %v", ok)
	}
	c.Set("testKey", "testValue1")
	actual, ok := c["testKey"]
	if !ok {
		t.Errorf("Expect true, but %v", ok)
	}
	expected := "testValue1"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c.Set("testKey", "testValue2")
	actual, ok = c["testKey"]
	if !ok {
		t.Errorf("Expect true, but %v", ok)
	}
	expected = "testValue2"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_configInt(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	c["testKey"] = 777
	actual, ok := c.Int("testKey")
	if !ok {
		t.Errorf("Expect true, but %v", ok)
	}
	expected := 777
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_configIntDefault(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	actual := c.IntDefault("testKey", 777)
	expected := 777
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c["testKey"] = 888
	actual = c.IntDefault("testKey", 777)
	expected = 888
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_configString(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	c["testKey"] = "testValue"
	actual, ok := c.String("testKey")
	if !ok {
		t.Errorf("Expect true, but %v", ok)
	}
	expected := "testValue"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_configStringDefault(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	actual := c.StringDefault("testKey", "testValue1")
	expected := "testValue1"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c["testKey"] = "testValue2"
	actual = c.StringDefault("testKey", "testValue1")
	expected = "testValue2"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_configBool(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	c["testKey"] = true
	actual, ok := c.Bool("testKey")
	if !ok {
		t.Errorf("Expect true, but %v", ok)
	}
	expected := true
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_configBoolDefault(t *testing.T) {
	c := Config("testAppName")
	defer func() {
		delete(configs, "testAppName")
	}()
	actual := c.BoolDefault("testKey", true)
	expected := true
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c["testKey"] = false
	actual = c.BoolDefault("testKey", true)
	expected = false
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
