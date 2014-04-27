package kocha

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParams_From(t *testing.T) {
	for _, v := range []struct {
		name     string
		children []string
		expected string
	}{
		{"aaa", nil, "aaa"},
		{"abc", []string{}, "abc"},
		{"abc", []string{"def"}, "abc.def"},
		{"abc", []string{"def", "ghi"}, "abc.def.ghi"},
	} {
		p := &Params{}
		actual := p.From(v.name, v.children...).prefix
		expected := v.expected
		if actual != expected {
			t.Errorf("Params.From(%q, %q) => %q, want %q", v.name, v.children, actual, expected)
		}
	}
}

func TestFromParams_Bind(t *testing.T) {
	func() {
		type User struct{}
		p := &Params{Values: url.Values{}}
		user := User{}
		err := p.From("user").Bind(user)
		if err == nil {
			t.Errorf("From(%#v).Bind(%#v) => %#v, want error", "user", user, err)
		}
	}()

	func() {
		p := &Params{Values: url.Values{}}
		var s string
		err := p.From("user").Bind(&s)
		if err == nil {
			t.Errorf("From.Bind(%#v) => %#v, want error", s, err)
		}
	}()

	func() {
		type User struct {
			Name string
			Age  int
		}
		p := &Params{Values: url.Values{}}
		user := &User{}
		err := p.From("user").Bind(user)
		if err != nil {
			t.Errorf("From.Bind(%#v) => %#v, want nil", user, err)
		}

		actual := user
		expected := &User{}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%T => %#v, want %q", user, actual, expected)
		}
	}()

	func() {
		type User struct {
			Name    string
			Age     int
			Address string
		}
		p := &Params{Values: url.Values{
			"user.name":  {"naoina"},
			"user.age":   {"17"},
			"admin.name": {"administrator"},
		}}
		user := &User{}
		err := p.From("user").Bind(user, "name", "age")
		if err != nil {
			t.Errorf("From.Bind(%#v) => %#v, want nil", user, err)
		}

		actual := user
		expected := &User{
			Name:    "naoina",
			Age:     17,
			Address: "",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%T => %#v, want %#v", user, actual, expected)
		}
	}()

	func() {
		type User struct {
			Name string
		}
		type Admin struct {
			User
			Name string
		}
		p := &Params{Values: url.Values{
			"user.name": {"naoina"},
		}}
		admin := &Admin{}
		err := p.From("user").Bind(admin, "name")
		if err != nil {
			t.Errorf("From.Bind(%#v) => %#v, want nil", admin, err)
		}

		actual := admin
		expected := &Admin{
			Name: "naoina",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%T => %#v, want %#v", admin, actual, expected)
		}
	}()
}
