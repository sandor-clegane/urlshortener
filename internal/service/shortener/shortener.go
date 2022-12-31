package shortener

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"

	"github.com/sandor-clegane/urlshortener/internal/common"
	"github.com/sandor-clegane/urlshortener/internal/common/myerrors"
	"github.com/sandor-clegane/urlshortener/internal/storages"
	errors2 "github.com/sandor-clegane/urlshortener/internal/storages/errors"
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

//TODO разделить сокращение и добавление префикса
func (s *urlshortenerServiceImpl) shorten(url *url.URL) (*url.URL, error) {
	hash := md5.Sum([]byte(url.String()))
	return common.Join(s.baseURL, hex.EncodeToString(hash[:]))
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
	err = s.storage.Insert(ctx, strings.TrimPrefix(shortURL.Path, "/"), rawURL, userID)
	var uv *errors2.UniqueViolationStorage
	if err != nil {
		if errors.As(err, uv) {
			return "", myerrors.NewUniqueViolation(shortURL.String(), err)
		}
		return "", err
	}
	return shortURL.String(), nil
}

func (s *urlshortenerServiceImpl) ExpandURL(ctx context.Context, shortURL string) (string, error) {
	res, err := s.storage.LookUp(ctx, strings.TrimPrefix(shortURL, "/"))
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *urlshortenerServiceImpl) GetAllURL(ctx context.Context, userID string) ([]common.PairURL, error) {
	res, err := s.storage.GetPairsByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(res); i++ {
		shortWithBase, _ := common.Join(s.baseURL, res[i].ShortURL)
		res[i].ShortURL = (*shortWithBase).String()
	}

	return res, nil
}

func (s *urlshortenerServiceImpl) ShortenSomeURL(ctx context.Context,
	userID string, expandURLwIDslice []common.PairURLwithCIDin) ([]common.PairURLwithCIDout, error) {
	cap := len(expandURLwIDslice)
	ResponseURLwIDslice := make([]common.PairURLwithCIDout, 0, cap)
	tempURLpairSlice := make([]common.PairURL, 0, cap)

	for _, v := range expandURLwIDslice {
		correlationID := v.CorrelationID
		urlParsed, err := url.Parse(v.OriginalURL)
		if err != nil {
			return nil, err
		}
		shortURL, err := s.shorten(urlParsed)
		if err != nil {
			return nil, err
		}

		pairURL := common.PairURL{
			ExpandURL: v.OriginalURL,
			ShortURL:  strings.TrimPrefix(shortURL.Path, "/"),
		}

		URLwCIDout := common.PairURLwithCIDout{
			CorrelationID: correlationID,
			ShortURL:      shortURL.String(),
		}

		ResponseURLwIDslice = append(ResponseURLwIDslice, URLwCIDout)
		tempURLpairSlice = append(tempURLpairSlice, pairURL)
	}

	s.storage.InsertSome(ctx, tempURLpairSlice, userID)

	return ResponseURLwIDslice, nil
}

func (s *urlshortenerServiceImpl) DeleteMultipleURLs(ctx context.Context, userID string, sliceShortID []string) error {
	var delSLiceURL = make([]common.DeletableURL, 0, len(sliceShortID))
	for _, u := range sliceShortID {
		ud := common.DeletableURL{
			ShortURL:  u,
			UserID:    userID,
			IsDeleted: true,
		}
		delSLiceURL = append(delSLiceURL, ud)
	}
	return s.storage.DeleteMultipleURLs(ctx, delSLiceURL)
}

func (s *urlshortenerServiceImpl) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
