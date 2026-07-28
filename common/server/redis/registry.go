package redis

import (
	"fmt"
	"sync"
)

var (
	registry   = make(map[string]RueidisClient)
	registryMu sync.RWMutex
)

func Register(name string, c RueidisClient) error {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; exists {
		return fmt.Errorf("redis instance %q already registered", name)
	}
	registry[name] = c
	return nil
}

func Get(name string) (RueidisClient, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	c, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("redis instance %q not found", name)
	}
	return c, nil
}

// MustGet Retrieves the named instance; if it does not exist, it panics (used only during initialization)
func MustGet(name string) RueidisClient {
	c, err := Get(name)
	if err != nil {
		panic(err)
	}
	return c
}

// GetAll Returns all registered instances (for unified shutdown)
func GetAll() map[string]RueidisClient {
	registryMu.RLock()
	defer registryMu.RUnlock()

	result := make(map[string]RueidisClient, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}

// CloseAll Close all instances and clean registry
func CloseAll() {
	registryMu.Lock()
	defer registryMu.Unlock()

	for name, c := range registry {
		c.Close()
		delete(registry, name)
	}
}
