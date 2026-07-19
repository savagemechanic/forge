package testadapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ADAPTER TESTS (TESTING THE ADAPTER FRAMEWORK ITSELF)
// ============================================================================

func TestStoreTypes(t *testing.T) {
	t.Run("in-memory store type exists", func(t *testing.T) {
		assert.Equal(t, StoreType("in-memory"), InMemory)
	})

	t.Run("sqlite store type exists", func(t *testing.T) {
		assert.Equal(t, StoreType("sqlite"), SQLite)
	})

	t.Run("mock store type exists", func(t *testing.T) {
		assert.Equal(t, StoreType("mock"), Mock)
	})

	t.Run("spy store type exists", func(t *testing.T) {
		assert.Equal(t, StoreType("spy"), Spy)
	})
}

func TestDefaultConfig(t *testing.T) {
	t.Run("in-memory default config", func(t *testing.T) {
		config := DefaultConfig(InMemory)
		assert.Equal(t, InMemory, config.Type)
		assert.True(t, config.Reset)
		assert.True(t, config.Teardown)
	})

	t.Run("sqlite default config", func(t *testing.T) {
		config := DefaultConfig(SQLite)
		assert.Equal(t, SQLite, config.Type)
		assert.True(t, config.Reset)
		assert.True(t, config.Teardown)
		assert.Contains(t, config.DataDir, "forge-test")
	})
}

func TestSessionStoreAdapter(t *testing.T) {
	t.Run("create in-memory adapter", func(t *testing.T) {
		config := DefaultConfig(InMemory)
		adapter := NewSessionStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})

	t.Run("create sqlite adapter", func(t *testing.T) {
		config := DefaultConfig(SQLite)
		config.DataDir = t.TempDir()
		adapter := NewSessionStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})

	t.Run("create mock adapter", func(t *testing.T) {
		config := DefaultConfig(Mock)
		adapter := NewSessionStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})

	t.Run("get operations returns empty for non-spy", func(t *testing.T) {
		adapter := NewSessionStoreAdapter(DefaultConfig(InMemory))
		ops := adapter.GetOperations()

		assert.NotNil(t, ops)
		assert.Empty(t, ops)
		adapter.Teardown()
	})
}

func TestRunStoreAdapter(t *testing.T) {
	t.Run("create in-memory adapter", func(t *testing.T) {
		config := DefaultConfig(InMemory)
		adapter := NewRunStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})
}

func TestMemoryEntryStoreAdapter(t *testing.T) {
	t.Run("create in-memory adapter", func(t *testing.T) {
		config := DefaultConfig(InMemory)
		adapter := NewMemoryEntryStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})
}

func TestSkillStoreAdapter(t *testing.T) {
	t.Run("create in-memory adapter", func(t *testing.T) {
		config := DefaultConfig(InMemory)
		adapter := NewSkillStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})
}

func TestFolderStoreAdapter(t *testing.T) {
	t.Run("create in-memory adapter", func(t *testing.T) {
		config := DefaultConfig(InMemory)
		adapter := NewFolderStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})
}

func TestGraphStoreAdapter(t *testing.T) {
	t.Run("create in-memory adapter", func(t *testing.T) {
		config := DefaultConfig(InMemory)
		adapter := NewGraphStoreAdapter(config)

		assert.NotNil(t, adapter)
		assert.NotNil(t, adapter.store)
		adapter.Teardown()
	})
}

func TestTestFixture(t *testing.T) {
	t.Run("create quick fixture with in-memory stores", func(t *testing.T) {
		fixture := NewQuickFixture()
		defer fixture.Teardown()

		assert.NotNil(t, fixture.SessionStore)
		assert.NotNil(t, fixture.RunStore)
		assert.NotNil(t, fixture.MemoryStore)
		assert.NotNil(t, fixture.SkillStore)
		assert.NotNil(t, fixture.FolderStore)
		assert.NotNil(t, fixture.GraphStore)
		assert.NotNil(t, fixture.Cleanup)
	})

	t.Run("create fixture with custom config", func(t *testing.T) {
		config := FixtureConfig{
			SessionStore: InMemory,
			RunStore:     InMemory,
			MemoryStore:  InMemory,
			SkillStore:   InMemory,
			FolderStore:  InMemory,
			GraphStore:   InMemory,
			DataDir:      t.TempDir(),
			GlobalTeardown: false,
		}

		fixture := NewFixture(config)
		defer fixture.Teardown()

		assert.NotNil(t, fixture)
	})

	t.Run("reset all stores", func(t *testing.T) {
		fixture := NewQuickFixture()
		defer fixture.Teardown()

		err := fixture.Reset()
		// This will fail in stub phase, but that's expected
		assert.Error(t, err, "Stub implementations should fail")
	})
}

func TestMockSessionStore(t *testing.T) {
	t.Run("create mock store", func(t *testing.T) {
		mock := NewMockSessionStore()

		assert.NotNil(t, mock)
	})

	t.Run("with custom get function", func(t *testing.T) {
		mock := NewMockSessionStore()
		called := false

		mock.WithGet(func(id string) (interface{}, error) {
			called = true
			return "result", nil
		})

		result, err := mock.Get("test-id")
		require.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "result", result)
	})

	t.Run("with custom save function", func(t *testing.T) {
		mock := NewMockSessionStore()
		called := false

		mock.WithSave(func(sessions interface{}) {
			called = true
		})

		mock.Save("test-data")
		assert.True(t, called)
	})

	t.Run("with custom count function", func(t *testing.T) {
		mock := NewMockSessionStore()
		mock.WithCount(func() int {
			return 42
		})

		count := mock.Count()
		assert.Equal(t, 42, count)
	})

	t.Run("chainable configuration", func(t *testing.T) {
		mock := NewMockSessionStore().
			WithGet(func(id string) (interface{}, error) {
				return nil, nil
			}).
			WithSave(func(sessions interface{}) {}).
			WithCount(func() int {
				return 10
			})

		assert.NotNil(t, mock)
		assert.Equal(t, 10, mock.Count())
	})
}

func TestSpySessionStore(t *testing.T) {
	t.Run("create spy store", func(t *testing.T) {
		spy := NewSpySessionStore()

		assert.NotNil(t, spy)
		assert.NotNil(t, spy.backend)
	})

	t.Run("records get call", func(t *testing.T) {
		spy := NewSpySessionStore()

		spy.Get("test-id")
		calls := spy.GetCalls()

		assert.Len(t, calls, 1)
		assert.Equal(t, "Get", calls[0].Method)
		assert.Contains(t, calls[0].Args, "test-id")
	})

	t.Run("records save call", func(t *testing.T) {
		spy := NewSpySessionStore()

		spy.Save("test-data")
		calls := spy.GetCalls()

		assert.Len(t, calls, 1)
		assert.Equal(t, "Save", calls[0].Method)
	})

	t.Run("records count call", func(t *testing.T) {
		spy := NewSpySessionStore()

		spy.Count()
		calls := spy.GetCalls()

		assert.Len(t, calls, 1)
		assert.Equal(t, "Count", calls[0].Method)
	})

	t.Run("wasCalled checks method was called", func(t *testing.T) {
		spy := NewSpySessionStore()

		assert.False(t, spy.WasCalled("Save"))
		spy.Save("data")
		assert.True(t, spy.WasCalled("Save"))
	})

	t.Run("callCount returns number of calls", func(t *testing.T) {
		spy := NewSpySessionStore()

		assert.Equal(t, 0, spy.CallCount("Save"))
		spy.Save("data1")
		assert.Equal(t, 1, spy.CallCount("Save"))
		spy.Save("data2")
		assert.Equal(t, 2, spy.CallCount("Save"))
	})

	t.Run("resetCalls clears call history", func(t *testing.T) {
		spy := NewSpySessionStore()

		spy.Save("data")
		spy.Get("id")
		assert.Equal(t, 2, len(spy.GetCalls()))

		spy.ResetCalls()
		assert.Equal(t, 0, len(spy.GetCalls()))
	})
}

func TestSpyRunStore(t *testing.T) {
	t.Run("records calls", func(t *testing.T) {
		spy := NewSpyRunStore()

		spy.Save("run-data")
		spy.Get("run-id")

		calls := spy.GetCalls()
		assert.Len(t, calls, 2)
		assert.Equal(t, "Save", calls[0].Method)
		assert.Equal(t, "Get", calls[1].Method)
	})

	t.Run("wasCalled works", func(t *testing.T) {
		spy := NewSpyRunStore()

		assert.False(t, spy.WasCalled("Save"))
		spy.Save("data")
		assert.True(t, spy.WasCalled("Save"))
	})
}

func TestInMemoryBackends(t *testing.T) {
	t.Run("in-memory session store exists", func(t *testing.T) {
		store := NewInMemorySessionStore()
		assert.NotNil(t, store)
	})

	t.Run("in-memory run store exists", func(t *testing.T) {
		store := NewInMemoryRunStore()
		assert.NotNil(t, store)
	})

	t.Run("in-memory memory entry store exists", func(t *testing.T) {
		store := NewInMemoryMemoryEntryStore()
		assert.NotNil(t, store)
	})

	t.Run("in-memory skill store exists", func(t *testing.T) {
		store := NewInMemorySkillStore()
		assert.NotNil(t, store)
	})

	t.Run("in-memory folder store exists", func(t *testing.T) {
		store := NewInMemoryFolderStore()
		assert.NotNil(t, store)
	})

	t.Run("in-memory graph store exists", func(t *testing.T) {
		store := NewInMemoryGraphStore()
		assert.NotNil(t, store)
	})
}

func TestSQLiteBackends(t *testing.T) {
	t.Run("sqlite session store exists", func(t *testing.T) {
		store := NewSQLiteSessionStore(t.TempDir())
		assert.NotNil(t, store)
		assert.Equal(t, t.TempDir(), store.dataDir)
	})

	t.Run("sqlite run store exists", func(t *testing.T) {
		store := NewSQLiteRunStore(t.TempDir())
		assert.NotNil(t, store)
	})

	t.Run("sqlite memory entry store exists", func(t *testing.T) {
		store := NewSQLiteMemoryEntryStore(t.TempDir())
		assert.NotNil(t, store)
	})

	t.Run("sqlite skill store exists", func(t *testing.T) {
		store := NewSQLiteSkillStore(t.TempDir())
		assert.NotNil(t, store)
	})

	t.Run("sqlite folder store exists", func(t *testing.T) {
		store := NewSQLiteFolderStore(t.TempDir())
		assert.NotNil(t, store)
	})

	t.Run("sqlite graph store exists", func(t *testing.T) {
		store := NewSQLiteGraphStore(t.TempDir())
		assert.NotNil(t, store)
	})
}