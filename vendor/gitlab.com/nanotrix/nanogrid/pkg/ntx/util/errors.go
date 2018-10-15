package util

func ErrToString(err error) string {
	if err == nil {
		return ""
	} else {
		return err.Error()
	}
}

func NewNotFoundError(msg string) NotFoundError {
	return NotFoundError{msg}
}

type NotFoundError struct {
	msg string
}

func (e NotFoundError) Error() string {
	return e.msg
}

func IsNotFound(e error) bool {
	if e == nil {
		return false
	}

	_, ok := e.(NotFoundError)
	return ok
}
