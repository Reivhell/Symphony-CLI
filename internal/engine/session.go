package engine

import (
	"errors"
	"os"
	"sync"
)

// WriteSession tracks files written during a single generation for rollback on cancel.
type WriteSession struct {
	mu           sync.Mutex
	writtenFiles []string
}

// NewWriteSession creates an empty session.
func NewWriteSession() *WriteSession {
	return &WriteSession{}
}

// Record adds a successfully written file path (absolute).
func (s *WriteSession) Record(path string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writtenFiles = append(s.writtenFiles, path)
}

// Rollback removes recorded files in reverse order. Empty directories are not removed.
func (s *WriteSession) Rollback() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for i := len(s.writtenFiles) - 1; i >= 0; i-- {
		p := s.writtenFiles[i]
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	s.writtenFiles = nil
	return errors.Join(errs...)
}
