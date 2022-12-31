package errors

type UniqueViolationStorage struct {
	Err error
}

func (uv UniqueViolationStorage) Error() string {
	return "something already exists"
}

func (uv UniqueViolationStorage) Unwrap() error {
	return uv.Err
}

func NewUniqueViolationStorage(err error) error {
	return &UniqueViolationStorage{
		Err: err,
	}
}
