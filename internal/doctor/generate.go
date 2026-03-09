package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	yaml "gopkg.in/yaml.v3"
)

const (
	policiesDir     = ".devforge/policies"
	archRelPath     = ".devforge/policies/architecture.yaml"
	securityRelPath = ".devforge/policies/security.yaml"
)

// generatedPolicy is the structure written to .devforge/policies/*.yaml.
// Rules are emitted in sorted key order for deterministic output.
type generatedPolicy struct {
	Name  string                 `yaml:"name"`
	Type  string                 `yaml:"type"`
	Rules map[string]interface{} `yaml:"rules"`
}

// GeneratePolicies analyzes the repository at root, derives policy rules from
// existing doctor checks (DetectAdapterImports, DetectDangerousImports,
// DetectTimeNowUsage), and writes architecture.yaml and security.yaml under
// .devforge/policies/. Returns the list of generated file paths, or nil if
// no issues were detected (no policies needed). Overwrites existing files.
// Output is deterministic: rules are sorted, YAML is stable.
func GeneratePolicies(root string) ([]string, error) {
	dangerous := DetectDangerousImports(root)
	timeNow := DetectTimeNowUsage(root)
	adapterImports := DetectAdapterImports(root)

	var archRules map[string]interface{}
	var securityRules map[string]interface{}

	if adapterImports {
		if archRules == nil {
			archRules = make(map[string]interface{})
		}
		archRules["forbid_import"] = "internal/adapters"
	}
	if timeNow {
		if archRules == nil {
			archRules = make(map[string]interface{})
		}
		archRules["forbid_time_now"] = "domain"
	}
	if len(dangerous) > 0 {
		securityRules = make(map[string]interface{})
		sort.Strings(dangerous)
		if len(dangerous) == 1 {
			securityRules["forbid_import"] = dangerous[0]
		} else {
			vals := make([]interface{}, len(dangerous))
			for i, s := range dangerous {
				vals[i] = s
			}
			securityRules["forbid_import"] = vals
		}
	}

	var generated []string
	dir := filepath.Join(root, policiesDir)
	if len(archRules) > 0 || len(securityRules) > 0 {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, fmt.Errorf("create %s: %w", dir, err)
		}
	}

	if len(archRules) > 0 {
		p := generatedPolicy{Name: "architecture", Type: "architecture", Rules: archRules}
		absPath := filepath.Join(dir, "architecture.yaml")
		if err := writePolicyFile(absPath, &p); err != nil {
			return nil, err
		}
		generated = append(generated, archRelPath)
	}
	if len(securityRules) > 0 {
		p := generatedPolicy{Name: "security", Type: "security", Rules: securityRules}
		absPath := filepath.Join(dir, "security.yaml")
		if err := writePolicyFile(absPath, &p); err != nil {
			return nil, err
		}
		generated = append(generated, securityRelPath)
	}

	return generated, nil
}

// writePolicyFile writes the policy to path with deterministic YAML (sorted rule keys).
func writePolicyFile(path string, p *generatedPolicy) error {
	node, err := policyToNode(p)
	if err != nil {
		return fmt.Errorf("encode policy: %w", err)
	}
	out, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}
	if err := os.WriteFile(path, out, 0600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// policyToNode builds a yaml.Node for the policy with rules in sorted key order.
func policyToNode(p *generatedPolicy) (*yaml.Node, error) {
	root := &yaml.Node{Kind: yaml.MappingNode}
	root.Content = append(root.Content,
		scalarNode("name"), scalarNode(p.Name),
		scalarNode("type"), scalarNode(p.Type),
	)
	// Rules mapping with sorted keys
	keys := make([]string, 0, len(p.Rules))
	for k := range p.Rules {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rulesNode := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range keys {
		v := p.Rules[k]
		rulesNode.Content = append(rulesNode.Content, scalarNode(k))
		valNode, err := valueToNode(v)
		if err != nil {
			return nil, err
		}
		rulesNode.Content = append(rulesNode.Content, valNode)
	}
	root.Content = append(root.Content, scalarNode("rules"), rulesNode)
	return root, nil
}

func scalarNode(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: s}
}

func valueToNode(v interface{}) (*yaml.Node, error) {
	switch x := v.(type) {
	case string:
		return scalarNode(x), nil
	case []interface{}:
		node := &yaml.Node{Kind: yaml.SequenceNode}
		for _, e := range x {
			child, err := valueToNode(e)
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content, child)
		}
		return node, nil
	default:
		// fallback: marshal then unmarshal to node
		var n yaml.Node
		if err := n.Encode(v); err != nil {
			return nil, err
		}
		return &n, nil
	}
}
