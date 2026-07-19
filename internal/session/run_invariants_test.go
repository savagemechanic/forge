package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// RUN INVARIANT TESTS
// ============================================================================

// INVARIANT 1: State transitions only forward: planning → approved → executing → validating → (done|failed)
func TestRun_StateTransitionsForwardOnly(t *testing.T) {
	t.Run("should fail when transitioning backward from approved to planning", func(t *testing.T) {
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateApproved,
		}

		valid, err := run.ValidateTransition(RunStatePlanning)
		require.NoError(t, err)
		assert.False(t, valid, "Backward transition should fail")
	})

	t.Run("should fail when transitioning backward from executing to approved", func(t *testing.T) {
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateExecuting,
		}

		valid, err := run.ValidateTransition(RunStateApproved)
		require.NoError(t, err)
		assert.False(t, valid, "Backward transition should fail")
	})

	t.Run("should fail when transitioning backward from done to validating", func(t *testing.T) {
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateDone,
		}

		valid, err := run.ValidateTransition(RunStateValidating)
		require.NoError(t, err)
		assert.False(t, valid, "Backward transition should fail")
	})

	t.Run("should fail when transitioning backward from failed to validating", func(t *testing.T) {
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateFailed,
		}

		valid, err := run.ValidateTransition(RunStateValidating)
		require.NoError(t, err)
		assert.False(t, valid, "Backward transition should fail")
	})

	t.Run("should fail when invalid transition from done to failed", func(t *testing.T) {
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateDone,
		}

		valid, err := run.ValidateTransition(RunStateFailed)
		require.NoError(t, err)
		assert.False(t, valid, "Done → Failed transition should fail")
	})

	t.Run("should fail when invalid transition from failed to done", func(t *testing.T) {
		run := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateFailed,
		}

		valid, err := run.ValidateTransition(RunStateDone)
		require.NoError(t, err)
		assert.False(t, valid, "Failed → Done transition should fail")
	})

	t.Run("should pass valid forward transitions", func(t *testing.T) {
		validTransitions := []struct {
			from     RunState
			to       RunState
			expected bool
		}{
			{RunStatePlanning, RunStateApproved, true},
			{RunStatePlanning, RunStateFailed, true}, // Can fail early
			{RunStateApproved, RunStateExecuting, true},
			{RunStateApproved, RunStateFailed, true}, // Can be rejected
			{RunStateExecuting, RunStateValidating, true},
			{RunStateExecuting, RunStateFailed, true}, // Can fail during execution
			{RunStateValidating, RunStateDone, true},
			{RunStateValidating, RunStateFailed, true}, // Can fail validation
		}

		for _, tt := range validTransitions {
			run := &Run{
				ID:        "run-1",
				SessionID: "sess-1",
				State:     tt.from,
			}

			valid, err := run.ValidateTransition(tt.to)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, valid,
				"Transition %s → %s should be %v", tt.from, tt.to, tt.expected)
		}
	})
}

// INVARIANT 2: Operations is append-only (immutable once written)
func TestRun_OperationsAppendOnly(t *testing.T) {
	t.Run("should fail when operation is modified after being added", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateExecuting,
			Operations: []Operation{
				{ID: "op-1", Type: "read", Path: "file.go", Content: "original"},
			},
		}

		// Try to modify an existing operation
		run.Operations[0].Content = "modified"

		valid := run.ValidateOperationsImmutable()
		assert.False(t, valid, "Modified operation should fail validation")
	})

	t.Run("should fail when operation is removed from middle", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateExecuting,
			Operations: []Operation{
				{ID: "op-1", Type: "read", Path: "file1.go"},
				{ID: "op-2", Type: "write", Path: "file2.go"},
				{ID: "op-3", Type: "read", Path: "file3.go"},
			},
		}

		// Remove operation from middle
		originalLen := len(run.Operations)
		originalSnapshot := make([]Operation, len(run.Operations))
		copy(originalSnapshot, run.Operations)
		run.Operations = append(run.Operations[:1], run.Operations[2:]...)

		valid := run.ValidateOperationsAgainstSnapshot(originalSnapshot)
		assert.False(t, valid, "Removed operation should fail validation")
		assert.Equal(t, originalLen, len(originalSnapshot), "Snapshot should remain unchanged")
	})

	t.Run("should pass when operations are only appended", func(t *testing.T) {
		run := &Run{
			ID:         "run-1",
			State:      RunStateExecuting,
			Operations: []Operation{},
		}

		snapshot := run.OperationsSnapshot()

		// Append new operations
		run.Operations = append(run.Operations,
			Operation{ID: "op-1", Type: "read", Path: "file1.go"},
			Operation{ID: "op-2", Type: "write", Path: "file2.go"},
		)

		valid := run.ValidateOperationsAgainstSnapshot(snapshot)
		assert.True(t, valid, "Append-only operations should pass validation")
	})

	t.Run("should fail when operation order is changed", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateExecuting,
			Operations: []Operation{
				{ID: "op-1", Type: "read", Path: "file1.go"},
				{ID: "op-2", Type: "write", Path: "file2.go"},
			},
		}

		originalSnapshot := run.OperationsSnapshot()

		// Swap order
		run.Operations[0], run.Operations[1] = run.Operations[1], run.Operations[0]

		valid := run.ValidateOperationsAgainstSnapshot(originalSnapshot)
		assert.False(t, valid, "Reordered operations should fail validation")
	})
}

// INVARIANT 3: Validation is only populated after state = validating or later
func TestRun_ValidationPopulatedAfterValidating(t *testing.T) {
	t.Run("should fail when Validation populated in planning state", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStatePlanning,
			Validation: ValidationResult{
				Passed:  true,
				Outputs: []string{"test passed"},
			},
		}

		valid := run.ValidateValidationState()
		assert.False(t, valid, "Validation in planning state should fail")
	})

	t.Run("should fail when Validation populated in approved state", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateApproved,
			Validation: ValidationResult{
				Passed:  false,
				Outputs: []string{"no tests yet"},
			},
		}

		valid := run.ValidateValidationState()
		assert.False(t, valid, "Validation in approved state should fail")
	})

	t.Run("should fail when Validation populated in executing state", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateExecuting,
			Validation: ValidationResult{
				Passed:  true,
				Outputs: []string{"test passed"},
			},
		}

		valid := run.ValidateValidationState()
		assert.False(t, valid, "Validation in executing state should fail")
	})

	t.Run("should pass when Validation populated in validating state", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateValidating,
			Validation: ValidationResult{
				Passed:  true,
				Outputs: []string{"test passed"},
			},
		}

		valid := run.ValidateValidationState()
		assert.True(t, valid, "Validation in validating state should pass")
	})

	t.Run("should pass when Validation populated in done state", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateDone,
			Validation: ValidationResult{
				Passed:  true,
				Outputs: []string{"all tests passed"},
			},
		}

		valid := run.ValidateValidationState()
		assert.True(t, valid, "Validation in done state should pass")
	})

	t.Run("should pass when Validation populated in failed state", func(t *testing.T) {
		run := &Run{
			ID:     "run-1",
			State:  RunStateFailed,
			Validation: ValidationResult{
				Passed:  false,
				Outputs: []string{"test failed"},
			},
		}

		valid := run.ValidateValidationState()
		assert.True(t, valid, "Validation in failed state should pass")
	})

	t.Run("should pass when Validation is empty before validating state", func(t *testing.T) {
		states := []RunState{
			RunStatePlanning,
			RunStateApproved,
			RunStateExecuting,
		}

		for _, state := range states {
			run := &Run{
				ID:         "run-1",
				State:      state,
				Validation: ValidationResult{}, // Empty
			}

			valid := run.ValidateValidationState()
			assert.True(t, valid, "Empty Validation in %s state should pass", state)
		}
	})
}

// INVARIANT 4: One Session can have multiple Runs, but only one Run can be Active per Session
func TestRun_AtMostOneActivePerSession(t *testing.T) {
	t.Run("should fail when two runs are active for same session", func(t *testing.T) {
		run1 := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateExecuting,
		}
		run2 := &Run{
			ID:        "run-2",
			SessionID: "sess-1", // SAME SESSION
			State:     RunStateExecuting, // BOTH ACTIVE
		}

		store := NewMemoryRunStore()
		store.Save(run1, run2)

		valid, err := ValidateSingleActivePerSession("sess-1", store)
		require.NoError(t, err)
		assert.False(t, valid, "Session with multiple active runs should fail validation")
	})

	t.Run("should fail when more than two runs are active", func(t *testing.T) {
		runs := []*Run{
			{ID: "run-1", SessionID: "sess-1", State: RunStateExecuting},
			{ID: "run-2", SessionID: "sess-1", State: RunStateExecuting},
			{ID: "run-3", SessionID: "sess-1", State: RunStateExecuting},
		}

		store := NewMemoryRunStore()
		store.Save(runs...)

		valid, err := ValidateSingleActivePerSession("sess-1", store)
		require.NoError(t, err)
		assert.False(t, valid, "Session with >2 active runs should fail validation")
	})

	t.Run("should pass when only one run is active", func(t *testing.T) {
		run1 := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateExecuting,
		}
		run2 := &Run{
			ID:        "run-2",
			SessionID: "sess-1",
			State:     RunStateDone, // NOT ACTIVE
		}
		run3 := &Run{
			ID:        "run-3",
			SessionID: "sess-1",
			State:     RunStatePlanning, // NOT ACTIVE
		}

		store := NewMemoryRunStore()
		store.Save(run1, run2, run3)

		valid, err := ValidateSingleActivePerSession("sess-1", store)
		require.NoError(t, err)
		assert.True(t, valid, "Session with exactly one active run should pass validation")
	})

	t.Run("should pass when runs are in different sessions", func(t *testing.T) {
		run1 := &Run{
			ID:        "run-1",
			SessionID: "sess-1",
			State:     RunStateExecuting,
		}
		run2 := &Run{
			ID:        "run-2",
			SessionID: "sess-2", // DIFFERENT SESSION
			State:     RunStateExecuting,
		}

		store := NewMemoryRunStore()
		store.Save(run1, run2)

		valid, err := ValidateSingleActivePerSession("sess-1", store)
		require.NoError(t, err)
		assert.True(t, valid, "Each session can have its own active run")
	})

	t.Run("should pass when no runs are active", func(t *testing.T) {
		runs := []*Run{
			{ID: "run-1", SessionID: "sess-1", State: RunStateDone},
			{ID: "run-2", SessionID: "sess-1", State: RunStateFailed},
			{ID: "run-3", SessionID: "sess-1", State: RunStatePlanning},
		}

		store := NewMemoryRunStore()
		store.Save(runs...)

		valid, err := ValidateSingleActivePerSession("sess-1", store)
		require.NoError(t, err)
		assert.True(t, valid, "Session with no active runs should pass validation")
	})

	t.Run("should define active states correctly", func(t *testing.T) {
		activeStates := []RunState{
			RunStateExecuting,
			RunStateValidating,
		}

		inactiveStates := []RunState{
			RunStatePlanning,
			RunStateApproved,
			RunStateDone,
			RunStateFailed,
		}

		for _, state := range activeStates {
			run := &Run{State: state}
			assert.True(t, run.IsActive(), "State %s should be active", state)
		}

		for _, state := range inactiveStates {
			run := &Run{State: state}
			assert.False(t, run.IsActive(), "State %s should be inactive", state)
		}
	})
}