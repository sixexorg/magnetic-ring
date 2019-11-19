package poa

import "errors"

var (
	ErrUnknownAncestor = errors.New("unknown ancestor")

	ErrFutureBlock = errors.New("block in the future")

	ErrInvalidNumber = errors.New("invalid block number")
)

