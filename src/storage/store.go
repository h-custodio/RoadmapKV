package storage

type Store struct {
	store map[string][]byte
}

// pointer is used to mutate the store instance
func (s *Store) Set(key string, value []byte) {
	storedValue := make([]byte, len(value))
	copy(storedValue, value)

	s.store[key] = storedValue
}

func (s *Store) Get(key string) []byte {
	value := s.store[key]

	// returns a copy of the value to avoid mutability
	result := make([]byte, len(value))
	copy(result, value)

	return result
}

func (s *Store) Delete(key string) {
	delete(s.store, key)
}
