package steps

import (
	"github.com/pablogore/devforge/internal/application"
	"github.com/pablogore/devforge/internal/plugins"
)

// DiscoveredPluginSteps returns one ExternalPluginStep per plugin discovered in PATH,
// in discovery order. Used to append plugin steps after core pipeline steps.
func DiscoveredPluginSteps() []application.Step {
	names := plugins.Discover()
	if len(names) == 0 {
		return nil
	}
	out := make([]application.Step, len(names))
	for i, name := range names {
		out[i] = &ExternalPluginStep{name: name}
	}
	return out
}

func init() {
	for _, name := range plugins.Discover() {
		n := name
		application.RegisterStep("plugin-"+n, func() application.Step {
			return &ExternalPluginStep{name: n}
		})
	}
}
