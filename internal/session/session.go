package session

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusIdle      SessionStatus = "idle"
	SessionStatusCompacted SessionStatus = "compacted"
)

type Message struct {
	ID        string
	Role      string // "user" or "assistant"
	Content   string
	CreatedAt time.Time
}

type Session struct {
	ID        string
	FolderID  string
	ParentID  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []Message
	Summary   string
	ActiveRun string
	Status    SessionStatus
}

// ============================================================================
// STUB IMPLEMENTATIONS (RED PHASE - ALL WILL FAIL)
// ============================================================================

// ValidateMessageOrder checks that messages are strictly ordered by timestamp
// STUB: Always returns false to fail tests
func (s *Session) ValidateMessageOrder() bool {
	// TODO: Implement proper validation
	// For now, always fail
	return false
}

// ValidateParentID checks that ParentID either empty or points to existing session, no cycles
// STUB: Always returns false to fail tests
func (s *Session) ValidateParentID(store SessionStore) (bool, error) {
	// TODO: Implement cycle detection and existence check
	// For now, always fail
	return false, nil
}

// ValidateActiveRun checks that ActiveRun points to a Run with matching SessionID
// STUB: Always returns false to fail tests
func (s *Session) ValidateActiveRun(store RunStore) (bool, error) {
	// TODO: Implement Run lookup and SessionID matching
	// For now, always fail
	return false, nil
}

// ============================================================================
// STORE INTERFACES
// ============================================================================

type SessionStore interface {
	Save(sessions ...*Session)
	Get(id string) (*Session, error)
	ListByFolder(folderID string) ([]*Session, error)
}

type RunStore interface {
	Save(runs ...*Run)
	Get(id string) (*Run, error)
}

// ============================================================================
// MEMORY STORES (FOR TESTING)
// ============================================================================

type MemorySessionStore struct {
	sessions map[string]*Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *MemorySessionStore) Save(sessions ...*Session) {
	for _, sess := range sessions {
		s.sessions[sess.ID] = sess
	}
}

func (s *MemorySessionStore) Get(id string) (*Session, error) {
	sess, ok := s.sessions[id]
	if !ok {
		return nil, errors.New("session not found")
	}
	return sess, nil
}

func (s *MemorySessionStore) ListByFolder(folderID string) ([]*Session, error) {
	// TODO: Implement filtering
	return nil, errors.New("not implemented")
}

type MemoryRunStore struct {
	runs map[string]*Run
}

func NewMemoryRunStore() *MemoryRunStore {
	return &MemoryRunStore{
		runs: make(map[string]*Run),
	}
}

func (r *MemoryRunStore) Save(runs ...*Run) {
	for _, run := range runs {
		r.runs[run.ID] = run
	}
}

func (r *MemoryRunStore) Get(id string) (*Run, error) {
	run, ok := r.runs[id]
	if !ok {
		return nil, errors.New("run not found")
	}
	return run, nil
}

// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// ValidateSingleActivePerFolder checks that at most one session per folder is active
// STUB: Always returns false to fail tests
func ValidateSingleActivePerFolder(folderID string, store SessionStore) (bool, error) {
	// TODO: Count active sessions for folder
	// For now, always fail
	return false, nil
}

// Helper to generate deterministic IDs
func GenerateID(prefix string) string {
	data := []byte(prefix + time.Now().String())
	hash := sha256.Sum256(data)
	return prefix + "-" + hex.EncodeToString(hash[:8])
}