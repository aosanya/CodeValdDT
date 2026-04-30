.PHONY: build build-server build-dev server dev dev-restart restart kill proto test cover test-arango test-all vet lint clean

export PATH := /usr/local/go/bin:$(PATH)

# ── Build ─────────────────────────────────────────────────────────────────────

## Verify the module compiles cleanly.
build:
	go build ./...

## Build the production server binary to bin/codevalddt-server.
build-server:
	go build -o bin/codevalddt-server ./cmd/server

## Build the dev binary to bin/codevalddt-dev.
build-dev:
	go build -o bin/codevalddt-dev ./cmd/dev

## Run the production server locally. Expects env vars to be set by the caller
## (or the shell) — does not source .env, to mirror container behaviour.
server: build-server
	./bin/codevalddt-server

## Run the dev binary with local-dev defaults. Sources .env if present so
## DT_ARANGO_PASSWORD etc. stay out of the source tree.
dev: build-dev
	@if [ -f .env ]; then \
		set -a && . ./.env && set +a; \
	fi; \
	./bin/codevalddt-dev

## Stop any running instance, rebuild, and run.
restart: kill build-server
	@echo "Running codevalddt-server..."
	@if [ -f .env ]; then \
		set -a && . ./.env && set +a; \
	fi; \
	./bin/codevalddt-server

## Stop any running dev instance, rebuild, and run.
dev-restart: kill dev

## Stop any running instances of the codevalddt binaries.
kill:
	@echo "Stopping any running instances..."
	-@pkill -x codevalddt-server 2>/dev/null || true
	-@pkill -x codevalddt-dev 2>/dev/null || true
	@sleep 1

# ── Proto Codegen ─────────────────────────────────────────────────────────────

## Regenerate Go stubs from proto/*.
## CodeValdDT currently has no proto files of its own — the EntityService is
## re-exported from CodeValdSharedLib. This target is kept for symmetry with
## sister services and is a no-op when the proto/ directory is empty.
## Requires: buf, protoc-gen-go, protoc-gen-go-grpc on PATH.
## Install: go install github.com/bufbuild/buf/cmd/buf@latest
##          go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
##          go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
proto:
	@if [ -d proto ] && [ -n "$$(find proto -name '*.proto' -print -quit 2>/dev/null)" ]; then \
		buf generate; \
	else \
		echo "codevalddt: no proto files — skipping buf generate"; \
	fi

# ── Tests ─────────────────────────────────────────────────────────────────────

## Run all unit tests with race detector (skips integration tests that need ArangoDB).
test:
	go test -v -race -count=1 ./...

## Run tests and produce an HTML coverage report (coverage.html).
cover:
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## Run ArangoDB integration tests.
## Loads .env if it exists, otherwise falls back to environment variables.
test-arango:
	@if [ -f .env ]; then \
		set -a && . ./.env && set +a; \
	fi; \
	go test -v -race -count=1 ./storage/arangodb/

## Run unit + ArangoDB integration tests.
test-all: test test-arango

# ── Quality ───────────────────────────────────────────────────────────────────

vet:
	go vet ./...

lint:
	golangci-lint run ./...

# ── Clean ─────────────────────────────────────────────────────────────────────

clean:
	go clean ./...
	rm -f coverage.out coverage.html
	rm -rf bin/
