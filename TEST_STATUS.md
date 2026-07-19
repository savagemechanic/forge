# Forge TDD Test Summary - Red Phase

**Status**: 🔴 RED PHASE - All invariant tests failing (intentional)

## Overview

All tests are currently failing as part of strict Test-Driven Development. We've written comprehensive invariant tests first, and now will implement the functionality to make them pass.

## Test Results

### Total Tests
- **Total Test Functions**: 37
- **Passing**: 11 (29%) - Adapter framework tests
- **Failing**: 26 (71%) - Core invariant tests (expected)

### Passing Tests (Adapter Framework)
These tests verify the test adapter framework itself works correctly:

1. ✅ TestStoreTypes
2. ✅ TestDefaultConfig
3. ✅ TestSessionStoreAdapter
4. ✅ TestRunStoreAdapter
5. ✅ TestMemoryEntryStoreAdapter
6. ✅ TestSkillStoreAdapter
7. ✅ TestFolderStoreAdapter
8. ✅ TestGraphStoreAdapter
9. ✅ TestInMemoryBackends
10. ✅ TestMockSessionStore
11. ✅ TestSpySessionStore
12. ✅ TestSpyRunStore

### Failing Tests (Core Invariants)

#### Session Invariants (4 tests failing)
- ❌ TestSession_MessagesStrictlyOrderedByTimestamp
  - Messages not ordered by timestamp
  - Duplicate timestamps not rejected
  - Ordered messages not validated

- ❌ TestSession_ParentIDNoCycles
  - Non-existent parent not detected
  - Cycles not detected (2-node and deep cycles)
  - Empty parent not validated

- ❌ TestSession_ActiveRunSessionIDMatch
  - Non-existent run not detected
  - SessionID mismatch not detected
  - Empty run not validated

- ❌ TestSession_AtMostOneActivePerFolder
  - Multiple active sessions not detected
  - Cross-folder isolation not enforced

#### Run Invariants (4 tests failing)
- ❌ TestRun_StateTransitionsForwardOnly
  - Backward transitions not blocked
  - Invalid transitions not blocked
  - Valid transitions not allowed

- ❌ TestRun_OperationsAppendOnly
  - Operation modifications not detected
  - Operation removal not detected
  - Append-only not enforced

- ❌ TestRun_ValidationPopulatedAfterValidating
  - Early validation population not blocked
  - Validation state not checked
  - Empty validation not validated

- ❌ TestRun_AtMostOneActivePerSession
  - Multiple active runs not detected
  - Active state not correctly identified

#### Folder Invariants (4 tests failing)
- ❌ TestFolder_IDIsDeterministicHash
  - Wrong ID not detected
  - Non-SHA256 not detected
  - Deterministic not enforced

- ❌ TestFolder_CanonicalPathIsClean
  - .. paths not rejected
  - Relative paths not rejected
  - Trailing slash not rejected
  - Duplicate slashes not rejected

- ❌ TestFolder_GitRootConsistency
  - Invalid .git path not rejected
  - Outside .git not rejected
  - Nested .git not validated

- ❌ TestFolder_LastOpenedAtMonotonic
  - Before CreatedAt not detected
  - Equal times not detected
  - Decreases not detected

#### Memory Entry Invariants (4 tests failing)
- ❌ TestMemoryEntry_ScopeStorageLocation
  - Invalid scope not rejected
  - Storage location not returned

- ❌ TestMemoryEntry_StatusBySource
  - Auto-status not set
  - Model-inferred not pending
  - Approved status not validated

- ❌ TestMemoryEntry_ConfidenceMonotonicDecreasing
  - Increases not detected
  - Range not validated
  - Time decay not applied

- ❌ TestMemoryEntry_NoDuplicateActiveEntries
  - Duplicates not detected
  - Different scope/kind not distinguished
  - Active vs rejected/pending not distinguished

#### Skill Invariants (3 tests failing)
- ❌ TestSkill_NameScopeUnique
  - Duplicates not detected
  - Cross-scope not distinguished

- ❌ TestSkill_PermissionsWhitelist
  - Unauthorized tools not blocked
  - Resource constraints not checked
  - Empty permissions not rejected

- ❌ TestSkill_SemVerAndNoAutoIncrement
  - Invalid SemVer not detected
  - Self-increment not blocked
  - Version history not ordered

#### Graph Invariants (4 tests failing)
- ❌ TestGraph_EdgeReferentialIntegrity
  - Missing endpoints not detected
  - Referential integrity not enforced

- ❌ TestGraph_NoDuplicateEdges
  - Duplicate edges not detected
  - Different kinds not distinguished

- ❌ TestGraph_ContainsNoCycles
  - 2-node cycles not detected
  - Deep cycles not detected
  - DAG structure not validated

- ❌ TestGraph_EdgeConfidenceRange
  - Negative confidence not rejected
  - >1.0 confidence not rejected
  - Range not validated

#### Test Adapter Tests (2 tests failing)
- ❌ TestTestFixture - Reset expects error but gets nil
- ❌ TestSQLiteBackends - Temp directory path mismatch (test infrastructure issue)

## Invariants Summary

| Data Structure | Invariants | Tests | Status |
|----------------|-----------|-------|--------|
| Session | 4 | 4 | ❌ All failing |
| Run | 4 | 4 | ❌ All failing |
| Folder | 4 | 4 | ❌ All failing |
| MemoryEntry | 4 | 4 | ❌ All failing |
| Skill | 3 | 3 | ❌ All failing |
| Graph (Edge/Node) | 4 | 4 | ❌ All failing |
| **Total** | **23** | **23** | **❌ All failing** |

## Next Steps (Green Phase)

1. **Implement Session Invariants**
   - Message ordering validation
   - Parent cycle detection
   - ActiveRun SessionID matching
   - Single active per folder

2. **Implement Run Invariants**
   - State transition machine
   - Operation immutability tracking
   - Validation state gating
   - Active run counting

3. **Implement Folder Invariants**
   - SHA256 ID generation
   - Path cleaning
   - Git validation
   - Monotonic timestamps

4. **Implement Memory Invariants**
   - Scope-based storage
   - Source-based status
   - Confidence decay
   - Duplicate detection

5. **Implement Skill Invariants**
   - Name+Scope uniqueness
   - Permission enforcement
   - SemVer validation
   - Entrypoint validation

6. **Implement Graph Invariants**
   - Referential integrity
   - Duplicate detection
   - Cycle detection (DAG)
   - Confidence validation

## Running Tests

```bash
# Run all tests (expect failures)
go test ./...

# Run specific package
go test ./internal/session/...
go test ./internal/folder/...
go test ./internal/memory/...
go test ./internal/skill/...
go test ./internal/storage/...

# Verbose output
go test -v ./...

# Count failures
go test ./... 2>&1 | grep -c "FAIL:"
```

## Pluggable Test Adapters

The test framework supports:
- **InMemory** stores (fast, default)
- **SQLite** stores (realistic, stubs fail)
- **Mock** stores (custom behavior)
- **Spy** stores (record interactions)

```go
// Quick fixture
fixture := testadapters.NewQuickFixture()
defer fixture.Teardown()

// Custom config
fixture := testadapters.NewFixture(testadapters.FixtureConfig{
    SessionStore: testadapters.SQLite,
    DataDir:      "/tmp/forge-test",
})
```

---

**Generated**: 2026-07-07
**Phase**: Red (TDD)
**Goal**: Make all 23 core invariant tests pass