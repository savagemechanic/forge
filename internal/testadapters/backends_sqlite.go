package testadapters

import (
	"fmt"
	"sync"
)

// ============================================================================
// FOLDER STORE BACKENDS
// ============================================================================

type FolderStoreBackend interface {
	Save(folders interface{})
	Get(id string) (interface{}, error)
	GetByPath(path string) (interface{}, error)
	Exists(id string) bool
	Count() int
	Reset() error
}

type FolderStoreAdapter struct {
	store   FolderStoreBackend
	config  StoreConfig
	cleanup func()
	mu      sync.RWMutex
}

func NewFolderStoreAdapter(config StoreConfig) *FolderStoreAdapter {
	var backend FolderStoreBackend
	var cleanup func()

	switch config.Type {
	case InMemory:
		backend = NewInMemoryFolderStore()
		cleanup = func() {}
	case SQLite:
		backend = NewSQLiteFolderStore(config.DataDir)
		cleanup = func() {} // TODO: cleanup
	default:
		panic(fmt.Sprintf("unknown store type: %s", config.Type))
	}

	return &FolderStoreAdapter{
		store:   backend,
		config:  config,
		cleanup: cleanup,
	}
}

func (a *FolderStoreAdapter) Save(folders ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.store.Save(folders)
}

func (a *FolderStoreAdapter) Get(id string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.Get(id)
}

func (a *FolderStoreAdapter) GetByPath(path string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.GetByPath(path)
}

func (a *FolderStoreAdapter) Exists(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.Exists(id)
}

func (a *FolderStoreAdapter) Count() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.Count()
}

func (a *FolderStoreAdapter) Reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.store.Reset()
}

func (a *FolderStoreAdapter) Teardown() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cleanup != nil {
		a.cleanup()
	}
}

// ============================================================================
// GRAPH STORE BACKENDS
// ============================================================================

type GraphStoreBackend interface {
	SaveNodes(nodes interface{})
	SaveEdges(edges interface{})
	GetNode(id string) (interface{}, error)
	GetEdge(sourceID, targetID, kind string) (interface{}, error)
	NodeExists(id string) bool
	EdgeExists(sourceID, targetID, kind string) bool
	NodeCount() int
	EdgeCount() int
	Reset() error
}

type GraphStoreAdapter struct {
	store   GraphStoreBackend
	config  StoreConfig
	cleanup func()
	mu      sync.RWMutex
}

func NewGraphStoreAdapter(config StoreConfig) *GraphStoreAdapter {
	var backend GraphStoreBackend
	var cleanup func()

	switch config.Type {
	case InMemory:
		backend = NewInMemoryGraphStore()
		cleanup = func() {}
	case SQLite:
		backend = NewSQLiteGraphStore(config.DataDir)
		cleanup = func() {}
	default:
		panic(fmt.Sprintf("unknown store type: %s", config.Type))
	}

	return &GraphStoreAdapter{
		store:   backend,
		config:  config,
		cleanup: cleanup,
	}
}

func (a *GraphStoreAdapter) SaveNodes(nodes ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.store.SaveNodes(nodes)
}

func (a *GraphStoreAdapter) SaveEdges(edges ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.store.SaveEdges(edges)
}

func (a *GraphStoreAdapter) GetNode(id string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.GetNode(id)
}

func (a *GraphStoreAdapter) GetEdge(sourceID, targetID, kind string) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.GetEdge(sourceID, targetID, kind)
}

func (a *GraphStoreAdapter) NodeExists(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.NodeExists(id)
}

func (a *GraphStoreAdapter) EdgeExists(sourceID, targetID, kind string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.EdgeExists(sourceID, targetID, kind)
}

func (a *GraphStoreAdapter) NodeCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.NodeCount()
}

func (a *GraphStoreAdapter) EdgeCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store.EdgeCount()
}

func (a *GraphStoreAdapter) Reset() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.store.Reset()
}

func (a *GraphStoreAdapter) Teardown() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cleanup != nil {
		a.cleanup()
	}
}

// ============================================================================
// SQLITE BACKENDS (STUBS - ALL FAIL TESTS)
// ============================================================================

type SQLiteSessionStore struct {
	dataDir string
	db      interface{} // Will be *sql.DB
	mu      sync.RWMutex
}

func NewSQLiteSessionStore(dataDir string) *SQLiteSessionStore {
	return &SQLiteSessionStore{
		dataDir: dataDir,
	}
}

func (s *SQLiteSessionStore) Save(sessions interface{}) {
	// STUB: Does nothing to fail tests
}

func (s *SQLiteSessionStore) Get(id string) (interface{}, error) {
	// STUB: Always return error to fail tests
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteSessionStore) ListByFolder(folderID string) ([]interface{}, error) {
	// STUB: Always return error to fail tests
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteSessionStore) Exists(id string) bool {
	// STUB: Always return false to fail tests
	return false
}

func (s *SQLiteSessionStore) Count() int {
	// STUB: Always return 0 to fail tests
	return 0
}

func (s *SQLiteSessionStore) Reset() error {
	// STUB: Always return error to fail tests
	return fmt.Errorf("sqlite not implemented")
}

type SQLiteRunStore struct {
	dataDir string
	db      interface{}
	mu      sync.RWMutex
}

func NewSQLiteRunStore(dataDir string) *SQLiteRunStore {
	return &SQLiteRunStore{
		dataDir: dataDir,
	}
}

func (s *SQLiteRunStore) Save(runs interface{}) {
	// STUB: Does nothing
}

func (s *SQLiteRunStore) Get(id string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteRunStore) ListBySession(sessionID string) ([]interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteRunStore) Exists(id string) bool {
	return false
}

func (s *SQLiteRunStore) Count() int {
	return 0
}

func (s *SQLiteRunStore) Reset() error {
	return fmt.Errorf("sqlite not implemented")
}

type SQLiteMemoryEntryStore struct {
	dataDir string
	db      interface{}
	mu      sync.RWMutex
}

func NewSQLiteMemoryEntryStore(dataDir string) *SQLiteMemoryEntryStore {
	return &SQLiteMemoryEntryStore{
		dataDir: dataDir,
	}
}

func (s *SQLiteMemoryEntryStore) Save(entries interface{}) {
	// STUB: Does nothing
}

func (s *SQLiteMemoryEntryStore) Get(id string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteMemoryEntryStore) ListByScopeAndKind(scope, kind string) ([]interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteMemoryEntryStore) ListActive() ([]interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteMemoryEntryStore) Exists(id string) bool {
	return false
}

func (s *SQLiteMemoryEntryStore) Count() int {
	return 0
}

func (s *SQLiteMemoryEntryStore) Reset() error {
	return fmt.Errorf("sqlite not implemented")
}

type SQLiteSkillStore struct {
	dataDir string
	db      interface{}
	mu      sync.RWMutex
}

func NewSQLiteSkillStore(dataDir string) *SQLiteSkillStore {
	return &SQLiteSkillStore{
		dataDir: dataDir,
	}
}

func (s *SQLiteSkillStore) Save(skills interface{}) {
	// STUB: Does nothing
}

func (s *SQLiteSkillStore) Get(id string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteSkillStore) GetByNameAndScope(name, scope string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteSkillStore) ListByScope(scope string) ([]interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteSkillStore) Exists(id string) bool {
	return false
}

func (s *SQLiteSkillStore) Count() int {
	return 0
}

func (s *SQLiteSkillStore) Reset() error {
	return fmt.Errorf("sqlite not implemented")
}

type SQLiteFolderStore struct {
	dataDir string
	db      interface{}
	mu      sync.RWMutex
}

func NewSQLiteFolderStore(dataDir string) *SQLiteFolderStore {
	return &SQLiteFolderStore{
		dataDir: dataDir,
	}
}

func (s *SQLiteFolderStore) Save(folders interface{}) {
	// STUB: Does nothing
}

func (s *SQLiteFolderStore) Get(id string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteFolderStore) GetByPath(path string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteFolderStore) Exists(id string) bool {
	return false
}

func (s *SQLiteFolderStore) Count() int {
	return 0
}

func (s *SQLiteFolderStore) Reset() error {
	return fmt.Errorf("sqlite not implemented")
}

type SQLiteGraphStore struct {
	dataDir string
	db      interface{}
	mu      sync.RWMutex
}

func NewSQLiteGraphStore(dataDir string) *SQLiteGraphStore {
	return &SQLiteGraphStore{
		dataDir: dataDir,
	}
}

func (s *SQLiteGraphStore) SaveNodes(nodes interface{}) {
	// STUB: Does nothing
}

func (s *SQLiteGraphStore) SaveEdges(edges interface{}) {
	// STUB: Does nothing
}

func (s *SQLiteGraphStore) GetNode(id string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteGraphStore) GetEdge(sourceID, targetID, kind string) (interface{}, error) {
	return nil, fmt.Errorf("sqlite not implemented")
}

func (s *SQLiteGraphStore) NodeExists(id string) bool {
	return false
}

func (s *SQLiteGraphStore) EdgeExists(sourceID, targetID, kind string) bool {
	return false
}

func (s *SQLiteGraphStore) NodeCount() int {
	return 0
}

func (s *SQLiteGraphStore) EdgeCount() int {
	return 0
}

func (s *SQLiteGraphStore) Reset() error {
	return fmt.Errorf("sqlite not implemented")
}