package doctor

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var errStopWalkLayout = errors.New("stop walk")

const adaptersImportPath = "internal/adapters"

// DetectAdapterImports reports whether any Go file under root/internal/domain imports internal/adapters.
// Domain should not depend on adapters (clean architecture).
func DetectAdapterImports(root string) bool {
	domainDir := filepath.Join(root, "internal", "domain")
	info, err := os.Stat(domainDir)
	if err != nil || !info.IsDir() {
		return false
	}
	found := false
	_ = filepath.Walk(domainDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
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
			if pathVal == adaptersImportPath || strings.HasPrefix(pathVal, adaptersImportPath+"/") {
				found = true
				return errStopWalkLayout
			}
		}
		return nil
	})
	return found
}
