package main

type Store struct {
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Set(key string, value string) {
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	value, ok := s.data[key]
	return value, ok
}

func (s *Store) Del(key string) bool {
	_, ok := s.data[key]
	if !ok {
		return false
	}

	delete(s.data, key)
	return true
}

func (s *Store) Exists(key string) bool {
	_, ok := s.data[key]
	return ok
}

func (s *Store) Keys() []string {
	keys := make([]string, 0, len(s.data))

	for key := range s.data {
		keys = append(keys, key)
	}

	return keys
}

func (s *Store) Count() int {
	return len(s.data)
}

func (s *Store) Clear() {
	for key := range s.data {
		delete(s.data, key)
	}
}
