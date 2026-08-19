package driver

import "github.com/coditary/wuji/internal/capability"

// Info describes a registered driver.
type Info struct {
	ID           string
	Name         string
	Version      string
	Description  string
	Capabilities []capability.Type
	Remote       bool
	Endpoint     string
}

// Driver is the base interface every backend must implement.
type Driver interface {
	Info() Info
	Capabilities() []capability.Type
	Close() error
}
