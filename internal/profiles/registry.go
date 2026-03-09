package profiles

import "sort"

var registry = map[string]Profile{}

// Register adds a profile to the registry. Called from init() by profile implementations.
func Register(p Profile) {
	registry[p.Name] = p
}

// Get returns the profile by name and true if found, or a zero Profile and false otherwise.
func Get(name string) (Profile, bool) {
	p, ok := registry[name]
	return p, ok
}

// List returns all registered profile names in sorted order (for stable help text).
func List() []string {
	names := make([]string, 0, len(registry))
	for k := range registry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
