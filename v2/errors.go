package v2

import (
	"fmt"
)

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("error Wappin code:%d, title:%s and details:%s", e.Code, e.Title, e.Details)
}

// CastError casts the response to Wappin error.
func CastError(code int, title string, details string) *Error {
	return &Error{
		Code:    code,
		Title:   title,
		Details: details,
	}
}

func getError(statusCode int, errors []Error) (err error) {
	if statusCode == 200 {
		return
	}

	if len(errors) > 0 {
		code := errors[0].Code
		title := errors[0].Title
		details := errors[0].Details

		err = CastError(code, title, details)
		return
	}

	return
}
