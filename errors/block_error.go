package errors

import "errors"

var (
	ERR_BLOCK_ABNORMAL_HEIGHT = errors.New("Block height discontinuity.")
	ERR_BLOCK_PREVHASH_DIFF   = errors.New("The prevHash in the blockHeader is not consistent with the blockHash block at the specified height.")
	ERR_BLOCK_ABNORMAL_TXS    = errors.New("Block transaction does not match. Data exception.")
	ERR_PARAM_NOT_VALID       = errors.New("Parameter not valid.")
)
