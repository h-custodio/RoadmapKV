package storage

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

func TestNewStore(t *testing.T) {
	s := NewStore()

	if s == nil {
		t.Fatal("NewStore() returned nil")
	}

	s.Set("key", []byte("value"))

	if got := s.Get("key"); !bytes.Equal(got, []byte("value")) {
		t.Fatalf("Get(key) = %q, want %q", got, []byte("value"))
	}
}

func TestStoreMissingKey(t *testing.T) {
	s := NewStore()

	if got := s.Get("missing"); got != nil {
		t.Errorf("Get(missing) = %v, want nil", got)
	}

	if s.Exists("missing") {
		t.Error("Exists(missing) = true, want false")
	}
}

func TestStoreSetAndGet(t *testing.T) {
	s := NewStore()

	tests := []struct {
		name  string
		key   string
		value []byte
	}{
		{
			name:  "ordinary value",
			key:   "key",
			value: []byte("hello"),
		},
		{
			name:  "empty value",
			key:   "empty",
			value: []byte{},
		},
		{
			name:  "nil value",
			key:   "nil",
			value: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.Set(tt.key, tt.value)

			if !s.Exists(tt.key) {
				t.Fatal("Exists(key) = false after Set")
			}

			if got := s.Get(tt.key); !bytes.Equal(got, tt.value) {
				t.Errorf("Get(key) = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestStoreCopiesValues(t *testing.T) {
	s := NewStore()

	input := []byte("original")
	s.Set("key", input)

	// Changes to the caller's input must not affect stored data.
	input[0] = 'X'

	if got := s.Get("key"); !bytes.Equal(got, []byte("original")) {
		t.Fatalf("stored value changed after modifying input: got %q", got)
	}

	// Changes to a returned value must not affect stored data.
	result := s.Get("key")
	result[0] = 'Y'

	if got := s.Get("key"); !bytes.Equal(got, []byte("original")) {
		t.Fatalf("stored value changed after modifying result: got %q", got)
	}
}

func TestStoreOverwrite(t *testing.T) {
	s := NewStore()

	s.Set("key", []byte("first"))
	s.Set("key", []byte("second"))

	if got := s.Get("key"); !bytes.Equal(got, []byte("second")) {
		t.Errorf("Get(key) = %q, want %q", got, []byte("second"))
	}

	if !s.Exists("key") {
		t.Error("key should still exist after overwrite")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()

	s.Set("key", []byte("value"))

	// Deleting an existing key should return true
	if deleted := s.Delete("key"); !deleted {
		t.Error("Delete(key) = false, want true for existing key")
	}

	if s.Exists("key") {
		t.Error("key still exists after Delete")
	}

	if got := s.Get("key"); got != nil {
		t.Errorf("Get(deleted key) = %v, want nil", got)
	}

	// Deleting a missing key should return false and be harmless
	if deleted := s.Delete("missing"); deleted {
		t.Error("Delete(missing) = true, want false for non-existent key")
	}
}

func TestStoreIndependentKeys(t *testing.T) {
	s := NewStore()

	s.Set("first", []byte("one"))
	s.Set("second", []byte("two"))
	s.Delete("first")

	if s.Exists("first") {
		t.Error("first key still exists")
	}

	if got := s.Get("second"); !bytes.Equal(got, []byte("two")) {
		t.Errorf("second key changed after deleting first: got %q", got)
	}
}

func TestZeroValueStorePanics(t *testing.T) {
	tests := []struct {
		name string
		call func(*Store)
	}{
		{"Set", func(s *Store) { s.Set("key", []byte("value")) }},
		{"Get", func(s *Store) { s.Get("key") }},
		{"Delete", func(s *Store) { s.Delete("key") }},
		{"Exists", func(s *Store) { s.Exists("key") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("%s on a zero-value Store did not panic", tt.name)
				}
			}()

			tt.call(&Store{})
		})
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore()

	values := [][]byte{
		[]byte("alpha"),
		[]byte("bravo"),
		[]byte("charlie"),
	}

	s.Set("shared", values[0])

	var wg sync.WaitGroup

	// Concurrent writers replace the same key with complete values.
	for i := 0; i < 3; i++ {
		value := values[i]

		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 500; j++ {
				s.Set("shared", value)
			}
		}()
	}

	// Readers must only observe complete values written to the store.
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 500; j++ {
				got := s.Get("shared")

				valid := false
				for _, want := range values {
					if bytes.Equal(got, want) {
						valid = true
						break
					}
				}

				if !valid {
					t.Errorf("Get(shared) returned unexpected value %q", got)
					return
				}
			}
		}()
	}

	wg.Wait()

	if !s.Exists("shared") {
		t.Error("shared key should exist after concurrent access")
	}
}

func BenchmarkStoreSet(b *testing.B) {
	store := NewStore()
	val := []byte("value")

	// Pre-generate keys so string formatting isn't timed
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Set(keys[i], val)
	}
}

func BenchmarkStoreGet(b *testing.B) {
	store := NewStore()
	val := []byte("value")
	keys := make([]string, b.N)

	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
		store.Set(keys[i], val)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Get(keys[i])
	}
}

func BenchmarkStoreExists(b *testing.B) {
	store := NewStore()
	val := []byte("value")

	// Pre-populate keys so they exist to be looked up
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
		store.Set(keys[i], val)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Exists(keys[i])
	}
}

func BenchmarkStoreDelete(b *testing.B) {
	store := NewStore()
	val := []byte("value")
	keys := make([]string, b.N)

	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
		store.Set(keys[i], val)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Delete(keys[i])
	}
}
