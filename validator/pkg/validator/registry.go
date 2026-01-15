package validator

import (
	"fmt"
	"sync"
)

// Registry holds all registered validators
var globalRegistry = NewRegistry()

type Registry struct {
	mu         sync.RWMutex
	validators map[string]Validator
}

// NewRegistry creates a new validator registry
func NewRegistry() *Registry {
	return &Registry{
		validators: make(map[string]Validator),
	}
}

// Register adds a validator to the registry
func (r *Registry) Register(v Validator) {
	r.mu.Lock()
	defer r.mu.Unlock()

	meta := v.Metadata()
	// Allow overwriting for testing purposes
	r.validators[meta.Name] = v
}

// GetAll returns all registered validators
func (r *Registry) GetAll() []Validator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	validators := make([]Validator, 0, len(r.validators))
	for _, v := range r.validators {
		validators = append(validators, v)
	}
	return validators
}

// Get retrieves a validator by name
func (r *Registry) Get(name string) (Validator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.validators[name]
	return v, ok
}

// Package-level functions for global registry

// Register adds a validator to the global registry
// This is called from init() functions in validator implementations
func Register(v Validator) {
	meta := v.Metadata()
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if _, exists := globalRegistry.validators[meta.Name]; exists {
		panic(fmt.Sprintf("validator already registered: %s", meta.Name))
	}
	globalRegistry.validators[meta.Name] = v
}

// GetAll returns all registered validators from global registry
func GetAll() []Validator {
	return globalRegistry.GetAll()
}

// Get retrieves a validator by name from global registry
func Get(name string) (Validator, bool) {
	return globalRegistry.Get(name)
}

// ClearRegistry clears all validators from the global registry (for testing)
func ClearRegistry() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.validators = make(map[string]Validator)
}
