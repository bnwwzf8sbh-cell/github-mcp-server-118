// Package notes provides an in-memory notes store and MCP tools for managing notes.
package notes

import (
	"fmt"
	"sync"
	"time"
)

// Note represents a single note entry.
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Store defines the interface for note persistence.
type Store interface {
	List() []*Note
	Get(id string) (*Note, error)
	Create(title, content string) *Note
	Delete(id string) error
}

// MemoryStore is a thread-safe in-memory implementation of Store.
type MemoryStore struct {
	mu     sync.RWMutex
	notes  map[string]*Note
	nextID int
}

// NewMemoryStore creates an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		notes:  make(map[string]*Note),
		nextID: 1,
	}
}

// List returns all notes in unspecified order.
func (s *MemoryStore) List() []*Note {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Note, 0, len(s.notes))
	for _, n := range s.notes {
		result = append(result, n)
	}
	return result
}

// Get returns the note with the given ID or an error if it does not exist.
func (s *MemoryStore) Get(id string) (*Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.notes[id]
	if !ok {
		return nil, fmt.Errorf("note %q not found", id)
	}
	return n, nil
}

// Create adds a new note and returns it.
func (s *MemoryStore) Create(title, content string) *Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	id := fmt.Sprintf("%d", s.nextID)
	s.nextID++

	n := &Note{
		ID:        id,
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.notes[id] = n
	return n
}

// Delete removes the note with the given ID or returns an error if it does not exist.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.notes[id]; !ok {
		return fmt.Errorf("note %q not found", id)
	}
	delete(s.notes, id)
	return nil
}
