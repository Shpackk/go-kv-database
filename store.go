package main

import (
	"sync"
	"time"
)

type Store struct {
	mu      sync.RWMutex
	data    map[string]string
	expires map[string]time.Time
}

func NewStore() *Store {
	return &Store{
		data:    make(map[string]string),
		expires: make(map[string]time.Time),
	}
}

func (s *Store) isExpired(key string) bool {
	expireAt, ok := s.expires[key]
	if !ok {
		return false
	}

	return time.Now().After(expireAt)
}

func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	delete(s.expires, key)
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isExpired(key) {
		delete(s.data, key)
		delete(s.expires, key)
		return "", false
	}

	value, ok := s.data[key]
	return value, ok
}

func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if !ok {
		return false
	}

	delete(s.data, key)
	delete(s.expires, key)
	return true
}

func (s *Store) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isExpired(key) {
		delete(s.data, key)
		delete(s.expires, key)
		return false
	}

	_, ok := s.data[key]
	return ok
}

func (s *Store) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		if s.isExpired(key) {
			delete(s.data, key)
			delete(s.expires, key)
			continue
		}

		keys = append(keys, key)
	}

	return keys
}

func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.data {
		if s.isExpired(key) {
			delete(s.data, key)
			delete(s.expires, key)
		}
	}

	return len(s.data)
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.data {
		delete(s.data, key)
	}

	for key := range s.expires {
		delete(s.expires, key)
	}
}

func (s *Store) Expire(key string, seconds int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isExpired(key) {
		delete(s.data, key)
		delete(s.expires, key)

		return false
	}

	_, ok := s.data[key]
	if !ok {
		return false
	}

	s.expires[key] = time.Now().Add(time.Duration(seconds) * time.Second)
	return true
}

func (s *Store) TTL(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isExpired(key) {
		delete(s.data, key)
		delete(s.expires, key)

		return -2
	}

	_, ok := s.data[key]
	if !ok {
		return -2
	}

	expireAt, ok := s.expires[key]
	if !ok {
		return -1
	}

	remaining := time.Until(expireAt).Seconds()
	if remaining < 0 {
		return -2
	}

	return int(remaining)
}
