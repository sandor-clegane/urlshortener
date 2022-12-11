package shortener

import (
	"context"
	"net/url"

	"github.com/sandor-clegane/urlshortener/internal/common"
)

var _ URLshortenerService = &urlshortenerServiceImpl{}

type URLshortenerService interface {
	shorten(_ *url.URL) (*url.URL, error)
	ShortenURL(ctx context.Context, userID, url string) (string, error)
	ExpandURL(ctx context.Context, urlID string) (string, error)
	GetAllURL(ctx context.Context, userID string) ([]common.PairURL, error)
}
