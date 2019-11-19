package errors

import "errors"

var (
	ERR_COMMON_REFERENCE_EMPTY = errors.New("Reference is empty.")
	ERR_TIME_OUT               = errors.New("time out")
)
