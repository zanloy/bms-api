package models

import (
	"fmt"
	"sort"
)

// A Tenant is a tenant of our multi-tenant environment. They will have a name
// and a list of environments. The format for the namespace of any tenant/env
// combo would be '${name}-${env}'
type Tenant struct {
	Name string
	Envs []string
}

// AddEnv will add the env to the Tenant if it doesn't already exist.
func (t *Tenant) AddEnv(env string) {
	if !t.HasEnv(env) {
		t.Envs = append(t.Envs, env)
		sort.Strings(t.Envs)
	}
}

// DeleteEnv will delete the env from the Tenant or return error if not found.
func (t *Tenant) DeleteEnv(env string) error {
	idx := t.findEnv(env)
	if idx == -1 {
		return fmt.Errorf("The tenant [%s] does not have the environment [%s].", t.Name, env)
	}
	t.Envs[idx] = t.Envs[len(t.Envs)-1] // Copy last element to "deleted" one
	t.Envs = t.Envs[:len(t.Envs)-1]     // Chop last element off slice
	sort.Strings(t.Envs)                // Re-sort our environments
	return nil
}

// HasEnv will iterate over Envs and return true if the env exists.
func (t *Tenant) HasEnv(env string) bool {
	return t.findEnv(env) != -1 // Our index is -1 for a failure.
}

// findEnv will iterate over the Envs and return the index of the element of -1
// if the env doesn't exist.
func (t *Tenant) findEnv(env string) int {
	// Iterate all Envs, we do not assume a long list so this is KISS over a binary search.
	for idx, e := range t.Envs {
		if e == env {
			return idx
		}
	}
	return -1
}
