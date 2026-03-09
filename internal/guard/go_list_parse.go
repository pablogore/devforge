package guard

import (
	"bytes"
	"encoding/json"
	"strings"
)

// goListPackage is a minimal subset of go list -json output for import checks.
type goListPackage struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

// pathContainsSegment reports whether path contains segment as a path component
// (avoids false positives like "internal/domainutils" matching "internal/domain").
func pathContainsSegment(path, segment string) bool {
	if segment == "" {
		return false
	}
	if path == segment {
		return true
	}
	suffix := "/" + segment
	if strings.HasSuffix(path, suffix) {
		return true
	}
	return strings.Contains(path, suffix+"/")
}

// checkImportsContain parses go list -json output (one or more concatenated JSON objects)
// and returns errOnMatch if any package's Imports contains the forbidden path segment.
// Matching is segment-based to avoid false positives (e.g. internal/domainutils vs internal/domain).
func checkImportsContain(goListOutput, forbiddenSegment string, errOnMatch error) error {
	dec := json.NewDecoder(bytes.NewBufferString(strings.TrimSpace(goListOutput)))
	for {
		var p goListPackage
		if err := dec.Decode(&p); err != nil {
			break
		}
		for _, imp := range p.Imports {
			if pathContainsSegment(imp, forbiddenSegment) {
				return errOnMatch
			}
		}
	}
	return nil
}
