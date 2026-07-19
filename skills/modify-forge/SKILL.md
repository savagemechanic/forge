---
name: modify-forge
description: Modifies the Forge codebase itself. Knows the hexagonal architecture layout, the port/adapter boundaries, and the invariant-driven TDD discipline. Use when Forge needs to modify its own source code.
version: 1.0.0
---

# Modify-Forge — Self-Modification Skill

You are modifying Forge, a terminal coding agent built on **Hexagonal Architecture**.

## Architecture Map (do not violate these boundaries)

```
                    COMPOSITION ROOT
                  cmd/forge/main.go
            (the ONLY place that knows concrete adapters)
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
   DRIVING ADAPTER   APPLICATION    DRIVEN ADAPTERS
    internal/tui     internal/      internal/adapters
    internal/cli     runtime        internal/tools
                     internal/      internal/providers
                     ports          internal/storage
                         │          internal/sessionpersistence
         ┌───────────────┼───────────┐
         ▼               ▼           ▼
       DOMAIN (pure, no infrastructure deps)
    internal/folder   internal/session   internal/memory
    internal/skill    internal/storage   internal/goengine
```

## Rules

1. **Domain packages import nothing external** — no os, no net, no adapters.
2. **Ports live in `internal/ports`** — interfaces the application depends on.
3. **Adapters implement ports** — never the reverse dependency.
4. **The runtime (application) never imports concrete adapters** — only ports.
5. **Every new feature needs an invariant test first** (Red → Green → Refactor).
6. **All changes go through a worktree** — never modify the main tree directly.

## When Modifying Forge

1. Read the target package to understand current invariants.
2. Write the test first (invariant that must hold).
3. Implement in the correct layer (domain? port? adapter? application?).
4. Verify `go build ./...` and `go test ./...` pass.
5. Run `go vet ./...` before committing.
