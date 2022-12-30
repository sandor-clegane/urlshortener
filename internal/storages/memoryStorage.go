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
	defer s.lock.RUnlock()
	res, ok := s.storage[trimmedStr]

	if !ok {
		return "", fmt.Errorf("no %s short URL in database", str)
	}
	return res, nil
}

func (s *InMemoryStorage) Insert(_ context.Context, key, value, userID string) error {
	trimmedKey := strings.TrimPrefix(key, "/")

	s.lock.Lock()
	defer s.lock.Unlock()
	_, isExists := s.storage[trimmedKey]
	if isExists {
		return fmt.Errorf("key %s already exists", key)
	}
	s.storage[trimmedKey] = value
	s.userToKeys[userID] = append(s.userToKeys[userID], trimmedKey)

	return nil
}

func (s *InMemoryStorage) InsertSome(_ context.Context, expandURLwIDslice []common.PairURL, userID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, p := range expandURLwIDslice {
		trimmedKey := strings.TrimPrefix(p.ShortURL, "/")
		s.storage[trimmedKey] = p.ExpandURL
		s.userToKeys[userID] = append(s.userToKeys[userID], trimmedKey)
	}

	return nil
}

func (s *InMemoryStorage) GetPairsByID(_ context.Context, userID string) ([]common.PairURL, error) {
	s.lock.RLock()
	keys, ok := s.userToKeys[userID]
	s.lock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("user with ID %s did not shorten any URL", userID)
	}
	result := make([]common.PairURL, 0, len(keys))

	s.lock.RLock()
	for _, key := range keys {
		result = append(result, common.PairURL{
			ExpandURL: s.storage[key],
			ShortURL:  key,
		})
	}
	s.lock.RUnlock()

	return result, nil
}

//dummy for interface implementation
func (s *InMemoryStorage) DeleteMultipleURLs(_ context.Context, _ []common.DeletableURL) error {
	return nil
}

func (s *InMemoryStorage) Ping(_ context.Context) error {
	return nil
}

func (s *InMemoryStorage) Stop() {}

func NewInMemoryStorage() (*InMemoryStorage, error) {
	return &InMemoryStorage{
		storage:    make(map[string]string),
		userToKeys: make(map[string][]string),
	}, nil
}
