package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Store struct {
	mu      sync.RWMutex
	data    map[string]string
	expires map[string]time.Time
}

type storeSnapshot struct {
	Data    map[string]string    `json:"data"`
	Expires map[string]time.Time `json:"expires"`
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

func (s *Store) deleteExpiredLocked() int {
	deleted := 0

	for key := range s.data {
		if s.isExpired(key) {
			delete(s.data, key)
			delete(s.expires, key)
			deleted++
		}
	}

	return deleted
}

func (s *Store) DeleteExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.deleteExpiredLocked()
}

func (s *Store) StartActiveExpiration(interval time.Duration) func() {
	ticker := time.NewTicker(interval)
	stop := make(chan struct{})

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.DeleteExpired()
			case <-stop:
				return
			}
		}
	}()

	return func() {
		close(stop)
	}
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

func (s *Store) Save(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deleteExpiredLocked()

	snapshot := storeSnapshot{
		Data:    make(map[string]string, len(s.data)),
		Expires: make(map[string]time.Time, len(s.expires)),
	}

	for key, value := range s.data {
		snapshot.Data[key] = value
	}

	for key, expireAt := range s.expires {
		snapshot.Expires[key] = expireAt
	}

	bytes, err := json.MarshalIndent(snapshot, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, bytes, 0644)
}

func (s *Store) Load(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var snapshot storeSnapshot
	if err := json.Unmarshal(bytes, &snapshot); err != nil {
		return err
	}

	if snapshot.Data == nil {
		snapshot.Data = make(map[string]string)
	}

	if snapshot.Expires == nil {
		snapshot.Expires = make(map[string]time.Time)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = snapshot.Data
	s.expires = snapshot.Expires
	s.deleteExpiredLocked()

	return nil
}
