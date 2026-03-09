package application

import "fmt"

// RunSteps runs the named steps in order using the given context and runner. Returns on first error.
func RunSteps(ctx *Context, runner *StepRunner, names []string) error {
	for _, name := range names {
		step, ok := GetStep(name)
		if !ok {
			return fmt.Errorf("unknown step: %s", name)
		}

		if err := runner.Run(ctx, step); err != nil {
			return err
		}
	}
	return nil
}
