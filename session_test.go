package kocha

import (
	"reflect"
	"testing"
	"testing/quick"
)

func Test_Constants(t *testing.T) {
	actual := SessionExpiresKey
	expected := "_kocha._sess._expires"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_Session_Clear(t *testing.T) {
	sess := make(Session)
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
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	if err := quick.Check(func(k, v string) bool {
		expected := make(Session)
		expected[k] = v
		store := &SessionCookieStore{}
		r := store.Save(expected)
		actual := store.Load(r)
		return reflect.DeepEqual(actual, expected)
	}, nil); err != nil {
		t.Error(err)
	}

	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Error("panic doesn't occurs")
			} else if _, ok := err.(ErrSession); !ok {
				t.Error("Expect %T, but %T", ErrSession{}, err)
			}
		}()
		store := &SessionCookieStore{}
		store.Load("invalid")
	}()
}

func Test_GenerateRandomKey(t *testing.T) {
	if err := quick.Check(func(length uint16) bool {
		already := make([][]byte, 0, 100)
		for i := 0; i < 100; i++ {
			buf := GenerateRandomKey(int(length))
			for _, v := range already {
				if !reflect.DeepEqual(buf, v) {
					return false
				}
			}
		}
		return true
	}, &quick.Config{MaxCount: 10}); err != nil {
		t.Error(err)
	}
}
