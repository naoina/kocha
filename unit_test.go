package kocha_test

import (
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

type testUnit struct {
	name      string
	active    bool
	callCount int
}

func (u *testUnit) ActiveIf() bool {
	u.callCount++
	return u.active
}

type testUnit2 struct{}

func (u *testUnit2) ActiveIf() bool {
	return true
}

func TestInvoke(t *testing.T) {
	// test that it invokes newFunc when ActiveIf returns true.
	testInvokeWrapper(func() {
		unit := &testUnit{"test1", true, 0}
		called := false
		kocha.Invoke(unit, func() {
			called = true
		}, func() {
			t.Errorf("defaultFunc has been called")
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	// test that it invokes defaultFunc when ActiveIf returns false.
	testInvokeWrapper(func() {
		unit := &testUnit{"test2", false, 0}
		called := false
		kocha.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			called = true
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	// test that it invokes defaultFunc when any errors occurred in newFunc.
	testInvokeWrapper(func() {
		unit := &testUnit{"test3", true, 0}
		called := false
		kocha.Invoke(unit, func() {
			panic("expected error")
		}, func() {
			called = true
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	// test that it will be panic when panic occurred in defaultFunc.
	testInvokeWrapper(func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't occurred")
			} else if err != "expected error in defaultFunc" {
				t.Errorf("panic doesn't occurred in defaultFunc: %v", err)
			}
		}()
		unit := &testUnit{"test4", false, 0}
		kocha.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			panic("expected error in defaultFunc")
		})
	})

	// test that it panic when panic occurred in both newFunc and defaultFunc.
	testInvokeWrapper(func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't occurred")
			} else if err != "expected error in defaultFunc" {
				t.Errorf("panic doesn't occurred in defaultFunc: %v", err)
			}
		}()
		unit := &testUnit{"test5", true, 0}
		called := false
		kocha.Invoke(unit, func() {
			called = true
			panic("expected error")
		}, func() {
			panic("expected error in defaultFunc")
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})

	testInvokeWrapper(func() {
		unit := &testUnit{"test6", true, 0}
		kocha.Invoke(unit, func() {
			panic("expected error")
		}, func() {
			// do nothing.
		})
		var actual interface{} = unit.callCount
		var expected interface{} = 1
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}

		// again.
		kocha.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			// do nothing.
		})
		actual = unit.callCount
		expected = 1
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}

		// same unit type.
		unit = &testUnit{"test7", true, 0}
		called := false
		kocha.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			called = true
		})
		actual = called
		expected = true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}

		// different unit type.
		unit2 := &testUnit2{}
		called = false
		kocha.Invoke(unit2, func() {
			called = true
		}, func() {
			t.Errorf("defaultFunc has been called")
		})
		actual = called
		expected = true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	})
}
