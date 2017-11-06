package kocha_test

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/naoina/kocha"
)

func TestSession(t *testing.T) {
	sess := make(kocha.Session)
	key := "test_key"
	var actual interface{} = sess.Get(key)
	var expected interface{} = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Get(%#v) => %#v; want %#v`, key, actual, expected)
	}
	actual = len(sess)
	expected = 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`len(Session) => %#v; want %#v`, actual, expected)
	}

	value := "test_value"
	sess.Set(key, value)
	actual = sess.Get(key)
	expected = value
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); Session.Get(%#v) => %#v; want %#v`, key, value, key, actual, expected)
	}
	actual = len(sess)
	expected = 1
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); len(Session) => %#v; want %#v`, key, value, actual, expected)
	}

	key2 := "test_key2"
	value2 := "test_value2"
	sess.Set(key2, value2)
	actual = sess.Get(key2)
	expected = value2
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); Session.Get(%#v) => %#v; want %#v`, key2, value2, key2, actual, expected)
	}
	actual = len(sess)
	expected = 2
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); len(Session) => %#v; want %#v`, key2, value2, actual, expected)
	}

	value3 := "test_value3"
	sess.Set(key, value3)
	actual = sess.Get(key)
	expected = value3
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); Session.Get(%#v) => %#v; want %#v`, key, value3, key, actual, expected)
	}
	actual = len(sess)
	expected = 2
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); len(Session) => %#v; want %#v`, key, value3, actual, expected)
	}

	sess.Clear()
	for _, key := range []string{key, key2} {
		actual = sess.Get(key)
		expected = ""
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`Session.Clear(); Session.Get(%#v) => %#v; want %#v`, key, actual, expected)
		}
	}
	actual = len(sess)
	expected = 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Clear(); len(Session) => %#v; want %#v`, actual, expected)
	}
}

func TestSession_Get(t *testing.T) {
	sess := make(kocha.Session)
	key := "test_key"
	var actual interface{} = sess.Get(key)
	var expected interface{} = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Get(%#v) => %#v; want %#v`, key, actual, expected)
	}

	value := "test_value"
	sess[key] = value
	actual = sess.Get(key)
	expected = value
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Get(%#v) => %#v; want %#v`, key, actual, expected)
	}

	delete(sess, key)
	actual = sess.Get(key)
	expected = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Get(%#v) => %#v; want %#v`, key, actual, expected)
	}
}

func TestSession_Set(t *testing.T) {
	sess := make(kocha.Session)
	key := "test_key"
	var actual interface{} = sess[key]
	var expected interface{} = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session[%#v] => %#v; want %#v`, key, actual, expected)
	}

	value := "test_value"
	sess.Set(key, value)
	actual = sess[key]
	expected = value
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); Session[%#v] => %#v; want %#v`, key, value, key, actual, expected)
	}

	value2 := "test_value2"
	sess.Set(key, value2)
	actual = sess[key]
	expected = value2
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Set(%#v, %#v); Session[%#v] => %#v; want %#v`, key, value2, key, actual, expected)
	}
}

func TestSession_Del(t *testing.T) {
	sess := make(kocha.Session)
	key := "test_key"
	value := "test_value"
	sess[key] = value
	var actual interface{} = sess[key]
	var expected interface{} = value
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session[%#v] => %#v; want %#v`, key, actual, expected)
	}

	sess.Del(key)
	actual = sess[key]
	expected = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Session.Del(%#v); Session[%#v] => %#v; want %#v`, key, key, actual, expected)
	}
}

func Test_Session_Clear(t *testing.T) {
	sess := make(kocha.Session)
	sess["hoge"] = "foo"
	sess["bar"] = "baz"
	actual := len(sess)
	expected := 2
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	sess.Clear()
	actual = len(sess)
	expected = 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_SessionCookieStore(t *testing.T) {
	if err := quick.Check(func(k, v string) bool {
		expected := make(kocha.Session)
		expected[k] = v
		store := kocha.NewTestSessionCookieStore()
		r, err := store.Save(expected)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := store.Load(r)
		if err != nil {
			t.Fatal(err)
		}
		return reflect.DeepEqual(actual, expected)
	}, nil); err != nil {
		t.Error(err)
	}

	func() {
		store := kocha.NewTestSessionCookieStore()
		key := "invalid"
		_, err := store.Load(key)
		actual := err
		expect := fmt.Errorf("kocha: session cookie value too short")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`SessionCookieStore.Load(%#v) => _, %#v; want %#v`, key, actual, expect)
		}
	}()
}

func Test_SessionCookieStore_Validate(t *testing.T) {
	// tests for validate the key size.
	for _, keySize := range []int{16, 24, 32} {
		store := &kocha.SessionCookieStore{
			SecretKey:  base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", keySize))),
			SigningKey: base64.StdEncoding.EncodeToString([]byte("a")),
		}
		if err := store.Validate(); err != nil {
			t.Errorf("Expect key size %v is valid, but returned error: %v", keySize, err)
		}
	}
	// boundary tests
	for _, keySize := range []int{15, 17, 23, 25, 31, 33} {
		store := &kocha.SessionCookieStore{
			SecretKey:  base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", keySize))),
			SigningKey: base64.StdEncoding.EncodeToString([]byte("a")),
		}
		if err := store.Validate(); err == nil {
			t.Errorf("Expect key size %v is invalid, but doesn't returned error", keySize)
		}
	}
}
