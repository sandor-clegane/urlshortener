package storages

import (
	"context"
	"strings"
	"sync"

	"github.com/sandor-clegane/urlshortener/internal/common"
)

type InMemoryStorage struct {
	storage    map[string]string
	userToKeys map[string][]string
	lock       sync.RWMutex
}

func (s *InMemoryStorage) LookUp(_ context.Context, str string) (string, bool) {
	trimmedStr := strings.TrimPrefix(str, "/")

	s.lock.RLock()
	res, ok := s.storage[trimmedStr]
	s.lock.RUnlock()

	if !ok {
		return "", ok
	}
	return res, ok
}

func (s *InMemoryStorage) Insert(_ context.Context, key, value, userID string) {
	trimmedKey := strings.TrimPrefix(key, "/")

	s.lock.Lock()
	s.storage[trimmedKey] = value
	s.userToKeys[userID] = append(s.userToKeys[userID], trimmedKey)
	s.lock.Unlock()
}

func (s *InMemoryStorage) GetPairsByID(_ context.Context, userID string) ([]common.PairURL, bool) {
	keys, ok := s.userToKeys[userID]
	if !ok {
		return nil, ok
	}
	result := make([]common.PairURL, 0, len(keys))

	for _, key := range keys {
		result = append(result, common.PairURL{
			ExpandURL: s.storage[key],
			ShortURL:  key,
		})
	}

	return result, true
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		storage:    make(map[string]string),
		userToKeys: make(map[string][]string),
	}
}
