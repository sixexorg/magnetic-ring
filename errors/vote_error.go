package errors

import "errors"

var (
	ERR_VOTE_DATA_EOF      = errors.New("vote data eof")
	ERR_VOTE_NOT_FOUND     = errors.New("vote not found")
	ERR_VOTE_ALREADY_VOTED = errors.New("The vote has already been taken.")
	ERR_VOTE_OUT_OF_RANGE  = errors.New("The voting time is not within the specified time interval")
)
