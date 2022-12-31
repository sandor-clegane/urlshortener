package myerrors

import (
	"fmt"
)

type DeleteViolation struct {
	Err error
	URL string
}

func (dv DeleteViolation) Error() string {
	return fmt.Sprintf("URL %s has been deleted", dv.URL)
}

func (dv DeleteViolation) Unwrap() error {
	return dv.Err
}

func NewDeleteViolation(deletedURL string, err error) error {
	return &DeleteViolation{
		URL: deletedURL,
		Err: err,
	}
}
