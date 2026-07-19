package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// TYPES
// ============================================================================

type MemoryScope string

const (
	MemoryScopeSession MemoryScope = "session"
	MemoryScopeFolder  MemoryScope = "folder"
	MemoryScopeGlobal  MemoryScope = "global"
)

type MemoryKind string

const (
	MemoryKindFact       MemoryKind = "fact"
	MemoryKindInstruction MemoryKind = "instruction"
	MemoryKindPreference MemoryKind = "preference"
	MemoryKindDecision   MemoryKind = "decision"
	MemoryKindWorkflow   MemoryKind = "workflow"
	MemoryKindSummary    MemoryKind = "summary"
	MemoryKindWarning    MemoryKind = "warning"
)

type MemorySource string

const (
	MemorySourceUserExplicit  MemorySource = "user-explicit"
	MemorySourceRepoDerived   MemorySource = "repo-derived"
	MemorySourceToolObserved  MemorySource = "tool-observed"
	MemorySourceSkillDerived  MemorySource = "skill-derived"
	MemorySourceModelInferred MemorySource = "model-inferred"
)

type MemoryStatus string

const (
	MemoryStatusActive   MemoryStatus = "active"
	MemoryStatusRejected MemoryStatus = "rejected"
	MemoryStatusPending  MemoryStatus = "pending"
)

type EvidenceRef struct {
	Type      string
	ID        string
	Timestamp time.Time
	Details   map[string]string
}

type MemoryEntry struct {
	ID            string
	Scope         MemoryScope
	Kind          MemoryKind
	Content       string
	Source        MemorySource
	Confidence    float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Evidence      []EvidenceRef
	Status        MemoryStatus
	SemanticHash  string
	FolderID      string // For folder-scoped entries
	ApprovedAt    *time.Time
	ApprovedBy    string
}

// ============================================================================
// STUB IMPLEMENTATIONS (RED PHASE - ALL WILL FAIL)
// ============================================================================

// GetStorageLocation returns where this entry should be stored based on scope
// STUB: Always returns error to fail tests
func (e *MemoryEntry) GetStorageLocation() (string, error) {
	// TODO: Implement scope-based storage location
	// - session: "in-memory"
	// - folder: ".forge/memory/<folder-id>.db"
	// - global: "~/.forge/global.db"
	// For now, return error
	return "", &InvalidScopeError{Scope: e.Scope}
}

// SetInitialStatus sets status based on source
// STUB: Does nothing to fail tests
func (e *MemoryEntry) SetInitialStatus() {
	// TODO: Implement initial status setting
	// - user-explicit, repo-derived, tool-observed, skill-derived → active
	// - model-inferred → pending
	// For now, do nothing
}

// ValidateStatusBySource checks status is appropriate for source
// STUB: Always returns false to fail tests
func (e *MemoryEntry) ValidateStatusBySource() bool {
	// TODO: Implement status validation
	// - model-inferred can only be active if approved
	// - For now, always fail
	return false
}

// ValidateConfidenceNotIncreased checks confidence hasn't increased
// STUB: Always returns false to fail tests
func (e *MemoryEntry) ValidateConfidenceNotIncreased(previous float64) bool {
	// TODO: Implement confidence validation
	// - New confidence must be <= previous confidence
	// - For now, always fail
	return false
}

// ValidateConfidenceRange checks confidence is in [0, 1]
// STUB: Always returns false to fail tests
func (e *MemoryEntry) ValidateConfidenceRange() bool {
	// TODO: Implement range validation
	// - Confidence must be >= 0 and <= 1
	// - For now, always fail
	return false
}

// ApplyTimeDecay applies time-based confidence decay
// STUB: Returns 0 to fail tests
func (e *MemoryEntry) ApplyTimeDecay(now time.Time) float64 {
	// TODO: Implement time decay function
	// - Decay based on age and source
	// - For now, return 0
	return 0
}

// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// ValidateNoDuplicateActiveEntries checks no two active entries share Scope+Kind+SemanticHash
// STUB: Always returns false to fail tests
func ValidateNoDuplicateActiveEntries(store MemoryEntryStore) (bool, error) {
	// TODO: Implement duplicate detection
	// - Group active entries by (Scope, Kind, SemanticHash)
	// - No group should have >1 entry
	// - For now, always fail
	return false, nil
}

// ============================================================================
// STORE INTERFACE
// ============================================================================

type MemoryEntryStore interface {
	Save(entries ...*MemoryEntry)
	Get(id string) (*MemoryEntry, error)
	ListByScopeAndKind(scope MemoryScope, kind MemoryKind) ([]*MemoryEntry, error)
	ListActive() ([]*MemoryEntry, error)
}

type InMemoryMemoryEntryStore struct {
	entries map[string]*MemoryEntry
}

func NewMemoryEntryStore() *InMemoryMemoryEntryStore {
	return &InMemoryMemoryEntryStore{
		entries: make(map[string]*MemoryEntry),
	}
}

func (s *InMemoryMemoryEntryStore) Save(entries ...*MemoryEntry) {
	for _, entry := range entries {
		s.entries[entry.ID] = entry
	}
}

func (s *InMemoryMemoryEntryStore) Get(id string) (*MemoryEntry, error) {
	entry, ok := s.entries[id]
	if !ok {
		return nil, &EntryNotFoundError{ID: id}
	}
	return entry, nil
}

func (s *InMemoryMemoryEntryStore) ListByScopeAndKind(scope MemoryScope, kind MemoryKind) ([]*MemoryEntry, error) {
	// STUB: Always return error
	return nil, &NotImplementedError{}
}

func (s *InMemoryMemoryEntryStore) ListActive() ([]*MemoryEntry, error) {
	// STUB: Always return error
	return nil, &NotImplementedError{}
}

// ============================================================================
// ERRORS
// ============================================================================

type InvalidScopeError struct {
	Scope MemoryScope
}

func (e *InvalidScopeError) Error() string {
	return "invalid memory scope: " + string(e.Scope)
}

type EntryNotFoundError struct {
	ID string
}

func (e *EntryNotFoundError) Error() string {
	return "memory entry not found: " + e.ID
}

type NotImplementedError struct{}

func (e *NotImplementedError) Error() string {
	return "not implemented"
}

// ============================================================================
// HELPERS
// ============================================================================

// computeSemanticHash creates a deterministic hash of content
func computeSemanticHash(content string) string {
	// Normalize content: lowercase, trim, remove extra whitespace
	normalized := strings.ToLower(strings.TrimSpace(content))
	normalized = strings.Join(strings.Fields(normalized), " ")

	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// GetForgeGlobalPath returns path to global forge data
func GetForgeGlobalPath() string {
	return filepath.Join(GetHomeDir(), ".forge", "global.db")
}

// timePtr returns a pointer to time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}