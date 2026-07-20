package adapters_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudspacelab/forge/internal/adapters"
	"github.com/cloudspacelab/forge/internal/ports"
	"github.com/cloudspacelab/forge/internal/sessionpersistence"
	"github.com/cloudspacelab/forge/internal/tools"
)

// CONTRACT TEST SUITE: ports.ToolExecutor
// INVARIANT: every ToolExecutor must honor the same shape — List() returns
// specs, Execute() returns a result (never panics), unknown tools error.

func TestToolExecutorContract(t *testing.T) {
	impls := map[string]ports.ToolExecutor{
		"registry": adapters.NewToolExecutorAdapter(tools.NewRegistry()),
	}
	for name, te := range impls {
		t.Run(name, func(t *testing.T) {
			runToolExecutorContract(t, te)
		})
	}
}

func runToolExecutorContract(t *testing.T, te ports.ToolExecutor) {
	t.Helper()
	ctx := &ports.ToolContext{WorkDir: t.TempDir()}

	t.Run("List returns specs", func(t *testing.T) {
		specs := te.List()
		assert.NotEmpty(t, specs)
		for _, s := range specs {
			assert.NotEmpty(t, s.Name)
			assert.NotEmpty(t, s.Description)
		}
	})

	t.Run("Execute unknown tool errors", func(t *testing.T) {
		r := te.Execute(ctx, "nonexistent_tool", nil)
		assert.Error(t, r.Err)
	})

	t.Run("Execute never panics", func(t *testing.T) {
		assert.NotPanics(t, func() {
			te.Execute(ctx, "list", map[string]string{"dir": "."})
		})
	})

	t.Run("read tool works", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(dir+"/x.txt", []byte("data"), 0644)
		r := te.Execute(&ports.ToolContext{WorkDir: dir}, "read", map[string]string{"path": "x.txt"})
		assert.NoError(t, r.Err)
		assert.Equal(t, "data", r.Output)
	})
}

// CONTRACT TEST SUITE: ports.EventBus
// INVARIANT: Emit never blocks; Channel returns a readable stream.

func TestEventBusContract(t *testing.T) {
	impls := map[string]ports.EventBus{
		"channel": adapters.NewChannelEventBus(8),
	}
	for name, bus := range impls {
		t.Run(name, func(t *testing.T) {
			runEventBusContract(t, bus)
		})
	}
}

func runEventBusContract(t *testing.T, bus ports.EventBus) {
	t.Helper()

	t.Run("Emit never blocks", func(t *testing.T) {
		// Emit way more than buffer — must not deadlock
		for i := 0; i < 1000; i++ {
			bus.Emit(ports.Event{Type: ports.EventSystem, Text: "x"})
		}
	})

	t.Run("Emit then read", func(t *testing.T) {
		bus2 := adapters.NewChannelEventBus(4)
		bus2.Emit(ports.Event{Type: ports.EventAssistant, Text: "hi"})
		e := <-bus2.Channel()
		assert.Equal(t, ports.EventAssistant, e.Type)
		assert.Equal(t, "hi", e.Text)
	})
}

// CONTRACT TEST SUITE: ports.SessionRepository
// INVARIANT: Save → Get round-trips; Get missing errors; List filters.

func TestSessionRepositoryContract(t *testing.T) {
	impls := map[string]ports.SessionRepository{
		"file": sessionpersistence.NewFileSessionStore(t.TempDir()),
	}
	for name, repo := range impls {
		t.Run(name, func(t *testing.T) {
			runSessionRepoContract(t, repo)
		})
	}
}

func runSessionRepoContract(t *testing.T, repo ports.SessionRepository) {
	t.Helper()

	t.Run("Get missing errors", func(t *testing.T) {
		_, err := repo.Get("nope")
		assert.Error(t, err)
	})

	t.Run("Save then Get", func(t *testing.T) {
		rec := &ports.SessionRecord{ID: "c1", FolderID: "f", State: "active"}
		assert.NoError(t, repo.Save(rec))
		got, err := repo.Get("c1")
		assert.NoError(t, err)
		assert.Equal(t, "c1", got.ID)
	})

	t.Run("List by folder", func(t *testing.T) {
		repo.Save(&ports.SessionRecord{ID: "l1", FolderID: "target", State: "x"})
		repo.Save(&ports.SessionRecord{ID: "l2", FolderID: "other", State: "x"})
		list, err := repo.ListByFolder("target")
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "l1", list[0].ID)
	})
}
