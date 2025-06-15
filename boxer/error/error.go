package error

import (
	"errors"
	"fmt"
)

type BoxerErrorCode int

const (
	SystemError BoxerErrorCode = iota
	InternalError
	InvalidConfig
	InvalidArgument
	InvalidState
	InvalidOperation
	Timeout
	Full
)

type BoxerError struct {
	Code   BoxerErrorCode
	Origin error
	Msg    string
}

func (e BoxerError) Error() string {
	return fmt.Sprintf("Error: %s\n\t: %s", e.Msg, e.Origin.Error())
}

func Is(err error, code BoxerErrorCode) bool {
	var be BoxerError

	if ok := errors.As(err, &be); ok {
		return be.Code == code
	}
	return false
}
