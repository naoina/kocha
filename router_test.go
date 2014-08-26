package kocha_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestRouter_Reverse(t *testing.T) {
	app := kocha.NewTestApp()
	for _, v := range []struct {
		name   string
		args   []interface{}
		expect string
	}{
		{"root", []interface{}{}, "/"},
		{"user", []interface{}{77}, "/user/77"},
		{"date", []interface{}{2013, 10, 26, "naoina"}, "/2013/10/26/user/naoina"},
		{"static", []interface{}{"/hoge.png"}, "/static/hoge.png"},
		{"static", []interface{}{"hoge.png"}, "/static/hoge.png"},
	} {
		r, err := app.Router.Reverse(v.name, v.args...)
		if err != nil {
			t.Errorf(`Router.Reverse(%#v, %#v) => (_, %#v); want (_, %#v)`, v.name, v.args, err, err)
			continue
		}
		actual := r
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`Router.Reverse(%#v, %#v) => (%#v, %#v); want (%#v, %#v)`, v.name, v.args, actual, err, expect, err)
		}
	}
}

func TestRouter_Reverse_withUnknownRouteName(t *testing.T) {
	app := kocha.NewTestApp()
	name := "unknown"
	_, err := app.Router.Reverse(name)
	actual := err
	expect := fmt.Errorf("kocha: no match route found: %s ()", name)
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("Router.Reverse(%#v) => (_, %#v); want (_, %#v)", name, actual, expect)
	}
}

func TestRouter_Reverse_withFewArguments(t *testing.T) {
	app := kocha.NewTestApp()
	name := "user"
	_, err := app.Router.Reverse(name)
	actual := err
	expect := fmt.Errorf("kocha: too few arguments: %s (controller is %T)", name, &kocha.FixtureUserTestCtrl{})
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`Router.Reverse(%#v) => (_, %#v); want (_, %#v)`, name, actual, expect)
	}
}

func TestRouter_Reverse_withManyArguments(t *testing.T) {
	app := kocha.NewTestApp()
	name := "user"
	args := []interface{}{77, 100}
	_, err := app.Router.Reverse(name, args...)
	actual := err
	expect := fmt.Errorf("kocha: too many arguments: %s (controller is %T)", name, &kocha.FixtureUserTestCtrl{})
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`Router.Reverse(%#v, %#v) => (_, %#v); want (_, %#v)`, name, args, actual, expect)
	}
}
