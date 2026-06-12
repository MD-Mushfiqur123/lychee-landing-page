package server

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// ModelAliases manages mapping virtual alias names to real model names.
type ModelAliases struct {
	storePath string
	aliases   map[string]string // map[alias_name]target_model_name
	mu        sync.RWMutex
}

// NewModelAliases initializes a new ModelAliases manager.
func NewModelAliases(storePath string) *ModelAliases {
	ma := &ModelAliases{
		storePath: storePath,
		aliases:   make(map[string]string),
	}
	_ = ma.load()
	return ma
}

func (ma *ModelAliases) load() error {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	data, err := os.ReadFile(ma.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &ma.aliases)
}

func (ma *ModelAliases) save() error {
	dir := filepath.Dir(ma.storePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(ma.aliases, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ma.storePath, data, 0644)
}

// Set adds or updates a model alias mapping.
func (ma *ModelAliases) Set(name, target string) error {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	if name == "" {
		return errors.New("alias name cannot be empty")
	}
	if target == "" {
		return errors.New("target model cannot be empty")
	}

	ma.aliases[name] = target
	return ma.save()
}

// Resolve resolves an alias (handling cycle detection).
func (ma *ModelAliases) Resolve(name string) string {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	current := name
	visited := make(map[string]bool)
	for {
		if target, ok := ma.aliases[current]; ok {
			if visited[target] {
				break
			}
			visited[current] = true
			current = target
		} else {
			break
		}
	}
	return current
}

// Delete removes an alias mapping.
func (ma *ModelAliases) Delete(name string) error {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	if _, ok := ma.aliases[name]; !ok {
		return errors.New("alias not found")
	}

	delete(ma.aliases, name)
	return ma.save()
}

// List returns a copy of all alias mappings.
func (ma *ModelAliases) List() map[string]string {
	ma.mu.RLock()
	defer ma.mu.RUnlock()

	res := make(map[string]string)
	for k, v := range ma.aliases {
		res[k] = v
	}
	return res
}
