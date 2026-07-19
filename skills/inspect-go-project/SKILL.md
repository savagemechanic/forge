---
name: inspect-go-project
description: Analyzes a Go project's structure — packages, symbols, call graph, dependencies. Builds a mental model using go/packages and AST traversal. Use when you need to understand how a Go codebase is organized.
version: 1.0.0
---

# Inspect-Go-Project — Go Codebase Analysis

You analyze Go projects to build an accurate mental model.

## Analysis Steps

1. **Start with `go.mod`** — module path, Go version, key dependencies.
2. **Map the packages** — run `/index` then `/packages` to see the full layout.
3. **Identify the entry point** — `main()` in `cmd/`.
4. **Trace the domain core** — which packages are pure logic vs infrastructure?
5. **Find the boundaries** — where are interfaces defined? Where implemented?
6. **Build the dependency graph** — what imports what? Are there cycles?

## Key Questions to Answer

- What is the module path and what does it produce?
- Where is the domain logic (pure, testable, no I/O)?
- Where are the adapters (HTTP, DB, CLI)?
- What are the public APIs (exported symbols)?
- What are the test patterns (table-driven? fixtures? mocks)?
- What build commands does it use?

## Reading Strategy

- **Top-down:** entry point → main flow → domain.
- **Bottom-up:** types/invariants → services → handlers → main.
- **Sideways:** pick one feature, trace it through all layers.

## Output Format

```
MODULE: github.com/example/project
ENTRY:   cmd/project/main.go
DOMAIN:  internal/domain (pure, no I/O)
ADAPTERS: internal/http, internal/db
TESTS:   table-driven, testify, 85% coverage

KEY SYMBOLS:
  - OrderService.PlaceOrder()    internal/domain/order.go
  - PaymentGateway.Charge()      internal/domain/payment.go (interface)
  - StripeGateway.Charge()       internal/adapters/stripe.go (impl)
```
