---
name: modify-go-code
description: Safely modifies Go source code using find-and-replace edits, AST-aware changes, and worktree isolation. Enforces build and test verification after every change. Use when modifying any Go project.
version: 1.0.0
---

# Modify-Go-Code — Safe Code Modification

You modify Go code safely. Every change follows this protocol:

## Change Protocol

1. **Read before write.** Always read the target file first. Never edit blind.
2. **Use exact-match edits.** The `edit` tool requires unique `oldText`. If the text appears multiple times, include more context to make it unique.
3. **One logical change at a time.** Don't batch unrelated edits.
4. **Verify after every change:**
   ```bash
   go build ./...     # must compile
   go vet ./...       # must pass
   go test ./...      # must pass
   ```
5. **If any step fails, revert.** Don't leave the tree in a broken state.

## Safety Rules

- **Worktree isolation:** For multi-file changes, create a worktree first so the main tree stays clean until you approve.
- **No force pushes.** Never `git push --force` or amend shared history.
- **Test-first for new logic:** If adding behavior, write the test before the implementation.
- **Respect existing patterns:** Match the code style, naming, and structure already in the file.

## Edit Strategies

- **Rename:** Find all references first (`grep`), then rename atomically.
- **Extract:** Create the new file/function, update callers, verify build.
- **Inline:** Move the body to the call site, remove the original, verify.
- **Add field:** Update the struct, all constructors, all usages — in one pass.

## What to Never Do

- Edit a file you haven't read.
- Leave `go build` failing.
- Commit without running tests.
- Modify `go.mod` without running `go mod tidy`.
