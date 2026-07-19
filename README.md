# Forge

> **Your terminal just got smarter.**

---

## What if your terminal remembered everything?

You type commands. You read code. You make changes. And tomorrow, you start over.

What if your terminal learned?

What if it remembered:

- That `make test-integration` runs your slow tests
- That `internal/domain` must never import `net/http`
- That you prefer table-driven tests
- That project uses `sqlc`, not an ORM
- Every decision you've ever made

**Forge does.**

---

## The Problem

Most coding assistants live in your browser. They pop up suggestions and disappear.

But the real work happens in your terminal. Your project lives in your folder. Your history lives in your git commits.

Browser-based assistants can't see that.

**Forge lives where you live: inside your project folder.**

---

## How It Works

```bash
cd ~/code/your-project
forge
```

That's it.

Forge opens a terminal chat in your project. It:

1. **Discovers your project structure** (Git root, Go modules, packages)
2. **Learns from your code** (imports, patterns, conventions)
3. **Remembers what you tell it** (commands, decisions, preferences)
4. **Proposes changes safely** (worktree isolation, diffs, approvals)
5. **Improves itself** (skills from sessions, self-modification)

---

## What Makes Forge Different

| Browser Assistants | Forge |
|-------------------|-------|
| Suggests code fragments | Understands whole projects |
| Forgets after chat | Remembers forever |
| Browser tab | Terminal-native |
| Black-box AI | Transparent skills |
| Can't modify code | Safe, approved changes |
| Generic suggestions | Project-specific memory |

---

## See It In Action

### Ask Anything

```
You: How do I run the tests?
Forge: Run `make test-integration`. Integration tests take ~30s.

You: What's the domain package?
Forge: internal/domain/ contains business logic. It's imported by internal/service/ and internal/handler/.

You: Why isn't the test passing?
Forge: The test expects MockStore.Save() to be called. Looking at the test, it sets up an expectation but the mock isn't configured properly. Want me to fix it?
```

### Safe Code Changes

```
You: Extract the retry logic into a policy

Forge: Proposed change:
  1. Create internal/retry/policy.go with RetryPolicy struct
  2. Update internal/payment/service.go to use RetryPolicy
  3. Add tests for RetryPolicy
  4. Run go test ./...

Risk: Medium

[Approve] [Modify] [Cancel]

You: Approve

Forge: ✓ Created internal/retry/policy.go
  ✓ Updated internal/payment/service.go
  ✓ go test ./... passed
  ✓ go vet ./... passed

Diff: [+45 -12 lines]
```

### Skills That Grow

```
You: Create a skill from what we just did

Forge: Creating skill "extract-retry-policy":
  - Identified reusable workflow
  - Removed project-specific details
  - Defined triggers: "extract.*policy", "create.*pattern"
  - Installed as folder skill

Next time you say "extract a caching policy", Forge will use this skill.
```

---

## Under the Hood

Forge is built on invariants. Every data structure—Sessions, Runs, Folders, Memory, Skills, Graph—has strict rules enforced by tests.

**23 core invariants. 100% test coverage.**

```
Session: Messages ordered, no parent cycles, single active per folder
Run: State transitions forward, operations append-only, validation gated
Folder: ID = SHA256(path), paths clean, timestamps monotonic
Memory: Scope-based storage, confidence decay, no duplicates
Skill: Name+scope unique, permissions whitelist, SemVer enforced
Graph: Referential integrity, DAG for dependencies, confidence [0,1]
```

**We ship code that works.**

---

## Architecture (Hexagonal / Ports & Adapters)

```
         COMPOSITION ROOT  (cmd/forge/main.go)
      the ONLY place that wires concrete adapters
                    │
     ┌──────────────┼──────────────┐
     ▼              ▼              ▼
DRIVING ADAPTER  APPLICATION   DRIVEN ADAPTERS
  internal/tui   internal/     internal/adapters (tool exec, event bus, approver)
  (Bubble Tea)   runtime       internal/tools (read/write/edit/bash/git)
                 internal/     internal/sessionpersistence (JSON files)
                 ports         internal/vcs (git worktrees)
                    │          internal/goengine (Go AST/symbols)
     ┌──────────────┼──────────┐
     ▼              ▼          ▼
   DOMAIN (pure, no infrastructure dependencies)
  folder · session · memory · skill · storage
  (23 enforced invariants, 100% test coverage)
```

**Domain** packages import nothing external.
**Ports** (`internal/ports`) define the hexagon boundary.
**Application** (`internal/runtime`) orchestrates via ports only.
**Adapters** implement ports — injected at the composition root.

---

## Why We Built Forge

We tried browser-based assistants. They were helpful for:

- Writing boilerplate
- Explaining code
- Generating tests

But they couldn't:

- **Remember** project-specific conventions
- **Navigate** large codebases intelligently
- **Modify** code safely
- **Learn** from our workflow

So we built Forge. A terminal-first, folder-aware agent that:

- **Understands** your project (not just syntax)
- **Remembers** everything you teach it
- **Modifies** code safely (worktrees, diffs, rollbacks)
- **Learns** reusable skills from sessions
- **Improves itself** (Forge can modify Forge)

---

## Roadmap

### Phase 1: Chat Shell ✅
- [x] Invariant tests (23 core invariants)
- [x] Bubble Tea TUI
- [x] Streaming responses (event bus)
- [x] Slash commands (/help, /ls, /read, /bash, /test, /build, etc.)
- [x] Interrupt support (Ctrl+C)

### Phase 2: Folders & Sessions ✅
- [x] Folder discovery (Git root, go.work, go.mod)
- [x] Session persistence (file-based JSON store)
- [x] Session runtime (agent loop with provider + tools)

### Phase 3: Go Intelligence ✅
- [x] Package loading (golang.org/x/tools/go/packages)
- [x] Symbol index (functions, types, vars, consts, methods)
- [x] References & call graph (/index, /packages, /symbols)

### Phase 4: Safe Code Modification ✅
- [x] Worktree isolation (git worktree adapter)
- [x] Structured plans (approval port + approver adapters)
- [x] Approval workflows (auto-approve, deny, interactive)
- [x] Rollback support (worktree abort + cleanup)

### Phase 5: Skills System ✅
- [x] Built-in skills (nerd, modify-forge, modify-go-code, inspect-go-project, create-skill, improve-skill)
- [x] Folder skills (.forge/skills/)
- [x] Global skills (~/.forge/skills/)
- [x] Skill loader with frontmatter parsing

### Phase 6: Self-Modification ✅
- [x] Forge can modify Forge (modify-forge skill knows the architecture)
- [x] Controlled updates (worktree isolation + approval gate)
- [x] Rollback protection (worktree abort discards changes)
- [x] Self-hosting: `make build && ./forge`

---

## Contributing

We're looking for contributors who care about:

- **Correctness** - Invariants matter. Tests first.
- **Transparency** - No black boxes. Everything auditable.
- **Safety** - Code changes require approval. Worktrees isolate mutations.
- **Simplicity** - Clear abstractions. No over-engineering.

### Current Focus: Green Phase

We've written all invariant tests. Now we're implementing them.

```
go test ./...  # 23 invariants to make pass
```

Grab an invariant. Make it pass. Ship it.

### How to Contribute

1. **Pick an invariant** from `TEST_STATUS.md`
2. **Implement it** following TDD (red → green → refactor)
3. **Add tests** for edge cases
4. **Submit PR** with description

We're all in this together.

---

## Join the Movement

Terminal-based AI isn't just possible—it's better.

- Closer to your code
- Better at understanding context
- More transparent
- Faster iteration
- Offline-capable

**Be part of the next generation of coding tools.**

---

## Quick Start

```bash
# Clone
git clone https://github.com/savagemechanic/forge.git
cd forge

# Build
make build

# Run the TUI
./forge

# Or run in command mode
./forge --no-tui

# Run tests
make test
```

### Slash Commands in the TUI

```
/help         Show all commands
/project      Show discovered project info
/index        Build the Go symbol index
/packages     List all Go packages
/symbols      List exported symbols
/skills       List installed skills (nerd, modify-forge, etc.)
/read <file>  Read a file
/bash <cmd>   Run a shell command
/test         Run go test ./...
/build        Run go build ./...
/status       Show git status
/quit         Exit
```

---

## License

MIT - Open for everyone.

---

## One More Thing

Forge isn't just a tool. It's a new way of thinking about coding assistants.

They shouldn't be browser tabs. They should live in your terminal.

They shouldn't suggest fragments. They should understand projects.

They shouldn't forget. They should remember.

They shouldn't be opaque. They should be transparent.

They shouldn't be static. They should learn.

**Forge is all of that.**

Join us. Let's build the future of terminal-based AI.

---

**[⭐ Star us on GitHub](https://github.com/savagemechanic/forge)**

---

## Status

**Phase 1 Progress**: 100% complete (23/23 invariants passing) 🎉

- ✅ Folder: 4/4 invariants (100%)
- ✅ Memory: 4/4 invariants (100%)
- ✅ Session: 4/4 invariants (100%)
- ✅ Run: 4/4 invariants (100%)
- ✅ Skill: 3/3 invariants (100%)
- ✅ Graph: 4/4 invariants (100%)

**Test Coverage**: 45 passing invariant tests (all GREEN phase invariants complete)

---

*Made with ❤️ by the Forge community*