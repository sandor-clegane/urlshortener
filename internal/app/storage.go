package app

import (
	"strings"
	"sync"
)

type Storage struct {
	storage map[string]string
	lock    sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		storage: make(map[string]string),
	}
}

func (s *Storage) LookUp(str string) (string, bool) {
	s.lock.RLock()
	res, ok := s.storage[strings.TrimPrefix(str, "/")]
	s.lock.RUnlock()

	if !ok {
		return "", ok
	}
	return res, ok
}

func (s *Storage) Insert(key, value string) {
	s.lock.Lock()
	s.storage[strings.TrimPrefix(key, "/")] = value
	s.lock.Unlock()
}
