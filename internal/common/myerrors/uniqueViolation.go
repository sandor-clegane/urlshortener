package myerrors

import (
	"fmt"
)

type UniqueViolation struct {
	Err             error
	ExistedShortURL string
}

func (uv UniqueViolation) Error() string {
	return fmt.Sprintf("URL %s already exists in database", uv.ExistedShortURL)
}

func (uv UniqueViolation) Unwrap() error {
	return uv.Err
}

func NewUniqueViolation(existedURL string, err error) error {
	return &UniqueViolation{
		ExistedShortURL: existedURL,
		Err:             err,
	}
}
