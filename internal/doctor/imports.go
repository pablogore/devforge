package doctor

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Dangerous import paths that are commonly forbidden in production (e.g. security or debug).
var dangerousImportCandidates = []string{
	"net/http/pprof",
}

// DetectDangerousImports searches Go files under root for imports that match dangerous candidates.
// Returns a list of import paths that were found (to suggest forbid_import rules).
func DetectDangerousImports(root string) []string {
	var found []string
	seen := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil
		}
		for _, imp := range f.Imports {
			pathVal := strings.Trim(imp.Path.Value, `"`)
			for _, candidate := range dangerousImportCandidates {
				if pathVal == candidate && !seen[pathVal] {
					seen[pathVal] = true
					found = append(found, pathVal)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return found
}
