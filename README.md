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

## The Architecture

```
┌─────────────────────────────────────────────────┐
│                    TUI                          │
│ Header · Transcript · Composer · Overlays       │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│               Session Runtime                   │
│ Turns · Branches · Compaction · Runs · Events   │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│                  Agent Loop                     │
│ Intent · Context · Skills · Planning · Review   │
└───────────┬─────────────┬─────────────┬─────────┘
            │             │             │
┌───────────▼─────┐ ┌─────▼──────┐ ┌────▼────────┐
│ Go Intelligence│ │ Skill Engine│ │ Memory      │
│ AST/types/SSA   │ │ Load/create│ │ Folder/global│
└──────────────────┴─────────────┴───────────────┘
            │
┌───────────▼─────────────────────────────────────┐
│               Tool & Policy Layer               │
│ Files · Git · Commands · Validation · Approval  │
└─────────────────────────────────────────────────┘
```

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

### Phase 1: Chat Shell ✅ (In Progress)
- [x] Invariant tests (23 core invariants)
- [ ] Bubble Tea TUI
- [ ] Streaming responses
- [ ] Slash commands
- [ ] Interrupt support

### Phase 2: Folders & Sessions
- [ ] Folder discovery (Git root, go.work, go.mod)
- [ ] Session persistence
- [ ] Session resume/create/fork/compact

### Phase 3: Go Intelligence
- [ ] Package loading
- [ ] Symbol index
- [ ] References & call graph
- [ ] SSA analysis

### Phase 4: Safe Code Modification
- [ ] Worktree isolation
- [ ] Structured plans
- [ ] Approval workflows
- [ ] Rollback support

### Phase 5: Skills System
- [ ] Built-in skills
- [ ] Folder skills
- [ ] Global skills
- [ ] Skill creator from sessions

### Phase 6: Self-Modification
- [ ] Forge can modify Forge
- [ ] Controlled updates
- [ ] Rollback protection

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
git clone https://github.com/YOUR_USERNAME/forge.git
cd forge

# Run tests (red phase - all invariant tests fail intentionally)
go test ./...

# Build
go build -o forge cmd/forge/main.go

# Run (placeholder - in development)
./forge help
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

**[⭐ Star us on GitHub](https://github.com/YOUR_USERNAME/forge)**

---

*Made with ❤️ by the Forge community*