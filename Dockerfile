FROM golang:1.26.0-alpine

RUN apk add --no-cache git ca-certificates

# Development tools for local linting and debugging.
# The DevForge pipeline itself relies on:
#
#   golangci-lint
#   govulncheck
#
# Additional tools installed here may be used for development
# workflows but are not required by the CI pipeline.
RUN go install honnef.co/go/tools/cmd/staticcheck@v0.7.0 \
 && go install github.com/securego/gosec/v2/cmd/gosec@v2.24.0 \
 && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest \
 && go install golang.org/x/vuln/cmd/govulncheck@v1.1.4 \
 && go install github.com/pablogore/go-specs@v0.0.6

# Support two build contexts:
# - GoReleaser: context contains only Dockerfile + pre-built "forge" binary.
# - GitHub Action / local: context is the repo; no binary, so build from source.
COPY . /build
RUN set -e; \
  if [ -f /build/forge ] && [ ! -f /build/go.mod ]; then \
    cp /build/forge /usr/local/bin/forge; \
  else \
    (cd /build && go build -o /usr/local/bin/forge ./cmd/forge); \
  fi && \
  chmod +x /usr/local/bin/forge

ENTRYPOINT ["/usr/local/bin/forge"]