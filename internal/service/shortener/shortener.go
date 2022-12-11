package shortener

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/sandor-clegane/urlshortener/internal/common"
	"github.com/sandor-clegane/urlshortener/internal/storages"
)

type urlshortenerServiceImpl struct {
	storage storages.Storage
	baseURL string
}

func New(stg storages.Storage, baseURL string) URLshortenerService {
	return &urlshortenerServiceImpl{
		storage: stg,
		baseURL: baseURL,
	}
}

func (s *urlshortenerServiceImpl) shorten(_ *url.URL) (*url.URL, error) {
	return common.Join(s.baseURL, uuid.NewString())
}

func (s *urlshortenerServiceImpl) ShortenURL(ctx context.Context, userID, rawURL string) (string, error) {
	urlParsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	shortURL, err := s.shorten(urlParsed)
	if err != nil {
		return "", err
	}
	s.storage.Insert(ctx, shortURL.Path, rawURL, userID)

	return shortURL.String(), nil
}

func (s *urlshortenerServiceImpl) ExpandURL(ctx context.Context, shortURL string) (string, error) {
	res, ok := s.storage.LookUp(ctx, shortURL)
	if !ok {
		return "", fmt.Errorf("URL found")
	}
	return res, nil
}

func (s *urlshortenerServiceImpl) GetAllURL(ctx context.Context, userID string) ([]common.PairURL, error) {
	res, ok := s.storage.GetPairsByID(ctx, userID)
	if !ok {
		return nil, fmt.Errorf("user with id %s didn`t shorten any URL", userID)
	}
	for i := 0; i < len(res); i++ {
		shortWithBase, _ := common.Join(s.baseURL, res[i].ShortURL)
		res[i].ShortURL = (*shortWithBase).String()
	}

	return res, nil
}
