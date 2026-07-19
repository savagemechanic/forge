---
name: Implement Invariant
about: Implement one of the 23 core invariants to move us toward green phase
title: '[INVARIANT] Implement <Component> <Invariant Name>'
labels: good first issue, tdd, invariant
assignees: ''

---

## Which Invariant?

Select from `TEST_STATUS.md`:

- [ ] Session: Messages strictly ordered by timestamp
- [ ] Session: ParentID no cycles
- [ ] Session: ActiveRun SessionID match
- [ ] Session: At most one active per folder
- [ ] Run: State transitions forward only
- [ ] Run: Operations append-only
- [ ] Run: Validation gated to validating/done/failed
- [ ] Run: At most one active per session
- [ ] Folder: ID = SHA256(CanonicalPath)
- [ ] Folder: CanonicalPath cleaned
- [ ] Folder: GitRoot consistency
- [ ] Folder: LastOpenedAt monotonic
- [ ] Memory: Scope determines storage location
- [ ] Memory: Status by source
- [ ] Memory: Confidence monotonic decreasing
- [ ] Memory: No duplicate active entries
- [ ] Skill: Name+Scope unique
- [ ] Skill: Permissions whitelist
- [ ] Skill: SemVer + no self-increment
- [ ] Skill: Entrypoint exists
- [ ] Graph: Edge referential integrity
- [ ] Graph: No duplicate edges
- [ ] Graph: DAG for CONTAINS/DEPENDS_ON
- [ ] Graph: Edge confidence in [0,1]

## Approach

Describe how you plan to implement this invariant:

1. 
2. 
3. 

## Tests

Which tests need to pass?

```bash
go test ./internal/<package>/... -run Test<Function>
```

## Additional Context

Any edge cases or considerations?