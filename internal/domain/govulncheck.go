package domain

import (
	"encoding/json"
	"strings"
)

// GovulncheckSeverity levels that fail the PR.
const (
	GovulncheckSeverityHigh     = "HIGH"
	GovulncheckSeverityCritical = "CRITICAL"
)

// ValidateGovulncheckOutput parses govulncheck -json output and returns an error
// if any vulnerability has severity HIGH or CRITICAL. Returns nil when no
// HIGH/CRITICAL findings are present. Parsing is pure: no IO, no shell.
func ValidateGovulncheckOutput(jsonOutput string) error {
	lines := strings.Split(strings.TrimSpace(jsonOutput), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if severity := extractSeverity(msg); severity != "" {
			s := strings.ToUpper(severity)
			if s == GovulncheckSeverityHigh || s == GovulncheckSeverityCritical {
				return ErrGovulncheckHighOrCritical
			}
		}
	}
	return nil
}

func extractSeverity(m map[string]interface{}) string {
	if s, _ := m["Severity"].(string); s != "" {
		return s
	}
	if s, _ := m["severity"].(string); s != "" {
		return s
	}
	if v, ok := m["Vuln"]; ok {
		if vm, _ := v.(map[string]interface{}); vm != nil {
			if s := extractSeverity(vm); s != "" {
				return s
			}
			if db, _ := vm["database_specific"].(map[string]interface{}); db != nil {
				if s, _ := db["severity"].(string); s != "" {
					return s
				}
			}
		}
	}
	if osv, ok := m["osv"]; ok {
		if om, _ := osv.(map[string]interface{}); om != nil {
			if db, _ := om["database_specific"].(map[string]interface{}); db != nil {
				if s, _ := db["severity"].(string); s != "" {
					return s
				}
			}
		}
	}
	return ""
}
