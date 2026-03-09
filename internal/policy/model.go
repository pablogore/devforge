package policy

// Policy is a single policy file loaded from .syntegrity/policies/*.yaml.
// File is set by the loader to the policy filename so errors can identify the source.
// Rules supports both single values (string) and multiple values ([]interface{}) per rule.
// Severity is optional: "error" (default) fails the pipeline on violation; "warning" logs but does not fail.
type Policy struct {
	File     string
	Name     string                 `yaml:"name"`
	Type     string                 `yaml:"type"`
	Severity string                 `yaml:"severity"`
	Rules    map[string]interface{} `yaml:"rules"`
}
