// Package sessionpersistence provides a file-based SessionRepository
// adapter that saves sessions as JSON files under .forge/sessions/.
package sessionpersistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cloudspacelab/forge/internal/ports"
)

// FileSessionStore implements ports.SessionRepository using JSON files.
type FileSessionStore struct {
	baseDir string
}

// NewFileSessionStore creates a store rooted at the given project dir.
func NewFileSessionStore(projectDir string) *FileSessionStore {
	return &FileSessionStore{
		baseDir: filepath.Join(projectDir, ".forge", "sessions"),
	}
}

// Save writes a session record to disk as JSON.
func (s *FileSessionStore) Save(rec *ports.SessionRecord) error {
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(s.baseDir, rec.ID+".json")
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Get reads a session record by ID.
func (s *FileSessionStore) Get(id string) (*ports.SessionRecord, error) {
	path := filepath.Join(s.baseDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("session %s: %w", id, err)
	}
	var rec ports.SessionRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// ListByFolder returns all sessions for a folder, newest first.
func (s *FileSessionStore) ListByFolder(folderID string) ([]*ports.SessionRecord, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var records []*ports.SessionRecord
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		rec, err := s.Get(strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil {
			continue
		}
		if folderID == "" || rec.FolderID == folderID {
			records = append(records, rec)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].UpdatedAt.After(records[j].UpdatedAt)
	})

	return records, nil
}

// RecordFromSession converts internal session state to a persistable record.
func RecordFromSession(id, folderID, state string, messages []sessionMessage) *ports.SessionRecord {
	now := time.Now()
	rec := &ports.SessionRecord{
		ID:        id,
		FolderID:  folderID,
		State:     state,
		CreatedAt: now,
		UpdatedAt: now,
	}
	for _, m := range messages {
		rec.Messages = append(rec.Messages, ports.SessionMessage{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}
	return rec
}

// sessionMessage is a minimal message shape for conversion helpers.
type sessionMessage struct {
	ID        string
	Role      string
	Content   string
	CreatedAt time.Time
}
