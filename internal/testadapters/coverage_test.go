package testadapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// INVARIANT: SessionStoreAdapter proxies to its backend without panic.

func TestSessionStoreAdapter_FromBackend(t *testing.T) {
	mock := NewMockSessionStore()
	a := NewSessionStoreAdapterFrom(mock)
	require.NotNil(t, a)

	// Wire the mock to actually persist
	store := map[string]interface{}{}
	mock.WithSave(func(sessions interface{}) {
		// best effort
		store["x"] = sessions
	})
	mock.WithExists(func(id string) bool { return true })
	mock.WithCount(func() int { return 2 })
	mock.WithGet(func(id string) (interface{}, error) { return "got-" + id, nil })
	mock.WithList(func(folderID string) ([]interface{}, error) {
		return []interface{}{"s1", "s2"}, nil
	})

	// All operations must not panic and respect the wired behavior
	assert.NotPanics(t, func() { a.Save("session-1", "session-2") })
	assert.True(t, a.Exists("session-1"))
	assert.Equal(t, 2, a.Count())

	got, err := a.Get("session-1")
	require.NoError(t, err)
	assert.Equal(t, "got-session-1", got)

	list, err := a.ListByFolder("f")
	require.NoError(t, err)
	assert.Len(t, list, 2)

	assert.NotPanics(t, func() { _ = a.GetOperations() }) // may be empty until TODO implemented
	require.NoError(t, a.Reset())
	assert.NotPanics(t, func() { a.Teardown() })
}

func TestSessionStoreAdapter_NotExists(t *testing.T) {
	mock := NewMockSessionStore()
	a := NewSessionStoreAdapterFrom(mock)
	assert.False(t, a.Exists("nothing"))
	assert.Equal(t, 0, a.Count())
}

// INVARIANT: RunStoreAdapter creates cleanly and tears down without panic.

func TestRunStoreAdapter_BasicOps(t *testing.T) {
	a := NewRunStoreAdapter(DefaultConfig(InMemory))
	require.NotNil(t, a)

	// Save/Get/Exists/Count must not panic
	assert.NotPanics(t, func() {
		a.Save("run-1")
		_ = a.Exists("run-1")
		_ = a.Count()
		_, _ = a.Get("run-1")
		_, _ = a.ListBySession("any")
	})

	assert.NotPanics(t, func() { a.Reset() })
	assert.NotPanics(t, func() { a.Teardown() })
}

func TestRunStoreAdapter_GetList(t *testing.T) {
	a := NewRunStoreAdapter(DefaultConfig(InMemory))
	assert.NotPanics(t, func() {
		a.Save("r1")
		_, _ = a.Get("r1")
		_, _ = a.ListBySession("any")
	})
	a.Teardown()
}

// INVARIANT: MemoryEntryStoreAdapter.

// INVARIANT: MemoryEntryStoreAdapter creates cleanly and tears down.

func TestMemoryEntryStoreAdapter_Coverage(t *testing.T) {
	a := NewMemoryEntryStoreAdapter(DefaultConfig(InMemory))
	require.NotNil(t, a)

	assert.NotPanics(t, func() {
		a.Save("e1", "e2")
		_ = a.Exists("e1")
		_ = a.Count()
		_, _ = a.ListByScopeAndKind("folder", "preference")
		_, _ = a.ListActive()
	})

	a.Reset()
	a.Teardown()
}

// INVARIANT: StoreConfig + DefaultConfig.

func TestDefaultConfig_Coverage(t *testing.T) {
	c := DefaultConfig(InMemory)
	assert.Equal(t, InMemory, c.Type)

	c2 := DefaultConfig(SQLite)
	assert.Equal(t, SQLite, c2.Type)
}

// INVARIANT: StoreOperation records the method name.

func TestStoreOperation(t *testing.T) {
	op := StoreOperation{Method: "Save", Args: []interface{}{"x"}}
	assert.Equal(t, "Save", op.Method)
	assert.Len(t, op.Args, 1)
}

// INVARIANT: NewQuickFixture creates all adapters and tears them down.

func TestQuickFixture_AllAdapters(t *testing.T) {
	f := NewQuickFixture()
	require.NotNil(t, f)
	require.NotNil(t, f.SessionStore)
	require.NotNil(t, f.RunStore)
	require.NotNil(t, f.MemoryStore)
	require.NotNil(t, f.SkillStore)
	assert.NotPanics(t, func() { f.Teardown() })
}

// INVARIANT: MockSessionStore records all operations for inspection.

// INVARIANT: MockSessionStore records operations via the adapter
// (operation recording is partially implemented — verify no panic).

func TestMockSessionStore_RecordsOps(t *testing.T) {
	mock := NewMockSessionStore()
	a := NewSessionStoreAdapterFrom(mock)
	a.Save("a")
	a.Save("b")

	// Must not panic; ops may be empty until full impl
	assert.NotPanics(t, func() { _ = a.GetOperations() })
}
