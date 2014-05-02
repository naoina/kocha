package kocha

import (
	"fmt"
	"reflect"
	"strings"
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

func Test_SessionConfig_Validate(t *testing.T) {
	newSessionConfig := func() *SessionConfig {
		return &SessionConfig{
			Name: "testname",
		}
	}

	var config *SessionConfig
	if err := config.Validate(); err != nil {
		t.Errorf("Expect valid, but returned error")
	}

	config = newSessionConfig()
	if err := config.Validate(); err != nil {
		t.Errorf("Expect valid, but returned error: ", err)
	}

	config = newSessionConfig()
	config.Name = ""
	if err := config.Validate(); err == nil {
		t.Errorf("Expect invalid, but no error returned")
	}

	config = nil
	oldMiddlewares := appConfig.Middlewares
	appConfig.Middlewares = append(appConfig.Middlewares, &SessionMiddleware{})
	if err := config.Validate(); err == nil {
		t.Errorf("Expect invalid, but no error returned")
	}

	config = newSessionConfig()
	config.Store = nil
	if err := config.Validate(); err == nil {
		t.Errorf("Expect invalid, but no error returned")
	}

	store := &ValidateTestSessionStore{}
	config.Store = store
	if err := config.Validate(); err == nil {
		t.Errorf("Expect invalid, but no error returned")
	}
	if !store.validated {
		t.Errorf("Expect Validate() is called, but wasn't called")
	}
	appConfig.Middlewares = oldMiddlewares
}

type ValidateTestSessionStore struct{ validated bool }

func (s *ValidateTestSessionStore) Save(sess Session) string { return "" }
func (s *ValidateTestSessionStore) Load(key string) Session  { return nil }
func (s *ValidateTestSessionStore) Validate() error {
	s.validated = true
	return fmt.Errorf("")
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
		store := newTestSessionCookieStore()
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
		store := newTestSessionCookieStore()
		store.Load("invalid")
	}()
}

func Test_SessionCookieStore_Validate(t *testing.T) {
	// tests for validate the key size.
	for _, keySize := range []int{16, 24, 32} {
		store := &SessionCookieStore{
			SecretKey:  strings.Repeat("a", keySize),
			SigningKey: "a",
		}
		if err := store.Validate(); err != nil {
			t.Errorf("Expect key size %v is valid, but returned error: %v", keySize, err)
		}
	}
	// boundary tests
	for _, keySize := range []int{15, 17, 23, 25, 31, 33} {
		store := &SessionCookieStore{
			SecretKey:  strings.Repeat("a", keySize),
			SigningKey: "a",
		}
		if err := store.Validate(); err == nil {
			t.Errorf("Expect key size %v is invalid, but doesn't returned error", keySize)
		}
	}
}
