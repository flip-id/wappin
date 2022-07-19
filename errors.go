package wappin

import (
	"fmt"
	"github.com/fairyhunter13/reflecthelper/v5"
	"github.com/pkg/errors"
	"net/http"
)

// List of errors used in this package.
var (
	ErrNilArguments = errors.New("nil arguments")
)

// Error represents the error for Wappin.
type Error struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("error Wappin status:%s message:%s", e.Status, e.Message)
}

// CastError casts the response to Wappin error.
func CastError(status string, message string) *Error {
	return &Error{
		Status:  status,
		Message: message,
	}
}

func getError(statusCode int, status string, message string) (err error) {
	if !(reflecthelper.GetInt(status) >= http.StatusBadRequest ||
		statusCode >= http.StatusBadRequest) {
		return
	}

	err = CastError(status, message)
	return
}
