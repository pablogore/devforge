# DevForge — build and validation targets

BINARY     := forge
MODULE     := github.com/pablogore/devforge
LDFLAGS    := -s -w -X main.version=dev -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
GOFLAGS   ?= -v

.PHONY: help build test lint lint-deps install clean run-pr run-doctor

help:
	@echo "Targets:"
	@echo "  build       Compila el binario forge (CGO_ENABLED=0, ldflags version/commit)"
	@echo "  test        Ejecuta go test ./..."
	@echo "  lint        Ejecuta golangci-lint run ./..."
	@echo "  lint-deps   Instala golangci-lint con la Go actual (necesario con Go 1.26 si falla buildir)"
	@echo "  install     Compila e instala forge con go install"
	@echo "  clean       Elimina el binario forge"
	@echo "  run-pr      Compila y ejecuta forge pr --mode full"
	@echo "  run-doctor  Compila y ejecuta forge doctor"

build:
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/forge

test:
	go test ./...

lint-deps:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint:
	golangci-lint run ./...

install: build
	go install -ldflags "$(LDFLAGS)" ./cmd/forge

clean:
	rm -f $(BINARY)

run-pr: build
	./$(BINARY) pr --mode full

run-doctor: build
	./$(BINARY) doctor
