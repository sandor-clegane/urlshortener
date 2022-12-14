package storages

import (
	"context"
	"encoding/json"
	"fmt"
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
	defer fs.lock.Unlock()
	_, isExists := fs.storage[trimmedKey]
	if isExists {
		return fmt.Errorf("key %s already exists", key)
	}
	err := fs.enc.Encode(&r)
	if err != nil {
		return err
	}
	fs.storage[trimmedKey] = value
	fs.userToKeys[userID] = append(fs.userToKeys[userID], trimmedKey)

	return nil
}

func (fs *FileStorage) InsertSome(_ context.Context, expandURLwIDslice []common.PairURL, userID string) error {
	var r record

	fs.lock.Lock()
	defer fs.lock.Unlock()
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
	return nil
}

func NewFileStorage(fileName string) (*FileStorage, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	ims, err := NewInMemoryStorage()
	if err != nil {
		return nil, err
	}
	fs := &FileStorage{
		InMemoryStorage: ims,
		enc:             json.NewEncoder(file),
	}

	dec := json.NewDecoder(file)
	var r record

	for dec.More() {
		err = dec.Decode(&r)
		if err != nil {
			return nil, err
		}
		fs.storage[r.Key] = r.Value
	}

	return fs, nil
}
