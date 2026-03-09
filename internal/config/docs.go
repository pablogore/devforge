// Package config provides repository-level configuration for DevForge.
//
// Configuration is read from .devforge.yml in the repository workdir.
// The file is optional; when absent, default values are used and behavior
// is unchanged. This package only loads and parses configuration; pipeline
// integration is done by other layers.
package config
