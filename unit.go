package kocha

import "errors"

var (
	ErrInvokeDefault = errors.New("invoke default")
)

// Unit is an interface that Unit for FeatureToggle.
type Unit interface {
	// ActiveIf returns whether the Unit is active.
	ActiveIf() bool
}
