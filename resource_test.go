package kocha

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func Test_AddResource(t *testing.T) {
	if len(includedResources) != 0 {
		t.Errorf("Expect length is 0, but %v", len(includedResources))
	}
	AddResource("testname", "testdata")
	if len(includedResources) != 1 {
		t.Errorf("Expect length is 1, but %v", len(includedResources))
	}
	rc, ok := includedResources["testname"]
	if !ok {
		t.Errorf("Expect testname exists, but not exists")
	}
	actual := string(rc.data)
	expected := "testdata"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_resource_Open(t *testing.T) {
	rc := &resource{[]byte("testdata1")}
	rs := rc.Open()
	buf1, err := ioutil.ReadAll(rs)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf1)
	expected := "testdata1"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	rs = rc.Open()
	buf2, err := ioutil.ReadAll(rs)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf2)
	expected = "testdata1"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
