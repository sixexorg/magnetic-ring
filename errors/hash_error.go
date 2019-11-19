package errors

import "errors"

var (
	ERR_HASH_PARSE = errors.New("The bytes length is not 32.")
)
