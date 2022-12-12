package storages

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sandor-clegane/urlshortener/internal/common"
)

type FileStorage struct {
	enc *json.Encoder
	*InMemoryStorage
}

type record struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (fs *FileStorage) Insert(_ context.Context, key, value, userID string) error {
	trimmedKey := strings.TrimPrefix(key, "/")
	r := record{Key: key, Value: value}

	fs.lock.Lock()
	_, isExists := fs.storage[trimmedKey]
	if isExists {
		return fmt.Errorf("Key %s already exists", key)
	}
	err := fs.enc.Encode(&r)
	if err != nil {
		return err
	}
	fs.storage[trimmedKey] = value
	fs.userToKeys[userID] = append(fs.userToKeys[userID], value)
	fs.lock.Unlock()

	return nil
}

func (fs *FileStorage) InsertSome(ctx context.Context, expandURLwIDslice []common.PairURL, userID string) error {
	var r record

	fs.lock.Lock()
	for _, p := range expandURLwIDslice {
		trimmedKey := strings.TrimPrefix(p.ShortURL, "/")
		r = record{Key: trimmedKey, Value: p.ExpandURL}
		fs.storage[trimmedKey] = p.ExpandURL
		err := fs.enc.Encode(&r)
		if err != nil {
			return err
		}
		fs.userToKeys[userID] = append(fs.userToKeys[userID], trimmedKey)
	}
	fs.lock.Unlock()

	return nil
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
	var r record

	for dec.More() {
		err = dec.Decode(&r)
		if err != nil {
			log.Fatal(err)
		}
		fs.storage[r.Key] = r.Value
	}

	return fs
}
