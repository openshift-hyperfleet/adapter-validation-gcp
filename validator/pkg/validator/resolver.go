package validator

import (
	"fmt"
	"sort"
)

// ExecutionGroup represents validators that can run in parallel
type ExecutionGroup struct {
	Level      int         // Execution level (0 = first, 1 = second, etc.)
	Validators []Validator // Validators to run in parallel at this level
}

// DependencyResolver builds execution plan from validators
type DependencyResolver struct {
	validators map[string]Validator
}

// NewDependencyResolver creates a new resolver
func NewDependencyResolver(validators []Validator) *DependencyResolver {
	m := make(map[string]Validator)
	for _, v := range validators {
		meta := v.Metadata()
		m[meta.Name] = v
	}
	return &DependencyResolver{validators: m}
}

// ResolveExecutionGroups organizes validators into parallel execution groups
// Validators with no dependencies or same dependencies can run in parallel
func (r *DependencyResolver) ResolveExecutionGroups() ([]ExecutionGroup, error) {
	// 1. Detect cycles
	if err := r.detectCycles(); err != nil {
		return nil, err
	}

	// 2. Topological sort with level assignment
	levels := r.assignLevels()

	// 3. Group by level
	groups := make([]ExecutionGroup, 0)
	for level := 0; ; level++ {
		var validators []Validator
		for _, v := range r.validators {
			meta := v.Metadata()
			if levels[meta.Name] == level {
				validators = append(validators, v)
			}
		}
		if len(validators) == 0 {
			break
		}

		// Sort alphabetically by name within the same level for deterministic execution
		sort.Slice(validators, func(i, j int) bool {
			return validators[i].Metadata().Name < validators[j].Metadata().Name
		})

		groups = append(groups, ExecutionGroup{
			Level:      level,
			Validators: validators,
		})
	}

	return groups, nil
}

// assignLevels performs topological sort and assigns execution levels
func (r *DependencyResolver) assignLevels() map[string]int {
	levels := make(map[string]int)

	// Recursive DFS to calculate max depth
	var calcLevel func(name string) int
	calcLevel = func(name string) int {
		if level, ok := levels[name]; ok {
			return level
		}

		v := r.validators[name]
		meta := v.Metadata()

		maxDepLevel := -1
		// Check dependencies from metadata
		for _, dep := range meta.RunAfter {
			if depValidator, exists := r.validators[dep]; exists {
				depLevel := calcLevel(depValidator.Metadata().Name)
				if depLevel > maxDepLevel {
					maxDepLevel = depLevel
				}
			}
		}
		// If RunAfter is empty, maxDepLevel stays -1, so level = 0

		level := maxDepLevel + 1
		levels[name] = level
		return level
	}

	for name := range r.validators {
		calcLevel(name)
	}

	return levels
}

// detectCycles detects circular dependencies using DFS
func (r *DependencyResolver) detectCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(name string) error
	dfs = func(name string) error {
		visited[name] = true
		recStack[name] = true

		v := r.validators[name]
		meta := v.Metadata()

		// Check all dependencies from metadata
		for _, dep := range meta.RunAfter {
			// Skip dependencies that don't exist (will be ignored in level assignment)
			if _, exists := r.validators[dep]; !exists {
				continue
			}

			if !visited[dep] {
				if err := dfs(dep); err != nil {
					return err
				}
			} else if recStack[dep] {
				return fmt.Errorf("circular dependency detected: %s -> %s", name, dep)
			}
		}

		recStack[name] = false
		return nil
	}

	for name := range r.validators {
		if !visited[name] {
			if err := dfs(name); err != nil {
				return err
			}
		}
	}

	return nil
}
