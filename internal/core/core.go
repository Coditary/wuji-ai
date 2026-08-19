package core

import (
	"context"
	"fmt"

	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
	"github.com/coditary/wuji/internal/driver/dummy"
)

// Config holds core initialization options.
type Config struct {
	DefaultDriverID string
}

// Core is the central orchestrator for all driver operations.
type Core struct {
	registry      *driver.Registry
	defaultDriver string
}

// New creates a Core instance with built-in drivers registered.
func New(cfg Config) (*Core, error) {
	c := &Core{
		registry:      driver.NewRegistry(),
		defaultDriver: cfg.DefaultDriverID,
	}

	if err := c.registry.Register(dummy.New()); err != nil {
		return nil, fmt.Errorf("register dummy driver: %w", err)
	}

	if c.defaultDriver == "" {
		c.defaultDriver = dummy.DriverID
	}

	if _, err := c.registry.Get(c.defaultDriver); err != nil {
		return nil, fmt.Errorf("default driver %q: %w", c.defaultDriver, err)
	}

	return c, nil
}

// Registry returns the driver registry.
func (c *Core) Registry() *driver.Registry {
	return c.registry
}

// DefaultDriverID returns the configured default driver.
func (c *Core) DefaultDriverID() string {
	return c.defaultDriver
}

// ListDrivers returns metadata for all registered drivers.
func (c *Core) ListDrivers() []driver.Info {
	return c.registry.List()
}

// GenerateText runs text generation using the given or default driver.
func (c *Core) GenerateText(ctx context.Context, driverID string, req driver.TextRequest) (*driver.TextResponse, error) {
	if driverID == "" {
		driverID = c.defaultDriver
	}

	d, err := c.registry.Get(driverID)
	if err != nil {
		return nil, err
	}

	if !d.HasCapability(capability.TextGeneration) {
		return nil, fmt.Errorf("driver %q does not support text generation", driverID)
	}

	return d.GenerateText(ctx, req)
}

// Close shuts down the core and all registered drivers.
func (c *Core) Close() error {
	return c.registry.Close()
}
