package storages

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
)

type FileStorage struct {
	enc *json.Encoder
	*InMemoryStorage
}

type Record struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (fs *FileStorage) Insert(_ context.Context, key, value, userID string) {
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
