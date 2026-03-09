package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pablogore/devforge/internal/application"
)

func init() {
	application.RegisterStep("integration-tests", func() application.Step { return IntegrationTestsStep{} })
}

// IntegrationTestsStep runs integration tests (go test -tags=integration). Prefers ./integrationtest/... if the directory exists.
// If no Go files in the repository contain //go:build integration, the step is skipped and returns success.
type IntegrationTestsStep struct{}

// Name returns the step name for logging and registry.
func (IntegrationTestsStep) Name() string {
	return "integration-tests"
}

// hasIntegrationBuildTag reports whether any .go file under root contains a build constraint for integration.
func hasIntegrationBuildTag(root string) bool {
	found := false
	_ = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
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
		//nolint:gosec // G304: path is from filepath.Walk(root), under workdir only
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		if strings.Contains(content, "//go:build integration") || strings.Contains(content, "// +build integration") {
			found = true
			return errStopWalk
		}
		return nil
	})
	return found
}

var errStopWalk = fmt.Errorf("stop walk")

// Run executes go test -tags=integration. If integrationtest/ exists, runs ./integrationtest/...; otherwise ./...
// If no Go files have //go:build integration, skips and returns nil.
func (IntegrationTestsStep) Run(ctx *application.Context) error {
	if !hasIntegrationBuildTag(ctx.Workdir) {
		ctx.Log.Info("Skipping integration tests (no //go:build integration found)", "step", "integration-tests")
		return nil
	}
	integrationtestDir := filepath.Join(ctx.Workdir, "integrationtest")
	pkg := "./..."
	if info, err := os.Stat(integrationtestDir); err == nil && info.IsDir() {
		pkg = "./integrationtest/..."
	}
	ctx.Log.Info("Running integration tests", "step", "integration-tests", "pkg", pkg)
	out, err := ctx.Cmd.RunCombinedOutput(ctx.StdCtx, ctx.Workdir, "go", "test", "-tags=integration", "-count=1", pkg)
	if err != nil {
		return fmt.Errorf("integration tests failed: %w\n%s", err, out)
	}
	return nil
}
