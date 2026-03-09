// Package git provides Git repository operations.
//
// Architectural Layer: adapters (git)
//
// This package implements the GitClient port interface using git CLI.
//
// Responsibility Boundaries:
//   - Get current branch name
//   - Get latest tag
//   - Get commits since tag
//   - Create tags
//   - Check working tree status
//   - Verify full git history availability
//   - Get HEAD and tag hashes
//
// Invariants:
//   - MUST implement ports.GitClient interface
//   - Contains ALL git CLI interaction
//   - Contains NO business logic
//
// This adapter is infrastructure - it handles all Git operations required
// by the application layer through a defined port interface.
package git
