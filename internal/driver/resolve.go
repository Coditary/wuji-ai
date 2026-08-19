package driver

import (
	"fmt"

	"github.com/coditary/wuji/internal/capability"
)

// HasCapability reports whether a driver advertises the given capability.
func HasCapability(d Driver, c capability.Type) bool {
	for _, cap := range d.Capabilities() {
		if cap == c {
			return true
		}
	}
	return false
}

// As returns the driver cast to the requested capability interface.
func As[T any](d Driver, c capability.Type) (T, error) {
	var zero T
	if !HasCapability(d, c) {
		return zero, fmt.Errorf("driver %q does not support %s", d.Info().ID, c)
	}

	v, ok := d.(T)
	if !ok {
		return zero, fmt.Errorf("driver %q does not implement %T", d.Info().ID, zero)
	}
	return v, nil
}
