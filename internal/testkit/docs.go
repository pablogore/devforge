// Package testkit provides deterministic fakes for ports (Logger, CommandRunner, Clock, GitClient, EnvProvider)
// for use in application and usecase tests. Fakes store inputs and return configured outputs;
// no expectation APIs (On/Return/AssertExpectations). Use internal/testkit only from test code.
package testkit
