package application

import "sort"

var stepRegistry = map[string]func() Step{}

// RegisterStep registers a step constructor by name. Called from init() by step implementations.
func RegisterStep(name string, ctor func() Step) {
	stepRegistry[name] = ctor
}

// GetStep returns a new instance of the step by name, or (nil, false) if unknown.
func GetStep(name string) (Step, bool) {
	ctor, ok := stepRegistry[name]
	if !ok {
		return nil, false
	}
	return ctor(), true
}

// ListSteps returns all registered step names in sorted order (for stable help and listing).
func ListSteps() []string {
	names := make([]string, 0, len(stepRegistry))
	for k := range stepRegistry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
