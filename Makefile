.PHONY: build run test test-invariants vet tidy clean install-skills help

BINARY := forge
VERSION := 0.2.0

## build: Compile the forge binary
build:
	go build -o $(BINARY) ./cmd/forge

## run: Build and run Forge in the current directory
run: build
	./$(BINARY)

## test: Run all tests
test:
	go test ./...

## test-invariants: Run only the invariant tests
test-invariants:
	go test ./internal/folder ./internal/memory ./internal/session ./internal/skill ./internal/storage -v

## vet: Run go vet
vet:
	go vet ./...

## tidy: Run go mod tidy
tidy:
	go mod tidy

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)

## install-skills: Copy built-in skills to ~/.forge/skills
install-skills:
	mkdir -p ~/.forge/skills
	cp -r skills/* ~/.forge/skills/

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
