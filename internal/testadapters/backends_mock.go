package testadapters

import (
	"fmt"
	"sync"
)

// ============================================================================
// MOCK STORE BACKENDS (for custom test behavior)
// ============================================================================

type MockSessionStore struct {
	sessions   map[string]interface{}
	getFunc    func(id string) (interface{}, error)
	saveFunc   func(sessions interface{})
	listFunc   func(folderID string) ([]interface{}, error)
	existsFunc func(id string) bool
	countFunc  func() int
	resetFunc  func() error
	mu         sync.RWMutex
}

func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: make(map[string]interface{}),
	}
}

func (m *MockSessionStore) WithGet(fn func(id string) (interface{}, error)) *MockSessionStore {
	m.getFunc = fn
	return m
}

func (m *MockSessionStore) WithSave(fn func(sessions interface{})) *MockSessionStore {
	m.saveFunc = fn
	return m
}

func (m *MockSessionStore) WithList(fn func(folderID string) ([]interface{}, error)) *MockSessionStore {
	m.listFunc = fn
	return m
}

func (m *MockSessionStore) WithExists(fn func(id string) bool) *MockSessionStore {
	m.existsFunc = fn
	return m
}

func (m *MockSessionStore) WithCount(fn func() int) *MockSessionStore {
	m.countFunc = fn
	return m
}

func (m *MockSessionStore) WithReset(fn func() error) *MockSessionStore {
	m.resetFunc = fn
	return m
}

func (m *MockSessionStore) Save(sessions interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveFunc != nil {
		m.saveFunc(sessions)
	}
}

func (m *MockSessionStore) Get(id string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getFunc != nil {
		return m.getFunc(id)
	}
	return nil, fmt.Errorf("not found: %s", id)
}

func (m *MockSessionStore) ListByFolder(folderID string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.listFunc != nil {
		return m.listFunc(folderID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockSessionStore) Exists(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.existsFunc != nil {
		return m.existsFunc(id)
	}
	return false
}

func (m *MockSessionStore) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.countFunc != nil {
		return m.countFunc()
	}
	return 0
}

func (m *MockSessionStore) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.resetFunc != nil {
		return m.resetFunc()
	}
	return nil
}

// ============================================================================
// SPY STORE BACKENDS (record interactions)
// ============================================================================

type SpySessionStore struct {
	sessions    map[string]interface{}
	calls       []CallRecord
	mu          sync.RWMutex
	backend     SessionStoreBackend
}

type CallRecord struct {
	Method string
	Args   []interface{}
	Result interface{}
	Error  error
}

func NewSpySessionStore() *SpySessionStore {
	return &SpySessionStore{
		sessions: make(map[string]interface{}),
		backend:  NewInMemorySessionStore(),
	}
}

func (s *SpySessionStore) record(method string, args []interface{}, result interface{}, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, CallRecord{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
	})
}

func (s *SpySessionStore) Save(sessions interface{}) {
	s.record("Save", []interface{}{sessions}, nil, nil)
	s.backend.Save(sessions)
}

func (s *SpySessionStore) Get(id string) (interface{}, error) {
	result, err := s.backend.Get(id)
	s.record("Get", []interface{}{id}, result, err)
	return result, err
}

func (s *SpySessionStore) ListByFolder(folderID string) ([]interface{}, error) {
	result, err := s.backend.ListByFolder(folderID)
	s.record("ListByFolder", []interface{}{folderID}, result, err)
	return result, err
}

func (s *SpySessionStore) Exists(id string) bool {
	s.record("Exists", []interface{}{id}, nil, nil)
	return s.backend.Exists(id)
}

func (s *SpySessionStore) Count() int {
	s.record("Count", nil, nil, nil)
	return s.backend.Count()
}

func (s *SpySessionStore) Reset() error {
	err := s.backend.Reset()
	s.record("Reset", nil, nil, err)
	return err
}

func (s *SpySessionStore) GetCalls() []CallRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	calls := make([]CallRecord, len(s.calls))
	copy(calls, s.calls)
	return calls
}

func (s *SpySessionStore) WasCalled(method string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, call := range s.calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

func (s *SpySessionStore) CallCount(method string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, call := range s.calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

func (s *SpySessionStore) ResetCalls() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = nil
}

// ============================================================================
// MOCK RUN STORE
// ============================================================================

type MockRunStore struct {
	runs       map[string]interface{}
	getFunc    func(id string) (interface{}, error)
	saveFunc   func(runs interface{})
	listFunc   func(sessionID string) ([]interface{}, error)
	existsFunc func(id string) bool
	countFunc  func() int
	resetFunc  func() error
	mu         sync.RWMutex
}

func NewMockRunStore() *MockRunStore {
	return &MockRunStore{
		runs: make(map[string]interface{}),
	}
}

func (m *MockRunStore) WithGet(fn func(id string) (interface{}, error)) *MockRunStore {
	m.getFunc = fn
	return m
}

func (m *MockRunStore) WithSave(fn func(runs interface{})) *MockRunStore {
	m.saveFunc = fn
	return m
}

func (m *MockRunStore) WithList(fn func(sessionID string) ([]interface{}, error)) *MockRunStore {
	m.listFunc = fn
	return m
}

func (m *MockRunStore) WithExists(fn func(id string) bool) *MockRunStore {
	m.existsFunc = fn
	return m
}

func (m *MockRunStore) WithCount(fn func() int) *MockRunStore {
	m.countFunc = fn
	return m
}

func (m *MockRunStore) WithReset(fn func() error) *MockRunStore {
	m.resetFunc = fn
	return m
}

func (m *MockRunStore) Save(runs interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveFunc != nil {
		m.saveFunc(runs)
	}
}

func (m *MockRunStore) Get(id string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.getFunc != nil {
		return m.getFunc(id)
	}
	return nil, fmt.Errorf("not found: %s", id)
}

func (m *MockRunStore) ListBySession(sessionID string) ([]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.listFunc != nil {
		return m.listFunc(sessionID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockRunStore) Exists(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.existsFunc != nil {
		return m.existsFunc(id)
	}
	return false
}

func (m *MockRunStore) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.countFunc != nil {
		return m.countFunc()
	}
	return 0
}

func (m *MockRunStore) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.resetFunc != nil {
		return m.resetFunc()
	}
	return nil
}

// ============================================================================
// SPY RUN STORE
// ============================================================================

type SpyRunStore struct {
	runs   map[string]interface{}
	calls  []CallRecord
	mu     sync.RWMutex
	backend RunStoreBackend
}

func NewSpyRunStore() *SpyRunStore {
	return &SpyRunStore{
		runs:    make(map[string]interface{}),
		backend: NewInMemoryRunStore(),
	}
}

func (s *SpyRunStore) record(method string, args []interface{}, result interface{}, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, CallRecord{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
	})
}

func (s *SpyRunStore) Save(runs interface{}) {
	s.record("Save", []interface{}{runs}, nil, nil)
	s.backend.Save(runs)
}

func (s *SpyRunStore) Get(id string) (interface{}, error) {
	result, err := s.backend.Get(id)
	s.record("Get", []interface{}{id}, result, err)
	return result, err
}

func (s *SpyRunStore) ListBySession(sessionID string) ([]interface{}, error) {
	result, err := s.backend.ListBySession(sessionID)
	s.record("ListBySession", []interface{}{sessionID}, result, err)
	return result, err
}

func (s *SpyRunStore) Exists(id string) bool {
	s.record("Exists", []interface{}{id}, nil, nil)
	return s.backend.Exists(id)
}

func (s *SpyRunStore) Count() int {
	s.record("Count", nil, nil, nil)
	return s.backend.Count()
}

func (s *SpyRunStore) Reset() error {
	err := s.backend.Reset()
	s.record("Reset", nil, nil, err)
	return err
}

func (s *SpyRunStore) GetCalls() []CallRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	calls := make([]CallRecord, len(s.calls))
	copy(calls, s.calls)
	return calls
}

func (s *SpyRunStore) WasCalled(method string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, call := range s.calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

func (s *SpyRunStore) ResetCalls() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = nil
}