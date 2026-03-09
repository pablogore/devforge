package doctor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var errStopWalk = errors.New("stop walk")

const timeNowLiteral = "time.Now()"

// DetectTimeNowUsage reports whether any Go file under root/internal/domain contains time.Now().
// Domain logic should use injected clocks for determinism.
func DetectTimeNowUsage(root string) bool {
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
		//nolint:gosec // G304: path is from filepath.Walk under root/internal/domain; no user input
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.Contains(string(data), timeNowLiteral) {
			found = true
			return errStopWalk
		}
		return nil
	})
	return found
}
