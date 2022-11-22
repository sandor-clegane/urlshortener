package app

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

//general storage interface
type Storage interface {
	LookUp(str string) (string, bool)
	Insert(key, value string)
}

//MemoryStorage impl
type InMemoryStorage struct {
	storage map[string]string
	lock    sync.RWMutex
}

func (s *InMemoryStorage) LookUp(str string) (string, bool) {
	s.lock.RLock()
	res, ok := s.storage[strings.TrimPrefix(str, "/")]
	s.lock.RUnlock()

	if !ok {
		return "", ok
	}
	return res, ok
}

func (s *InMemoryStorage) Insert(key, value string) {
	s.lock.Lock()
	s.storage[strings.TrimPrefix(key, "/")] = value
	s.lock.Unlock()
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		storage: make(map[string]string),
	}
}

type Record struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

//TODO файл поменять на jsonEncoder от этого файла
type FileStorage struct {
	f *os.File
	*InMemoryStorage
}

func (fs *FileStorage) Insert(key, value string) {
	fs.lock.Lock()
	fs.storage[strings.TrimPrefix(key, "/")] = value
	_ = json.NewEncoder(fs.f).Encode(&Record{Key: key, Value: value})
	fs.lock.Unlock()
}

func NewFileStorage(fileName string) *FileStorage {
	file, _ := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)

	fs := &FileStorage{
		InMemoryStorage: NewInMemoryStorage(),
		f:               file,
	}

	dec := json.NewDecoder(file)
	for dec.More() {
		var r Record
		_ = dec.Decode(&r)
		fs.storage[r.Key] = r.Value
	}

	return fs
}
