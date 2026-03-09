package application

import "fmt"

var pipelineRegistry = map[string]Pipeline{}

// RegisterPipeline stores a pipeline by name. Overwrites if the name already exists.
func RegisterPipeline(p Pipeline) {
	pipelineRegistry[p.Name] = p
}

// GetPipeline returns the pipeline with the given name. The boolean is false if not found.
func GetPipeline(name string) (Pipeline, bool) {
	p, ok := pipelineRegistry[name]
	return p, ok
}

// ListPipelines returns the names of all registered pipelines (order unspecified).
func ListPipelines() []string {
	names := make([]string, 0, len(pipelineRegistry))
	for k := range pipelineRegistry {
		names = append(names, k)
	}
	return names
}

// RunPipeline retrieves the pipeline by name and runs its steps with the given context and runner.
// Returns an error if the pipeline is unknown or if any step fails.
func RunPipeline(name string, ctx *Context, runner *StepRunner) error {
	p, ok := GetPipeline(name)
	if !ok {
		return fmt.Errorf("unknown pipeline: %s", name)
	}
	return p.Run(ctx, runner)
}
