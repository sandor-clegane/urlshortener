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

type FileStorage struct {
	p *producer
	s *InMemoryStorage
}

func (s *FileStorage) LookUp(str string) (string, bool) {
	return s.LookUp(str)
}

func (s *FileStorage) Insert(key, value string) {
	s.Insert(key, value)

	s.s.lock.Lock()
	s.p.WriteRecord(&Record{Key: key, Value: value})
	s.s.lock.Unlock()
}

func NewFileStorage(fileName string) *FileStorage {
	fs := &FileStorage{
		s: NewInMemoryStorage(),
		p: NewProducer(fileName),
	}

	file, _ := os.Open(fileName)
	dec := json.NewDecoder(file)
	for dec.More() {
		var r Record
		_ = dec.Decode(&r)
		fs.s.Insert(r.Key, r.Value)
	}

	return fs
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(fileName string) *producer {
	file, _ := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}
}
func (p *producer) WriteRecord(record *Record) error {
	return p.encoder.Encode(&record)
}
func (p *producer) Close() error {
	return p.file.Close()
}
