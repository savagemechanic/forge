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
	MemoryKindFact        MemoryKind = "fact"
	MemoryKindInstruction MemoryKind = "instruction"
	MemoryKindPreference  MemoryKind = "preference"
	MemoryKindDecision    MemoryKind = "decision"
	MemoryKindWorkflow    MemoryKind = "workflow"
	MemoryKindSummary     MemoryKind = "summary"
	MemoryKindWarning     MemoryKind = "warning"
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
	ID           string
	Scope        MemoryScope
	Kind         MemoryKind
	Content      string
	Source       MemorySource
	Confidence   float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Evidence     []EvidenceRef
	Status       MemoryStatus
	SemanticHash string
	FolderID     string
	ApprovedAt   *time.Time
	ApprovedBy   string
}

// ============================================================================
// STUB IMPLEMENTATIONS - NOW IMPLEMENTED (GREEN PHASE)
// ============================================================================

// GetStorageLocation returns where this entry should be stored based on scope
func (e *MemoryEntry) GetStorageLocation() (string, error) {
	switch e.Scope {
	case MemoryScopeSession:
		return "in-memory", nil
	case MemoryScopeFolder:
		if e.FolderID == "" {
			return "", &InvalidScopeError{Scope: e.Scope}
		}
		return filepath.Join(".forge", "memory", e.FolderID+".db"), nil
	case MemoryScopeGlobal:
		return GetForgeGlobalPath(), nil
	default:
		return "", &InvalidScopeError{Scope: e.Scope}
	}
}

// SetInitialStatus sets status based on source
func (e *MemoryEntry) SetInitialStatus() {
	switch e.Source {
	case MemorySourceUserExplicit, MemorySourceRepoDerived, MemorySourceToolObserved, MemorySourceSkillDerived:
		e.Status = MemoryStatusActive
	case MemorySourceModelInferred:
		e.Status = MemoryStatusPending
	}
}

// ValidateStatusBySource checks status is appropriate for source
func (e *MemoryEntry) ValidateStatusBySource() bool {
	// model-inferred can only be active if approved
	if e.Source == MemorySourceModelInferred && e.Status == MemoryStatusActive {
		if e.ApprovedAt == nil || e.ApprovedBy == "" {
			return false
		}
	}
	return true
}

// ValidateConfidenceNotIncreased checks confidence hasn't increased
func (e *MemoryEntry) ValidateConfidenceNotIncreased(previous float64) bool {
	return e.Confidence <= previous
}

// ValidateConfidenceRange checks confidence is in [0, 1]
func (e *MemoryEntry) ValidateConfidenceRange() bool {
	return e.Confidence >= 0 && e.Confidence <= 1
}

// ApplyTimeDecay applies time-based confidence decay
func (e *MemoryEntry) ApplyTimeDecay(now time.Time) float64 {
	age := now.Sub(e.UpdatedAt)
	ageHours := age.Hours()

	// Decay based on source
	var decayFactor float64
	switch e.Source {
	case MemorySourceUserExplicit:
		decayFactor = 0.0001 // Very slow decay
	case MemorySourceRepoDerived, MemorySourceToolObserved, MemorySourceSkillDerived:
		decayFactor = 0.001 // Slow decay
	case MemorySourceModelInferred:
		decayFactor = 0.01  // Faster decay
	default:
		decayFactor = 0.001
	}

	// Exponential decay: new = old * e^(-rate * time)
	decayed := e.Confidence * exp(-decayFactor*ageHours)
	return decayed
}

// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// ValidateNoDuplicateActiveEntries checks no two active entries share Scope+Kind+SemanticHash
func ValidateNoDuplicateActiveEntries(store MemoryEntryStore) (bool, error) {
	entries, err := store.ListActive()
	if err != nil {
		return false, err
	}

	// Group by (Scope, Kind, SemanticHash)
	seen := make(map[string]bool)
	for _, entry := range entries {
		key := string(entry.Scope) + ":" + string(entry.Kind) + ":" + entry.SemanticHash
		if seen[key] {
			return false, nil
		}
		seen[key] = true
	}

	return true, nil
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
	var result []*MemoryEntry
	for _, entry := range s.entries {
		if entry.Scope == scope && entry.Kind == kind {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (s *InMemoryMemoryEntryStore) ListActive() ([]*MemoryEntry, error) {
	var result []*MemoryEntry
	for _, entry := range s.entries {
		if entry.Status == MemoryStatusActive {
			result = append(result, entry)
		}
	}
	return result, nil
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

// exp is a simple exponential function (we don't import math to keep it simple)
func exp(x float64) float64 {
	// Simple Taylor series approximation for e^x
	// e^x = 1 + x + x^2/2! + x^3/3! + ...
	result := 1.0
	term := 1.0
	for i := 1; i <= 20; i++ {
		term = term * x / float64(i)
		result += term
	}
	return result
}