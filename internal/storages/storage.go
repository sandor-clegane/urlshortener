package storages

import (
	"context"

	"github.com/sandor-clegane/urlshortener/internal/common"
	"github.com/sandor-clegane/urlshortener/internal/config"
)

var _ Storage = &InMemoryStorage{}
var _ Storage = &FileStorage{}
var _ Storage = &dbStorage{}

type Storage interface {
	LookUp(ctx context.Context, str string) (string, bool)
	Insert(ctx context.Context, key, value, userID string)
	InsertSome(ctx context.Context, expandURLwIDslice []common.PairURL, userID string) error
	GetPairsByID(ctx context.Context, userID string) ([]common.PairURL, bool)
}

func CreateStorage(cfg config.Config) Storage {
	if cfg.DatabaseDSN == config.DefaultDatabaseDSN {
		if cfg.FileStoragePath == config.DefaultFileStoragePath {
			return NewInMemoryStorage()
		} else {
			return NewFileStorage(cfg.FileStoragePath)
		}
	} else {
		return NewDBStorage(cfg.DatabaseDSN)
	}
}
