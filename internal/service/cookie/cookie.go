package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrInvalidValue = errors.New("invalid cookie value")
)

type cookieServiceImpl struct {
	Key []byte
}

func New(key string) *cookieServiceImpl {
	return &cookieServiceImpl{Key: []byte(key)}
}

func (c *cookieServiceImpl) GetUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("userID")
	if err != nil {
		return "", err
	}
	signedValue, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}
	return string(signedValue[sha256.Size:]), nil
}

func (c *cookieServiceImpl) createAndSign(w http.ResponseWriter, r *http.Request) error {
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
	cookie.Value = base64.URLEncoding.
		EncodeToString([]byte(string(signature) + cookie.Value))
	http.SetCookie(w, &cookie)
	r.AddCookie(&cookie)

	return nil
}

func (c *cookieServiceImpl) checkSign(r *http.Request, name string) error {
	//[signature][user_id]
	//Get and decode value from cookie
	cookie, err := r.Cookie(name)
	if err != nil {
		return err
	}
	signedValue, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return ErrInvalidValue
	}
	if len(signedValue) < sha256.Size {
		return ErrInvalidValue
	}

	//compute expected signature value
	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	mac := hmac.New(sha256.New, c.Key)
	mac.Write([]byte(name))
	mac.Write(value)
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return ErrInvalidValue
	}

	return nil
}

func (c *cookieServiceImpl) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := c.checkSign(r, "userID")
		if err == nil {
			next.ServeHTTP(w, r)
			return
		}

		if errors.Is(err, http.ErrNoCookie) || errors.Is(err, ErrInvalidValue) {
			err = c.createAndSign(w, r)
			if err == nil {
				next.ServeHTTP(w, r)
				return
			}
		}

		io.WriteString(w, err.Error())
	})
}
