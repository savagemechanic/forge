package runtime

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudspacelab/forge/internal/adapters"
	"github.com/cloudspacelab/forge/internal/tools"
)

// RACE REGRESSION TEST
// INVARIANT: Submit() (goroutine B) and MessageCount/SessionSnapshot
// (goroutine A) must never race on session.Messages.
// Run with: go test -race ./internal/runtime/...
//
// Before the mutex fix, this test would trigger a data race detector
// report. It stays in the suite to prevent regression.

func TestRuntime_ConcurrentAccessNoRace(t *testing.T) {
	rt, err := New(Config{FolderPath: "../..", AutoApprove: true})
	if err != nil {
		t.Fatal(err)
	}
	rt.SetToolExecutor(adapters.NewToolExecutorAdapter(tools.NewRegistry()))
	rt.SetProvider(&EchoProvider{})
	rt.SetEventBus(adapters.NewChannelEventBus(64))

	var wg sync.WaitGroup

	// Goroutine A: read message count repeatedly
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = rt.MessageCount()
			_, _, _, _ = rt.SessionSnapshot()
		}
	}()

	// Goroutine B: submit messages (writes to session.Messages)
// Use EchoProvider: it returns Stop=true immediately with no tool
// execution, so the test only exercises the session-access race surface.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			rt.Submit(fmt.Sprintf("hello %d", i))
		}
	}()

	wg.Wait()

	// After concurrent work, count must be consistent
	count := rt.MessageCount()
	assert.Greater(t, count, 0, "messages should have been added")
}
