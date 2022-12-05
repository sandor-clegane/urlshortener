package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrInvalidValue = errors.New("invalid cookie value")
)

type cookieService struct {
	Key []byte
}

func New(key string) *cookieService {
	return &cookieService{Key: []byte(key)}
}

func (c *cookieService) extractValue(cookie *http.Cookie) (string, error) {
	signedValue, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}

	return string(signedValue[sha256.Size:]), nil
}

func (c *cookieService) createAndSign(w http.ResponseWriter, r *http.Request) error {
	cookie := http.Cookie{
		Name:     "userID",
		Value:    uuid.New().String(),
		HttpOnly: true,
		Secure:   false,
	}

	mac := hmac.New(sha256.New, c.Key)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	//value structure is fixed :[signature][user_id]
	cookie.Value = string(signature) + cookie.Value

	return write(w, r, cookie)
}

func (c *cookieService) checkSign(r *http.Request, name string) error {
	//[signature][user_id]
	signedValue, err := read(r, name)
	if err != nil {
		return err
	}
	if len(signedValue) < sha256.Size {
		return ErrInvalidValue
	}

	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]
	mac := hmac.New(sha256.New, c.Key)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedSignature) {
		return ErrInvalidValue
	}

	return nil
}

func write(w http.ResponseWriter, r *http.Request, cookie http.Cookie) error {
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))
	http.SetCookie(w, &cookie)
	r.AddCookie(&cookie)
	return nil
}

func read(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", ErrInvalidValue
	}

	return string(value), nil
}
