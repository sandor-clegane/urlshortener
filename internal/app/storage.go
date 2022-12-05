package app

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
)

var _ Storage = &InMemoryStorage{}
var _ Storage = &FileStorage{}

//general storage interface
type Storage interface {
	LookUp(str string) (string, bool)
	Insert(key, value, userID string)
	GetPairsByID(userID string) ([]PairURL, bool)
}

//MemoryStorage impl
type InMemoryStorage struct {
	storage    map[string]string
	userToKeys map[string][]string
	lock       sync.RWMutex
}

func (s *InMemoryStorage) LookUp(str string) (string, bool) {
	trimmedStr := strings.TrimPrefix(str, "/")

	s.lock.RLock()
	res, ok := s.storage[trimmedStr]
	s.lock.RUnlock()

	if !ok {
		return "", ok
	}
	return res, ok
}

func (s *InMemoryStorage) Insert(key, value, userID string) {
	trimmedKey := strings.TrimPrefix(key, "/")

	s.lock.Lock()
	s.storage[trimmedKey] = value
	s.userToKeys[userID] = append(s.userToKeys[userID], trimmedKey)
	s.lock.Unlock()
}

func (s *InMemoryStorage) GetPairsByID(userID string) ([]PairURL, bool) {
	keys, ok := s.userToKeys[userID]
	if !ok {
		return nil, ok
	}
	result := make([]PairURL, 0, len(keys))

	for _, key := range keys {
		result = append(result, PairURL{
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

//file storage impl
type FileStorage struct {
	enc *json.Encoder
	*InMemoryStorage
}

type Record struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (fs *FileStorage) Insert(key, value, userID string) {
	trimmedKey := strings.TrimPrefix(key, "/")
	r := Record{Key: key, Value: value}

	fs.lock.Lock()
	fs.storage[trimmedKey] = value
	_ = fs.enc.Encode(&r)
	fs.userToKeys[userID] = append(fs.userToKeys[userID], value)
	fs.lock.Unlock()
}

func NewFileStorage(fileName string) *FileStorage {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	fs := &FileStorage{
		InMemoryStorage: NewInMemoryStorage(),
		enc:             json.NewEncoder(file),
	}

	dec := json.NewDecoder(file)
	for dec.More() {
		var r Record
		err = dec.Decode(&r)
		if err != nil {
			log.Fatal(err)
		}
		fs.storage[r.Key] = r.Value
	}

	return fs
}
