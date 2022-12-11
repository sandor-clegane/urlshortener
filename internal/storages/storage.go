package storages

import (
	"context"

	"github.com/sandor-clegane/urlshortener/internal/common"
	"github.com/sandor-clegane/urlshortener/internal/config"
)

var _ Storage = &InMemoryStorage{}
var _ Storage = &FileStorage{}

type Storage interface {
	LookUp(ctx context.Context, str string) (string, bool)
	Insert(ctx context.Context, key, value, userID string)
	GetPairsByID(ctx context.Context, userID string) ([]common.PairURL, bool)
}

func CreateStorage(cfg config.Config) Storage {
	if cfg.FileStoragePath == "" {
		return NewInMemoryStorage()
	} else {
		return NewFileStorage(cfg.FileStoragePath)
	}
}
