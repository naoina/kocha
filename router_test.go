package kocha_test

import (
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestRouter_Reverse(t *testing.T) {
	app := kocha.NewTestApp()
	actual := app.Router.Reverse("root")
	expected := "/"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = app.Router.Reverse("user", 77)
	expected = "/user/77"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = app.Router.Reverse("date", 2013, 10, 26, "naoina")
	expected = "/2013/10/26/user/naoina"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	for _, v := range []string{"/hoge.png", "hoge.png"} {
		actual = app.Router.Reverse("static", v)
		expected = "/static/hoge.png"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}
}

func TestRouter_Reverse_withUnknownRouteName(t *testing.T) {
	app := kocha.NewTestApp()
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	app.Router.Reverse("unknown")
}

func TestRouter_Reverse_withFewArguments(t *testing.T) {
	app := kocha.NewTestApp()
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	app.Router.Reverse("user")
}

func TestRouter_Reverse_withManyArguments(t *testing.T) {
	app := kocha.NewTestApp()
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("panic doesn't happened")
		}
	}()
	app.Router.Reverse("user", 77, 100)
}
