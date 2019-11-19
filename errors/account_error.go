package errors

import "errors"

var (
	ERR_ACCT_FMT1              = errors.New("wrong address format 1")
	ERR_ACCT_FMT2              = errors.New("wrong address format 2")
	ERR_ACCT_FMT3              = errors.New("wrong address format 3")
	ERR_ACCT_FMT4              = errors.New("wrong address format 4")
	ERR_ACCT_CRT_MULADDR       = errors.New("pubks can not less than m")
	ERR_ACCT_MULADDR_NOT_FOUND = errors.New("multiple account not found")
)
