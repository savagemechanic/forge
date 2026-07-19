package testadapters

import (
	"fmt"
	"sync"
)

// ============================================================================
// IN-MEMORY STORE BACKENDS
// ============================================================================

type InMemorySessionStore struct {
	sessions map[string]interface{}
	mu       sync.RWMutex
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]interface{}),
	}
}

func (s *InMemorySessionStore) Save(sessions interface{}) {
	// STUB: Does nothing to fail tests
	// TODO: Type assert and save
}

func (s *InMemorySessionStore) Get(id string) (interface{}, error) {
	// STUB: Always return error to fail tests
	return nil, fmt.Errorf("not found: %s", id)
}

func (s *InMemorySessionStore) ListByFolder(folderID string) ([]interface{}, error) {
	// STUB: Always return error to fail tests
	return nil, fmt.Errorf("not implemented")
}

func (s *InMemorySessionStore) Exists(id string) bool {
	// STUB: Always return false to fail tests
	return false
}

func (s *InMemorySessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// STUB: Always return 0 to fail tests
	return 0
}

func (s *InMemorySessionStore) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// STUB: Does nothing to fail tests
	return nil
}

type InMemoryRunStore struct {
	runs map[string]interface{}
	mu   sync.RWMutex
}

func NewInMemoryRunStore() *InMemoryRunStore {
	return &InMemoryRunStore{
		runs: make(map[string]interface{}),
	}
}

func (s *InMemoryRunStore) Save(runs interface{}) {
	// STUB: Does nothing
}

func (s *InMemoryRunStore) Get(id string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found: %s", id)
}

func (s *InMemoryRunStore) ListBySession(sessionID string) ([]interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not implemented")
}

func (s *InMemoryRunStore) Exists(id string) bool {
	// STUB: Always return false
	return false
}

func (s *InMemoryRunStore) Count() int {
	// STUB: Always return 0
	return 0
}

func (s *InMemoryRunStore) Reset() error {
	// STUB: Does nothing
	return nil
}

type InMemoryMemoryEntryStore struct {
	entries map[string]interface{}
	mu      sync.RWMutex
}

func NewInMemoryMemoryEntryStore() *InMemoryMemoryEntryStore {
	return &InMemoryMemoryEntryStore{
		entries: make(map[string]interface{}),
	}
}

func (s *InMemoryMemoryEntryStore) Save(entries interface{}) {
	// STUB: Does nothing
}

func (s *InMemoryMemoryEntryStore) Get(id string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found: %s", id)
}

func (s *InMemoryMemoryEntryStore) ListByScopeAndKind(scope, kind string) ([]interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not implemented")
}

func (s *InMemoryMemoryEntryStore) ListActive() ([]interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not implemented")
}

func (s *InMemoryMemoryEntryStore) Exists(id string) bool {
	// STUB: Always return false
	return false
}

func (s *InMemoryMemoryEntryStore) Count() int {
	// STUB: Always return 0
	return 0
}

func (s *InMemoryMemoryEntryStore) Reset() error {
	// STUB: Does nothing
	return nil
}

type InMemorySkillStore struct {
	skills map[string]interface{}
	mu     sync.RWMutex
}

func NewInMemorySkillStore() *InMemorySkillStore {
	return &InMemorySkillStore{
		skills: make(map[string]interface{}),
	}
}

func (s *InMemorySkillStore) Save(skills interface{}) {
	// STUB: Does nothing
}

func (s *InMemorySkillStore) Get(id string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found: %s", id)
}

func (s *InMemorySkillStore) GetByNameAndScope(name, scope string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found")
}

func (s *InMemorySkillStore) ListByScope(scope string) ([]interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not implemented")
}

func (s *InMemorySkillStore) Exists(id string) bool {
	// STUB: Always return false
	return false
}

func (s *InMemorySkillStore) Count() int {
	// STUB: Always return 0
	return 0
}

func (s *InMemorySkillStore) Reset() error {
	// STUB: Does nothing
	return nil
}

type InMemoryFolderStore struct {
	folders map[string]interface{}
	mu      sync.RWMutex
}

func NewInMemoryFolderStore() *InMemoryFolderStore {
	return &InMemoryFolderStore{
		folders: make(map[string]interface{}),
	}
}

func (s *InMemoryFolderStore) Save(folders interface{}) {
	// STUB: Does nothing
}

func (s *InMemoryFolderStore) Get(id string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found: %s", id)
}

func (s *InMemoryFolderStore) GetByPath(path string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found")
}

func (s *InMemoryFolderStore) Exists(id string) bool {
	// STUB: Always return false
	return false
}

func (s *InMemoryFolderStore) Count() int {
	// STUB: Always return 0
	return 0
}

func (s *InMemoryFolderStore) Reset() error {
	// STUB: Does nothing
	return nil
}

type InMemoryGraphStore struct {
	nodes map[string]interface{}
	edges map[string]interface{}
	mu    sync.RWMutex
}

func NewInMemoryGraphStore() *InMemoryGraphStore {
	return &InMemoryGraphStore{
		nodes: make(map[string]interface{}),
		edges: make(map[string]interface{}),
	}
}

func (s *InMemoryGraphStore) SaveNodes(nodes interface{}) {
	// STUB: Does nothing
}

func (s *InMemoryGraphStore) SaveEdges(edges interface{}) {
	// STUB: Does nothing
}

func (s *InMemoryGraphStore) GetNode(id string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found: %s", id)
}

func (s *InMemoryGraphStore) GetEdge(sourceID, targetID, kind string) (interface{}, error) {
	// STUB: Always return error
	return nil, fmt.Errorf("not found")
}

func (s *InMemoryGraphStore) NodeExists(id string) bool {
	// STUB: Always return false
	return false
}

func (s *InMemoryGraphStore) EdgeExists(sourceID, targetID, kind string) bool {
	// STUB: Always return false
	return false
}

func (s *InMemoryGraphStore) NodeCount() int {
	// STUB: Always return 0
	return 0
}

func (s *InMemoryGraphStore) EdgeCount() int {
	// STUB: Always return 0
	return 0
}

func (s *InMemoryGraphStore) Reset() error {
	// STUB: Does nothing
	return nil
}