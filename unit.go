package kocha

import (
	"errors"
	"reflect"
)

var (
	ErrInvokeDefault = errors.New("invoke default")
)

var failedUnits = make(map[string]bool)

// Unit is an interface that Unit for FeatureToggle.
type Unit interface {
	// ActiveIf returns whether the Unit is active.
	ActiveIf() bool
}

// Invoke invokes newFunc.
// It invokes newFunc but will behave to fallback.
// When unit.ActiveIf returns false or any errors occurred in invoking, it invoke the defaultFunc if defaultFunc isn't nil.
// Also if any errors occurred at least once, next invoking will always invoke the defaultFunc.
func Invoke(unit Unit, newFunc func(), defaultFunc func()) {
	name := reflect.TypeOf(unit).String()
	defer func() {
		if err := recover(); err != nil {
			if err != ErrInvokeDefault {
				logStackAndError(err)
				failedUnits[name] = true
			}
			if defaultFunc != nil {
				defaultFunc()
			}
		}
	}()
	if failedUnits[name] {
		panic(ErrInvokeDefault)
	}
	if !unit.ActiveIf() {
		panic(ErrInvokeDefault)
	}
	newFunc()
}
