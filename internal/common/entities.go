package common

import (
	"encoding/json"
	url2 "net/url"
)

type PairURL struct {
	ShortURL  string `json:"short_url"`
	ExpandURL string `json:"original_url"`
}

type InMessage struct {
	ExpandURL url2.URL `json:"url"`
}

type OutMessage struct {
	ShortURL string `json:"result"`
}

type DeletableURL struct {
	ShortURL  string
	ExpandURL string
	UserID    string
	IsDeleted bool
}

func (im *InMessage) UnmarshalJSON(data []byte) error {
	aliasValue := &struct {
		RawURL string `json:"url"`
	}{}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	url, err := url2.Parse(aliasValue.RawURL)
	if err != nil {
		return err
	}

	im.ExpandURL = *url
	return nil
}

type PairURLwithCIDin struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type PairURLwithCIDout struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
