# Contributing to Forge

Thank you for considering contributing to Forge! This document explains how to get started.

## Our Philosophy

**Test-Driven Development, Always.**

Every feature starts with a failing test. We don't ship code without tests. Ever.

## Getting Started

### Prerequisites

- Go 1.23 or later
- Git

### Setup

```bash
# Clone the repository
git clone https://github.com/savagemechanic/forge.git
cd forge

# Download dependencies
go mod download

# Run tests (expect failures - we're in red phase!)
go test ./...
```

## Current Status: Red Phase

We're in strict TDD red phase. We've written comprehensive invariant tests, and now we're implementing them one by one.

See `TEST_STATUS.md` for the current state of all 23 invariants.

## How to Contribute

### 1. Pick an Invariant

Look at `TEST_STATUS.md`. Each invariant shows:
- What it validates
- Current test status
- What needs implementation

### 2. Make Tests Pass

Follow the red-green-refactor cycle:

```bash
# Run tests - see which ones fail
go test ./internal/session/...

# Implement the invariant (make tests green)
# ... edit code ...

# Verify tests pass
go test ./internal/session/...

# Run all tests to ensure no regressions
go test ./...
```

### 3. Add Edge Cases

Don't just make the basic tests pass. Add more tests for edge cases:

```go
// Test the happy path
t.Run("should pass with valid input", func(t *testing.T) {
    // ...
})

// Test edge cases
t.Run("should fail with empty input", func(t *testing.T) {
    // ...
})

t.Run("should fail with malformed input", func(t *testing.T) {
    // ...
})
```

### 4. Submit a Pull Request

```bash
# Create a branch
git checkout -b implement/session-invariant-1

# Commit your changes
git add .
git commit -m "feat: implement session message ordering invariant

- Validate messages strictly ordered by timestamp
- Reject duplicate timestamps
- Add edge case tests for time boundaries
- Fixes 4/23 invariants"

# Push
git push origin implement/session-invariant-1

# Create PR via GitHub or gh CLI
gh pr create --title "Implement Session Message Ordering Invariant" --body "..."
```

## Code Style

- Follow Go conventions (`gofmt`, `go vet`)
- Keep functions focused and small
- Add comments for non-obvious logic
- Use meaningful variable names

## Test Style

- Use table-driven tests where appropriate
- Name tests descriptively: `TestType_WhatItDoes`
- Use `require` for assertions that must pass
- Use `assert` for assertions that provide information
- Add subtests for edge cases

## Questions?

- Open an issue for questions
- Join discussions in existing issues
- Check `TEST_STATUS.md` for current priorities

---

**Happy coding!** 🚀