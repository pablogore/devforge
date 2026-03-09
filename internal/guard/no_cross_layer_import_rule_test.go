package guard

import (
	"context"
	"testing"

	"github.com/pablogore/go-specs/specs"
)

func TestNoCrossLayerImportRule(t *testing.T) {
	domainOK := `{"ImportPath":"github.com/foo/internal/domain","Imports":["fmt"]}`
	appOK := `{"ImportPath":"github.com/foo/internal/application","Imports":["github.com/foo/internal/domain","github.com/foo/internal/ports"]}`
	portsOK := `{"ImportPath":"github.com/foo/internal/ports","Imports":[]}`

	domainImportsApp := `{"ImportPath":"github.com/foo/internal/domain","Imports":["github.com/foo/internal/application"]}`
	domainImportsAdapters := `{"ImportPath":"github.com/foo/internal/domain","Imports":["github.com/foo/internal/adapters"]}`
	appImportsAdapters := `{"ImportPath":"github.com/foo/internal/application","Imports":["github.com/foo/internal/adapters"]}`
	portsImportsDomain := `{"ImportPath":"github.com/foo/internal/ports","Imports":["github.com/foo/internal/domain"]}`
	portsImportsApp := `{"ImportPath":"github.com/foo/internal/ports","Imports":["github.com/foo/internal/application"]}`
	portsImportsAdapters := `{"ImportPath":"github.com/foo/internal/ports","Imports":["github.com/foo/internal/adapters"]}`

	specs.Describe(t, "NoCrossLayerImportRule", func(s *specs.Spec) {
		s.It("Name returns NoCrossLayerImports", func(ctx *specs.Context) {
			r := NewNoCrossLayerImportRule()
			ctx.Expect(r.Name()).ToEqual("NoCrossLayerImports")
		})
		s.It("Validate covers cross-layer paths", func(ctx *specs.Context) {
			tests := []struct {
				name      string
				domainOut string
				domainErr error
				appOut    string
				appErr    error
				portsOut  string
				portsErr  error
				wantErr   error
			}{
				{"all pass", domainOK, nil, appOK, nil, portsOK, nil, nil},
				{"app empty skips check", domainOK, nil, "", nil, portsOK, nil, nil},
				{"domain imports application", domainImportsApp, nil, "", nil, "", nil, errCrossLayerDomain},
				{"domain imports adapters", domainImportsAdapters, nil, "", nil, "", nil, errCrossLayerDomain},
				{"application imports adapters", domainOK, nil, appImportsAdapters, nil, portsOK, nil, errCrossLayerApp},
				{"ports imports domain", domainOK, nil, appOK, nil, portsImportsDomain, nil, errCrossLayerPorts},
				{"ports imports application", domainOK, nil, appOK, nil, portsImportsApp, nil, errCrossLayerPorts},
				{"ports imports adapters", domainOK, nil, appOK, nil, portsImportsAdapters, nil, errCrossLayerPorts},
			}
			for _, tt := range tests {
				runner := &fakeRunnerFunc{
					f: func(_, _ string, args ...string) (string, error) {
						if len(args) >= 3 && args[0] == "list" && args[1] == "-json" {
							switch args[2] {
							case "./internal/domain/...":
								return tt.domainOut, tt.domainErr
							case "./internal/application/...":
								return tt.appOut, tt.appErr
							case "./internal/ports/...":
								return tt.portsOut, tt.portsErr
							}
						}
						return "", nil
					},
				}
				gCtx := &Context{StdCtx: context.Background(), Workdir: "/wd", CommandRunner: runner}
				r := NewNoCrossLayerImportRule()
				got := r.Validate(gCtx)
				if tt.wantErr != nil {
					ctx.Expect(got != nil).To(specs.BeTrue())
					ctx.Expect(got == tt.wantErr).To(specs.BeTrue())
				} else {
					ctx.Expect(got).To(specs.BeNil())
				}
			}
		})
	})
}

type fakeRunnerFunc struct {
	f func(dir, name string, args ...string) (string, error)
}

func (f *fakeRunnerFunc) Run(_ context.Context, dir, name string, args ...string) (string, error) {
	return f.f(dir, name, args...)
}

func (f *fakeRunnerFunc) RunCombinedOutput(_ context.Context, dir, name string, args ...string) (string, error) {
	return f.f(dir, name, args...)
}

func (f *fakeRunnerFunc) RunCombinedOutputWithEnv(_ context.Context, dir string, _ []string, name string, args ...string) (string, error) {
	return f.f(dir, name, args...)
}
