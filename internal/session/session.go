package session

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type SessionState string

const (
	SessionStateActive   SessionState = "active"
	SessionStateIdle     SessionState = "idle"
	SessionStateArchived SessionState = "archived"
)

type Message struct {
	ID        string
	Role      string // "user", "assistant", "system", "tool"
	Content   string
	Timestamp time.Time
	Metadata  map[string]string
}

type Session struct {
	ID             string
	FolderID       string
	ParentID       string
	State          SessionState
	Messages       []*Message
	ActiveRunID    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastActiveAt   time.Time
}

// ============================================================================
// STUB IMPLEMENTATIONS - NOW IMPLEMENTED (GREEN PHASE)
// ============================================================================

// ValidateIDIsDeterministicHash checks that ID is SHA256(folder_id + ":" + created_at)
func (s *Session) ValidateIDIsDeterministicHash() bool {
	expected := generateSessionID(s.FolderID, s.CreatedAt)
	return s.ID == expected
}

// ValidateMessageTimestamps checks that messages are strictly ordered by timestamp
func (s *Session) ValidateMessageTimestamps() bool {
	if len(s.Messages) == 0 {
		return true
	}

	for i := 1; i < len(s.Messages); i++ {
		if s.Messages[i].Timestamp.Equal(s.Messages[i-1].Timestamp) {
			return false // Duplicate timestamps rejected
		}
		if s.Messages[i].Timestamp.Before(s.Messages[i-1].Timestamp) {
			return false // Out of order
		}
	}
	return true
}

// ValidateParentIDNoCycles checks there are no cycles in parent chain
func (s *Session) ValidateParentIDNoCycles(store *InMemorySessionStore) (bool, error) {
	if s.ParentID == "" {
		return true, nil
	}

	// Check if parent exists
	parent, err := store.Get(s.ParentID)
	if err != nil {
		return false, &NonExistentParentError{ParentID: s.ParentID}
	}

	// Check for direct cycle (parent points to us)
	if parent.ParentID == s.ID {
		return false, &CycleDetectedError{SessionID: s.ID}
	}

	// Check for deep cycles (use visited set)
	visited := make(map[string]bool)
	current := parent
	for current.ParentID != "" {
		if visited[current.ParentID] {
			return false, &CycleDetectedError{SessionID: s.ID}
		}
		visited[current.ParentID] = true
		nextParent, err := store.Get(current.ParentID)
		if err != nil {
			return false, &NonExistentParentError{ParentID: current.ParentID}
		}
		if nextParent.ID == s.ID {
			return false, &CycleDetectedError{SessionID: s.ID}
		}
		current = nextParent
	}

	return true, nil
}

// ValidateActiveRunSessionIDMatch checks ActiveRunID references a run with matching SessionID
func (s *Session) ValidateActiveRunSessionIDMatch(runStore RunStore) bool {
	if s.ActiveRunID == "" {
		return true // Empty is OK
	}

	run, err := runStore.Get(s.ActiveRunID)
	if err != nil {
		return false // Non-existent run
	}

	return run.SessionID == s.ID
}

// UpdateLastActiveAt updates LastActiveAt to current time, enforcing monotonicity
func (s *Session) UpdateLastActiveAt() error {
	now := time.Now()
	if now.Before(s.LastActiveAt) {
		return &NonMonotonicTimestampError{
			Field:     "LastActiveAt",
			Previous:  s.LastActiveAt,
			Attempted: now,
		}
	}
	s.LastActiveAt = now
	s.UpdatedAt = now
	return nil
}

// ValidateAtMostOneActivePerFolder checks only one session is active per folder
func ValidateAtMostOneActivePerFolder(store *InMemorySessionStore) (bool, error) {
	folders := make(map[string]int)
	for _, session := range store.sessions {
		if session.State == SessionStateActive {
			folders[session.FolderID]++
			if folders[session.FolderID] > 1 {
				return false, nil
			}
		}
	}
	return true, nil
}

// ============================================================================
// STORE INTERFACE
// ============================================================================

type SessionStore interface {
	Save(sessions ...*Session)
	Get(id string) (*Session, error)
	ListByFolder(folderID string) ([]*Session, error)
	ListActive() ([]*Session, error)
}

type InMemorySessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *InMemorySessionStore) Save(sessions ...*Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, session := range sessions {
		s.sessions[session.ID] = session
	}
}

func (s *InMemorySessionStore) Get(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, &SessionNotFoundError{ID: id}
	}
	return session, nil
}

func (s *InMemorySessionStore) ListByFolder(folderID string) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Session
	for _, session := range s.sessions {
		if session.FolderID == folderID {
			result = append(result, session)
		}
	}
	return result, nil
}

func (s *InMemorySessionStore) ListActive() ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Session
	for _, session := range s.sessions {
		if session.State == SessionStateActive {
			result = append(result, session)
		}
	}
	return result, nil
}

// ============================================================================
// RUN TYPES AND IMPLEMENTATIONS
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

type Run struct {
	ID         string
	SessionID  string
	State      RunState
	Operations []Operation
	Validation Validation
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Operation struct {
	ID        string
	Type      string
	Path      string // For path-based operations
	Content   string // For content-based operations
	Target    string
	Arguments map[string]string
	CreatedAt time.Time
}

type Validation struct {
	RunID      string
	State      RunState
	CreatedAt  time.Time
	CompletedAt *time.Time
}

// ============================================================================
// RUN STUB IMPLEMENTATIONS - NOW IMPLEMENTED (GREEN PHASE)
// ============================================================================

var validTransitions = map[RunState][]RunState{
	RunStatePlanning:   {RunStateApproved, RunStateFailed},
	RunStateApproved:   {RunStateExecuting, RunStateFailed},
	RunStateExecuting:  {RunStateValidating, RunStateFailed},
	RunStateValidating: {RunStateDone, RunStateFailed},
	RunStateDone:       {}, // Terminal
	RunStateFailed:     {}, // Terminal
}

// canTransition checks if from → to is allowed
func canTransition(from, to RunState) bool {
	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}
	for _, allowedState := range allowed {
		if allowedState == to {
			return true
		}
	}
	return false
}

// ValidateTransition checks if state transition from current state to newState is valid
func (r *Run) ValidateTransition(newState RunState) (bool, error) {
	return canTransition(r.State, newState), nil
}

// ValidateStateTransitionsForwardOnly checks state transitions are forward-only (internal)
func (r *Run) ValidateStateTransitionsForwardOnly(previous RunState) bool {
	return canTransition(previous, r.State)
}

// ValidateOperationsAppendOnly checks operations are only appended, never modified or removed
func (r *Run) ValidateOperationsAppendOnly(original []Operation) bool {
	// Operations must not be shortened
	if len(r.Operations) < len(original) {
		return false
	}

	// First len(original) operations must be identical
	for i := 0; i < len(original); i++ {
		if r.Operations[i].ID != original[i].ID {
			return false
		}
		if r.Operations[i].Type != original[i].Type {
			return false
		}
		if r.Operations[i].Target != original[i].Target {
			return false
		}
		if !r.Operations[i].CreatedAt.Equal(original[i].CreatedAt) {
			return false
		}
	}

	// New operations must have later timestamps
	if len(r.Operations) > len(original) {
		lastOriginalTime := original[len(original)-1].CreatedAt
		for i := len(original); i < len(r.Operations); i++ {
			if r.Operations[i].CreatedAt.Before(lastOriginalTime) {
				return false
			}
		}
	}

	return true
}

// ValidateOperationsImmutable is an alias for ValidateOperationsAppendOnly
func (r *Run) ValidateOperationsImmutable(original []Operation) bool {
	return r.ValidateOperationsAppendOnly(original)
}

// ValidateOperationsAgainstSnapshot checks operations against a snapshot
func (r *Run) ValidateOperationsAgainstSnapshot(snapshot []Operation) bool {
	return r.ValidateOperationsAppendOnly(snapshot)
}

// OperationsSnapshot returns a snapshot of operations
func (r *Run) OperationsSnapshot() []Operation {
	snapshot := make([]Operation, len(r.Operations))
	copy(snapshot, r.Operations)
	return snapshot
}

// ValidateValidationPopulatedAfterValidating checks validation state gating
func (r *Run) ValidateValidationPopulatedAfterValidating() (bool, error) {
	if r.Validation.State == "" {
		return true, nil // Empty validation is OK
	}

	// Validation state must be one of: validating, done, failed
	validStates := map[RunState]bool{
		RunStateValidating: true,
		RunStateDone:       true,
		RunStateFailed:     true,
	}
	if !validStates[r.Validation.State] {
		return false, &InvalidValidationError{RunID: r.ID}
	}

	// Validation populated when run entered validating state
	if r.Validation.State == RunStateValidating {
		return true, nil
	}

	// Validation completed (done/failed) after validating
	if r.Validation.State == RunStateDone || r.Validation.State == RunStateFailed {
		if r.Validation.CompletedAt == nil {
			return false, &InvalidValidationError{RunID: r.ID, Reason: "missing completion time"}
		}
		// CompletedAt must be after CreatedAt
		if r.Validation.CompletedAt.Before(r.Validation.CreatedAt) {
			return false, &InvalidValidationError{RunID: r.ID, Reason: "completion time before creation time"}
		}
		return true, nil
	}

	return false, &InvalidValidationError{RunID: r.ID}
}

// ValidateAtMostOneActivePerRun checks at most one run is active per session
func ValidateAtMostOneActivePerSession(store *InMemoryRunStore) (bool, error) {
	sessions := make(map[string]bool)
	for _, run := range store.runs {
		if run.State == RunStateValidating {
			if sessions[run.SessionID] {
				return false, nil // Already have an active run for this session
			}
			sessions[run.SessionID] = true
		}
	}
	return true, nil
}

// ============================================================================
// RUN STORE
// ============================================================================

type RunStore interface {
	Save(runs ...*Run)
	Get(id string) (*Run, error)
	ListBySession(sessionID string) ([]*Run, error)
}

type InMemoryRunStore struct {
	runs map[string]*Run
	mu   sync.RWMutex
}

func NewRunStore() *InMemoryRunStore {
	return &InMemoryRunStore{
		runs: make(map[string]*Run),
	}
}

func (s *InMemoryRunStore) Save(runs ...*Run) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, run := range runs {
		s.runs[run.ID] = run
	}
}

func (s *InMemoryRunStore) Get(id string) (*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	run, ok := s.runs[id]
	if !ok {
		return nil, &RunNotFoundError{ID: id}
	}
	return run, nil
}

func (s *InMemoryRunStore) ListBySession(sessionID string) ([]*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Run
	for _, run := range s.runs {
		if run.SessionID == sessionID {
			result = append(result, run)
		}
	}
	return result, nil
}

// ============================================================================
// ERRORS
// ============================================================================

type SessionNotFoundError struct {
	ID string
}

func (e *SessionNotFoundError) Error() string {
	return "session not found: " + e.ID
}

type RunNotFoundError struct {
	ID string
}

func (e *RunNotFoundError) Error() string {
	return "run not found: " + e.ID
}

type NonMonotonicTimestampError struct {
	Field     string
	Previous  time.Time
	Attempted time.Time
}

func (e *NonMonotonicTimestampError) Error() string {
	return fmt.Sprintf("non-monotonic timestamp for %s: previous=%v, attempted=%v",
		e.Field, e.Previous, e.Attempted)
}

type NonExistentParentError struct {
	ParentID string
}

func (e *NonExistentParentError) Error() string {
	return fmt.Sprintf("non-existent parent session: %s", e.ParentID)
}

type CycleDetectedError struct {
	SessionID string
}

func (e *CycleDetectedError) Error() string {
	return fmt.Sprintf("cycle detected in session parent chain: %s", e.SessionID)
}

type InvalidValidationError struct {
	RunID  string
	Reason string
}

func (e *InvalidValidationError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("invalid validation state for run: %s", e.RunID)
	}
	return fmt.Sprintf("invalid validation state for run %s: %s", e.RunID, e.Reason)
}

// ============================================================================
// HELPERS
// ============================================================================

// generateSessionID creates a deterministic ID from folder ID and creation time
func generateSessionID(folderID string, createdAt time.Time) string {
	data := fmt.Sprintf("%s:%d", folderID, createdAt.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetCanonicalPath returns the cleaned, absolute path
func GetCanonicalPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absPath), nil
}

// ResolveSymlinks resolves symlinks in a path
func ResolveSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

// IsGitRepository checks if path is a git repository
func IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return true
	}
	return false
}