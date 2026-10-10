package storage

import "sync"

type Store struct {
	rwmu  sync.RWMutex
	store map[string][]byte
}

// NewStore initializes the Store. Required because map zero-values
// are nil and will panic on write.
func NewStore() *Store {
	return &Store{
		store: make(map[string][]byte),
	}
}

// a helper to explicitly catch the bug if the constructor was bypassed.
func (s *Store) checkInitialized() {
	if s.store == nil {
		panic("storage: uninitialized Store. Use storage.NewStore() instead of a direct struct literal.")
	}
}

// pointer is used to mutate the store instance
func (s *Store) Set(key string, value []byte) {
	s.checkInitialized()

	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	storedValue := make([]byte, len(value))
	copy(storedValue, value)

	s.store[key] = storedValue
}

func (s *Store) Get(key string) []byte {
	s.checkInitialized()

	// allows multiple goroutines to read
	s.rwmu.RLock()
	// garantees to cleanup, even if something goes wrong
	defer s.rwmu.RUnlock()

	if s.store == nil {
		// returns nil if uninitialized instead of panicking
		return nil
	}

	value := s.store[key]

	// returns a copy of the value to avoid mutability
	result := make([]byte, len(value))
	copy(result, value)

	return result
}

func (s *Store) Delete(key string) {
	s.checkInitialized()

	s.rwmu.Lock()
	defer s.rwmu.Unlock()

	delete(s.store, key)
}
