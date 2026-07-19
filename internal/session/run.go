package session

import (
	"encoding/json"
	"hash/fnv"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type RunState string

const (
	RunStatePlanning   RunState = "planning"
	RunStateApproved   RunState = "approved"
	RunStateExecuting  RunState = "executing"
	RunStateValidating RunState = "validating"
	RunStateDone       RunState = "done"
	RunStateFailed     RunState = "failed"
)

type Operation struct {
	ID        string
	Type      string // "read", "write", "patch", "command", etc.
	Path      string
	Content   string
	Timestamp time.Time
	Metadata  map[string]string
}

type Plan struct {
	ID          string
	Title       string
	Description string
	Steps       []PlanStep
	Risk        string
	Files       []string
	Tools       []string
	CreatedAt   time.Time
}

type PlanStep struct {
	ID          string
	Description string
	Tool        string
	Arguments   map[string]string
}

type ValidationResult struct {
	Passed     bool
	Outputs    []string
	Errors     []string
	Duration   time.Duration
	ValidatedAt time.Time
}

type Run struct {
	ID          string
	SessionID   string
	TriggerTurn string
	State       RunState
	Plan        Plan
	Operations  []Operation
	Validation  ValidationResult
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ============================================================================
// STUB IMPLEMENTATIONS (RED PHASE - ALL WILL FAIL)
// ============================================================================

// ValidateTransition checks that state transition is valid (forward only)
// STUB: Always returns false to fail tests
func (r *Run) ValidateTransition(newState RunState) (bool, error) {
	// TODO: Implement state transition validation
	// Valid transitions:
	// planning → approved, planning → failed
	// approved → executing, approved → failed
	// executing → validating, executing → failed
	// validating → done, validating → failed
	// done, failed → terminal (no further transitions)
	// For now, always fail
	return false, nil
}

// ValidateOperationsImmutable checks that operations haven't been modified
// STUB: Always returns false to fail tests
func (r *Run) ValidateOperationsImmutable() bool {
	// TODO: Implement operation immutability check
	// Need to track original operations and compare
	// For now, always fail
	return false
}

// OperationsSnapshot creates an immutable snapshot of current operations
func (r *Run) OperationsSnapshot() []Operation {
	snapshot := make([]Operation, len(r.Operations))
	copy(snapshot, r.Operations)
	return snapshot
}

// ValidateOperationsAgainstSnapshot checks operations match snapshot
// STUB: Always returns false to fail tests
func (r *Run) ValidateOperationsAgainstSnapshot(snapshot []Operation) bool {
	// TODO: Implement snapshot comparison
	// For now, always fail
	return false
}

// ValidateValidationState checks that Validation is only populated in validating/done/failed states
// STUB: Always returns false to fail tests
func (r *Run) ValidateValidationState() bool {
	// TODO: Implement validation state check
	// Validation should be empty in planning/approved/executing states
	// For now, always fail
	return false
}

// IsActive returns true if run is in an active state (executing or validating)
// STUB: Always returns false to fail tests
func (r *Run) IsActive() bool {
	// TODO: Return r.State == RunStateExecuting || r.State == RunStateValidating
	// For now, always fail
	return false
}

// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// ValidateSingleActivePerSession checks that at most one run per session is active
// STUB: Always returns false to fail tests
func ValidateSingleActivePerSession(sessionID string, store RunStore) (bool, error) {
	// TODO: Count active runs for session
	// For now, always fail
	return false, nil
}

// ============================================================================
// HELPERS
// ============================================================================

// operationHash creates a hash of an operation for immutability tracking
func operationHash(op Operation) string {
	data, _ := json.Marshal(op)
	h := fnv.New64a()
	h.Write(data)
	return string(h.Sum(nil))
}

// IsTerminalState returns true if state is terminal (done or failed)
func (r *Run) IsTerminalState() bool {
	return r.State == RunStateDone || r.State == RunStateFailed
}

// CanTransitionTo returns true if transition is valid
func (r *Run) CanTransitionTo(newState RunState) bool {
	// Terminal states can't transition
	if r.IsTerminalState() {
		return false
	}

	// Define valid transitions
	validTransitions := map[RunState][]RunState{
		RunStatePlanning:   {RunStateApproved, RunStateFailed},
		RunStateApproved:   {RunStateExecuting, RunStateFailed},
		RunStateExecuting:  {RunStateValidating, RunStateFailed},
		RunStateValidating: {RunStateDone, RunStateFailed},
	}

	allowed, ok := validTransitions[r.State]
	if !ok {
		return false
	}

	for _, state := range allowed {
		if state == newState {
			return true
		}
	}

	return false
}