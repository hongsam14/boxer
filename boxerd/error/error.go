package error

import "fmt"

type BoxerErrorCode int

const (
	SystemError BoxerErrorCode = iota
	InternalError
	InvalidConfig
	InvalidArgument
	InvalidState
	InvalidOperation
	Timeout
)

type BoxerError struct {
	Code   BoxerErrorCode
	Origin error
	Msg    string
}

func (e BoxerError) Error() string {
	return fmt.Sprintf("Error: %s\n\t: %s", e.Msg, e.Origin.Error())
}
