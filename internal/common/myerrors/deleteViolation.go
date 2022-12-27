package myerrors

import (
	"fmt"

	"github.com/sandor-clegane/urlshortener/internal/common"
)

type DeleteViolation struct {
	Err  error
	Data common.DeletableURL
}

func (dv DeleteViolation) Error() string {
	return fmt.Sprintf("URL %s has been deleted", dv.Data.ExpandURL)
}

func (dv DeleteViolation) Unwrap() error {
	return dv.Err
}

func NewDeleteViolation(deletableURL common.DeletableURL, err error) error {
	return &DeleteViolation{
		Data: deletableURL,
		Err:  err,
	}
}
