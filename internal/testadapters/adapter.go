// Package testadapters provides pluggable test adapters for Forge testing.
//
// This package enables easy swapping of implementations for testing:
// - In-memory stores (fast, no persistence)
// - SQLite stores (realistic persistence)
// - Mock stores (custom behavior)
// - Spy stores (record interactions)
//
// Example:
//
//	// Use in-memory store
//	sessionStore := testadapters.NewSessionStoreAdapter(testadapters.InMemory)
//
//	// Use SQLite store
//	sessionStore := testadapters.NewSessionStoreAdapter(testadapters.SQLite)
//
//	// Use custom mock
//	customStore := &MyCustomStore{}
//	sessionStore := testadapters.NewSessionStoreAdapterFrom(customStore)
package testadapters

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ============================================================================
// ADAPTER TYPES
// ============================================================================

// StoreType specifies the type of store implementation
type StoreType string

const (
	InMemory StoreType = "in-memory"
	SQLite   StoreType = "sqlite"
	Mock     StoreType = "mock"
	Spy      StoreType = "spy"
)

// StoreConfig configures a store adapter
type StoreConfig struct {
	Type     StoreType
	DataDir  string // For SQLite stores
	Reset    bool   // Reset state before each test
	Teardown bool   // Cleanup after tests
}

// DefaultConfig returns default configuration for a store type
func DefaultConfig(storeType StoreType) StoreConfig {
	return StoreConfig{
		Type:     storeType,
		DataDir:  filepath.Join(os.TempDir(), "forge-test"),
		Reset:    true,
		Teardown: true,
	}
}

// ============================================================================
// SESSION STORE ADAPTER
// ============================================================================

// SessionStoreAdapter wraps different session store implementations
type SessionStoreAdapter struct {
	store      SessionStoreBackend
	config     StoreConfig
	cleanup    func()
	mu         sync.RWMutex
	operations []StoreOperation // For spy stores
}

type SessionStoreBackend interface {
	Save(sessions interface{})
	Get(id string) (interface{}, error)
	ListByFolder(folderID string) ([]interface{}, error)
	Exists(id string) bool
	Count() int
	Reset() error
}

type StoreOperation struct {
	Method   string
	Args     []interface{}
	Result   interface{}
	Error    error
	Duration int64 // nanoseconds
}

// NewSessionStoreAdapter creates a new session store adapter
func NewSessionStoreAdapter(config StoreConfig) *SessionStoreAdapter {
	var backend SessionStoreBackend
	var cleanup func()

	switch config.Type {
	case InMemory:
		backend = NewInMemorySessionStore()
		cleanup = func() {} // No cleanup needed
	case SQLite:
		backend = NewSQLiteSessionStore(config.DataDir)
		cleanup = func() { os.RemoveAll(config.DataDir) }
	case Mock:
		backend = NewMockSessionStore()
		cleanup = func() {}
	case Spy:
		backend = NewSpySessionStore()
		cleanup = func() {}
	default:
		panic(fmt.Sprintf("unknown store type: %s", config.Type))
	}

	adapter := &SessionStoreAdapter{
		store:      backend,
		config:     config,
		cleanup:    cleanup,
		operations: []StoreOperation{},
	}

	if config.Reset {
		adapter.Reset()
	}

	return adapter
}

// NewSessionStoreAdapterFrom wraps an existing store backend
func NewSessionStoreAdapterFrom(backend SessionStoreBackend) *SessionStoreAdapter {
	return &SessionStoreAdapter{
		store:   backend,
		config:  StoreConfig{Type: Mock},
		cleanup: func() {},
	}
}

// Save saves sessions
func (a *SessionStoreAdapter) Save(sessions ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// TODO: Type assertion and conversion
	a.store.Save(sessions)
}

// Get retrieves a session by ID
func (a *SessionStoreAdapter) Get(id string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Get(id)
}

// ListByFolder lists sessions by folder ID
func (a *SessionStoreAdapter) ListByFolder(folderID string) ([]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.ListByFolder(folderID)
}

// Exists checks if a session exists
func (a *SessionStoreAdapter) Exists(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Exists(id)
}

// Count returns the number of sessions
func (a *SessionStoreAdapter) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Count()
}

// Reset clears all sessions
func (a *SessionStoreAdapter) Reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.store.Reset()
}

// Teardown cleans up resources
func (a *SessionStoreAdapter) Teardown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cleanup != nil {
		a.cleanup()
	}
}

// GetOperations returns recorded operations (for spy stores)
func (a *SessionStoreAdapter) GetOperations() []StoreOperation {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.operations
}

// ============================================================================
// RUN STORE ADAPTER
// ============================================================================

// RunStoreAdapter wraps different run store implementations
type RunStoreAdapter struct {
	store   RunStoreBackend
	config  StoreConfig
	cleanup func()
	mu      sync.RWMutex
}

type RunStoreBackend interface {
	Save(runs interface{})
	Get(id string) (interface{}, error)
	ListBySession(sessionID string) ([]interface{}, error)
	Exists(id string) bool
	Count() int
	Reset() error
}

// NewRunStoreAdapter creates a new run store adapter
func NewRunStoreAdapter(config StoreConfig) *RunStoreAdapter {
	var backend RunStoreBackend
	var cleanup func()

	switch config.Type {
	case InMemory:
		backend = NewInMemoryRunStore()
		cleanup = func() {}
	case SQLite:
		backend = NewSQLiteRunStore(config.DataDir)
		cleanup = func() { os.RemoveAll(config.DataDir) }
	case Mock:
		backend = NewMockRunStore()
		cleanup = func() {}
	default:
		panic(fmt.Sprintf("unknown store type: %s", config.Type))
	}

	return &RunStoreAdapter{
		store:   backend,
		config:  config,
		cleanup: cleanup,
	}
}

func (a *RunStoreAdapter) Save(runs ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.store.Save(runs)
}

func (a *RunStoreAdapter) Get(id string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Get(id)
}

func (a *RunStoreAdapter) ListBySession(sessionID string) ([]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.ListBySession(sessionID)
}

func (a *RunStoreAdapter) Exists(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Exists(id)
}

func (a *RunStoreAdapter) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Count()
}

func (a *RunStoreAdapter) Reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.store.Reset()
}

func (a *RunStoreAdapter) Teardown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cleanup != nil {
		a.cleanup()
	}
}

// ============================================================================
// MEMORY ENTRY STORE ADAPTER
// ============================================================================

// MemoryEntryStoreAdapter wraps different memory entry store implementations
type MemoryEntryStoreAdapter struct {
	store   MemoryEntryStoreBackend
	config  StoreConfig
	cleanup func()
	mu      sync.RWMutex
}

type MemoryEntryStoreBackend interface {
	Save(entries interface{})
	Get(id string) (interface{}, error)
	ListByScopeAndKind(scope, kind string) ([]interface{}, error)
	ListActive() ([]interface{}, error)
	Exists(id string) bool
	Count() int
	Reset() error
}

func NewMemoryEntryStoreAdapter(config StoreConfig) *MemoryEntryStoreAdapter {
	var backend MemoryEntryStoreBackend
	var cleanup func()

	switch config.Type {
	case InMemory:
		backend = NewInMemoryMemoryEntryStore()
		cleanup = func() {}
	case SQLite:
		backend = NewSQLiteMemoryEntryStore(config.DataDir)
		cleanup = func() { os.RemoveAll(config.DataDir) }
	default:
		panic(fmt.Sprintf("unknown store type: %s", config.Type))
	}

	return &MemoryEntryStoreAdapter{
		store:   backend,
		config:  config,
		cleanup: cleanup,
	}
}

func (a *MemoryEntryStoreAdapter) Save(entries ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.store.Save(entries)
}

func (a *MemoryEntryStoreAdapter) Get(id string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Get(id)
}

func (a *MemoryEntryStoreAdapter) ListByScopeAndKind(scope, kind string) ([]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.ListByScopeAndKind(scope, kind)
}

func (a *MemoryEntryStoreAdapter) ListActive() ([]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.ListActive()
}

func (a *MemoryEntryStoreAdapter) Exists(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Exists(id)
}

func (a *MemoryEntryStoreAdapter) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Count()
}

func (a *MemoryEntryStoreAdapter) Reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.store.Reset()
}

func (a *MemoryEntryStoreAdapter) Teardown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cleanup != nil {
		a.cleanup()
	}
}

// ============================================================================
// SKILL STORE ADAPTER
// ============================================================================

// SkillStoreAdapter wraps different skill store implementations
type SkillStoreAdapter struct {
	store   SkillStoreBackend
	config  StoreConfig
	cleanup func()
	mu      sync.RWMutex
}

type SkillStoreBackend interface {
	Save(skills interface{})
	Get(id string) (interface{}, error)
	GetByNameAndScope(name, scope string) (interface{}, error)
	ListByScope(scope string) ([]interface{}, error)
	Exists(id string) bool
	Count() int
	Reset() error
}

func NewSkillStoreAdapter(config StoreConfig) *SkillStoreAdapter {
	var backend SkillStoreBackend
	var cleanup func()

	switch config.Type {
	case InMemory:
		backend = NewInMemorySkillStore()
		cleanup = func() {}
	case SQLite:
		backend = NewSQLiteSkillStore(config.DataDir)
		cleanup = func() { os.RemoveAll(config.DataDir) }
	default:
		panic(fmt.Sprintf("unknown store type: %s", config.Type))
	}

	return &SkillStoreAdapter{
		store:   backend,
		config:  config,
		cleanup: cleanup,
	}
}

func (a *SkillStoreAdapter) Save(skills ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.store.Save(skills)
}

func (a *SkillStoreAdapter) Get(id string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Get(id)
}

func (a *SkillStoreAdapter) GetByNameAndScope(name, scope string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.GetByNameAndScope(name, scope)
}

func (a *SkillStoreAdapter) ListByScope(scope string) ([]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.ListByScope(scope)
}

func (a *SkillStoreAdapter) Exists(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Exists(id)
}

func (a *SkillStoreAdapter) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.Count()
}

func (a *SkillStoreAdapter) Reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.store.Reset()
}

func (a *SkillStoreAdapter) Teardown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cleanup != nil {
		a.cleanup()
	}
}

// ============================================================================
// TEST FIXTURE BUILDER
// ============================================================================

// TestFixture provides a complete set of stores for testing
type TestFixture struct {
	SessionStore     *SessionStoreAdapter
	RunStore         *RunStoreAdapter
	MemoryStore      *MemoryEntryStoreAdapter
	SkillStore       *SkillStoreAdapter
	FolderStore      *FolderStoreAdapter
	GraphStore       *GraphStoreAdapter
	Cleanup          func()
}

// FixtureConfig configures a test fixture
type FixtureConfig struct {
	SessionStore  StoreType
	RunStore      StoreType
	MemoryStore   StoreType
	SkillStore    StoreType
	FolderStore   StoreType
	GraphStore    StoreType
	DataDir       string
	GlobalTeardown bool
}

// DefaultFixtureConfig returns default fixture configuration
func DefaultFixtureConfig() FixtureConfig {
	return FixtureConfig{
		SessionStore:  InMemory,
		RunStore:      InMemory,
		MemoryStore:   InMemory,
		SkillStore:    InMemory,
		FolderStore:   InMemory,
		GraphStore:    InMemory,
		DataDir:       filepath.Join(os.TempDir(), "forge-test"),
		GlobalTeardown: true,
	}
}

// NewFixture creates a new test fixture with all stores
func NewFixture(config FixtureConfig) *TestFixture {
	if config.DataDir == "" {
		config.DataDir = filepath.Join(os.TempDir(), "forge-test")
	}

	dataDir := config.DataDir

	return &TestFixture{
		SessionStore: NewSessionStoreAdapter(StoreConfig{
			Type:    config.SessionStore,
			DataDir: filepath.Join(dataDir, "sessions"),
			Reset:   true,
		}),
		RunStore: NewRunStoreAdapter(StoreConfig{
			Type:    config.RunStore,
			DataDir: filepath.Join(dataDir, "runs"),
			Reset:   true,
		}),
		MemoryStore: NewMemoryEntryStoreAdapter(StoreConfig{
			Type:    config.MemoryStore,
			DataDir: filepath.Join(dataDir, "memory"),
			Reset:   true,
		}),
		SkillStore: NewSkillStoreAdapter(StoreConfig{
			Type:    config.SkillStore,
			DataDir: filepath.Join(dataDir, "skills"),
			Reset:   true,
		}),
		FolderStore: NewFolderStoreAdapter(StoreConfig{
			Type:    config.FolderStore,
			DataDir: filepath.Join(dataDir, "folders"),
			Reset:   true,
		}),
		GraphStore: NewGraphStoreAdapter(StoreConfig{
			Type:    config.GraphStore,
			DataDir: filepath.Join(dataDir, "graph"),
			Reset:   true,
		}),
		Cleanup: func() {
			if config.GlobalTeardown {
				os.RemoveAll(dataDir)
			}
		},
	}
}

// NewQuickFixture creates a quick in-memory fixture
func NewQuickFixture() *TestFixture {
	return NewFixture(DefaultFixtureConfig())
}

// Reset resets all stores in the fixture
func (f *TestFixture) Reset() error {
	if err := f.SessionStore.Reset(); err != nil {
		return err
	}
	if err := f.RunStore.Reset(); err != nil {
		return err
	}
	if err := f.MemoryStore.Reset(); err != nil {
		return err
	}
	if err := f.SkillStore.Reset(); err != nil {
		return err
	}
	if err := f.FolderStore.Reset(); err != nil {
		return err
	}
	if err := f.GraphStore.Reset(); err != nil {
		return err
	}
	return nil
}

// Teardown cleans up all stores
func (f *TestFixture) Teardown() {
	f.SessionStore.Teardown()
	f.RunStore.Teardown()
	f.MemoryStore.Teardown()
	f.SkillStore.Teardown()
	f.FolderStore.Teardown()
	f.GraphStore.Teardown()
	if f.Cleanup != nil {
		f.Cleanup()
	}
}