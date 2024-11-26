package apiutils

import "errors"

var (
	ErrUpdateObjectEmpty = errors.New("update object should have at least one field")
)

type HTTPError struct {
	Message string `json:"error,omitempty"`
}
