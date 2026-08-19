package driver

import (
	"fmt"
	"sync"

	"github.com/coditary/wuji/internal/capability"
)

// Registry holds all available drivers keyed by ID.
type Registry struct {
	mu      sync.RWMutex
	drivers map[string]Driver
}

func NewRegistry() *Registry {
	return &Registry{
		drivers: make(map[string]Driver),
	}
}

func (r *Registry) Register(d Driver) error {
	info := d.Info()
	if info.ID == "" {
		return fmt.Errorf("driver ID must not be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.drivers[info.ID]; exists {
		return fmt.Errorf("driver %q already registered", info.ID)
	}

	r.drivers[info.ID] = d
	return nil
}

func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.drivers[id]
	if !ok {
		return fmt.Errorf("driver %q not found", id)
	}

	if err := d.Close(); err != nil {
		return fmt.Errorf("close driver %q: %w", id, err)
	}

	delete(r.drivers, id)
	return nil
}

func (r *Registry) Get(id string) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.drivers[id]
	if !ok {
		return nil, fmt.Errorf("driver %q not found", id)
	}
	return d, nil
}

func (r *Registry) List() []Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]Info, 0, len(r.drivers))
	for _, d := range r.drivers {
		infos = append(infos, d.Info())
	}
	return infos
}

func (r *Registry) FindByCapability(c capability.Type) []Driver {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Driver
	for _, d := range r.drivers {
		if d.HasCapability(c) {
			result = append(result, d)
		}
	}
	return result
}

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for id, d := range r.drivers {
		if err := d.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close driver %q: %w", id, err))
		}
	}
	r.drivers = make(map[string]Driver)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing drivers: %v", errs)
	}
	return nil
}
