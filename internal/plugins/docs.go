// Package plugins discovers external plugin binaries (forge-plugin-<name>) in PATH.
// Plugins are standalone executables; DevForge runs them as pipeline steps after core steps.
// This package does not execute plugins; it only performs discovery.
package plugins
