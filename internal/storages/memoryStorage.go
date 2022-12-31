package storages

import (
	"context"
	"fmt"
	"sync"

	"github.com/sandor-clegane/urlshortener/internal/common"
	"github.com/sandor-clegane/urlshortener/internal/common/myerrors"
	"github.com/sandor-clegane/urlshortener/internal/storages/errors"
)

type InMemoryStorage struct {
	deletedItems map[string]struct{}
	storage      map[string]string
	userToKeys   map[string][]string
	lock         sync.RWMutex
}

func (s *InMemoryStorage) LookUp(_ context.Context, str string) (string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	res, exists := s.storage[str]
	if !exists {
		return "", fmt.Errorf("no %s short URL in database", str)
	}
	_, isDeleted := s.deletedItems[str]
	if isDeleted {
		return "", myerrors.NewDeleteViolation(res, nil)
	}
	return res, nil
}

func (s *InMemoryStorage) Insert(_ context.Context, key, value, userID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exists := s.storage[key]
	if exists {
		return errors.NewUniqueViolationStorage(nil)
	}
	s.storage[key] = value
	s.userToKeys[userID] = append(s.userToKeys[userID], key)

	return nil
}

func (s *InMemoryStorage) InsertSome(_ context.Context, expandURLwIDslice []common.PairURL, userID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, p := range expandURLwIDslice {
		s.storage[p.ShortURL] = p.ExpandURL
		s.userToKeys[userID] = append(s.userToKeys[userID], p.ShortURL)
	}

	return nil
}

func (s *InMemoryStorage) GetPairsByID(_ context.Context, userID string) ([]common.PairURL, error) {
	s.lock.RLock()
	keys, exists := s.userToKeys[userID]
	s.lock.RUnlock()

	if !exists {
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

func (s *InMemoryStorage) DeleteMultipleURLs(_ context.Context, delURLs []common.DeletableURL) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, du := range delURLs {
		s.deletedItems[du.ShortURL] = struct{}{}
	}
	return nil
}

func (s *InMemoryStorage) Ping(_ context.Context) error {
	return nil
}

func (s *InMemoryStorage) Stop() {}

func NewInMemoryStorage() (*InMemoryStorage, error) {
	return &InMemoryStorage{
		storage:      make(map[string]string),
		userToKeys:   make(map[string][]string),
		deletedItems: make(map[string]struct{}),
	}, nil
}
