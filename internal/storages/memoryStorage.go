package storages

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sandor-clegane/urlshortener/internal/common"
)

type InMemoryStorage struct {
	storage    map[string]string
	userToKeys map[string][]string
	lock       sync.RWMutex
}

func (s *InMemoryStorage) LookUp(_ context.Context, str string) (string, error) {
	trimmedStr := strings.TrimPrefix(str, "/")

	s.lock.RLock()
	res, ok := s.storage[trimmedStr]
	s.lock.RUnlock()

	if !ok {
		return "", fmt.Errorf("no %s short URL in database", str)
	}
	return res, nil
}

func (s *InMemoryStorage) Insert(_ context.Context, key, value, userID string) error {
	trimmedKey := strings.TrimPrefix(key, "/")

	s.lock.Lock()
	_, isExists := s.storage[trimmedKey]
	if isExists {
		return fmt.Errorf("key %s already exists", key)
	}
	s.storage[trimmedKey] = value
	s.userToKeys[userID] = append(s.userToKeys[userID], trimmedKey)
	s.lock.Unlock()

	return nil
}

func (s *InMemoryStorage) InsertSome(_ context.Context, expandURLwIDslice []common.PairURL, userID string) error {
	s.lock.Lock()
	for _, p := range expandURLwIDslice {
		trimmedKey := strings.TrimPrefix(p.ShortURL, "/")
		s.storage[trimmedKey] = p.ExpandURL
		s.userToKeys[userID] = append(s.userToKeys[userID], trimmedKey)
	}
	s.lock.Unlock()

	return nil
}

func (s *InMemoryStorage) GetPairsByID(_ context.Context, userID string) ([]common.PairURL, error) {
	keys, ok := s.userToKeys[userID]
	if !ok {
		return nil, fmt.Errorf("user with ID %s did not shorten any URL", userID)
	}
	result := make([]common.PairURL, 0, len(keys))

	for _, key := range keys {
		result = append(result, common.PairURL{
			ExpandURL: s.storage[key],
			ShortURL:  key,
		})
	}

	return result, nil
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		storage:    make(map[string]string),
		userToKeys: make(map[string][]string),
	}
}
