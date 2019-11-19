package errors

import "errors"

var (
	ERR_TX_TXDATA_NIL        = errors.New("txdata is nil.")
	ERR_TX_PARAM_REQUIRD     = errors.New("param must required.")
	ERR_TX_DATA_OVERFLOW     = errors.New("data in transaction overflow.")
	ERR_SINK_TYPE_DIFF       = errors.New("complex different internal types.")
	ERR_TXRAW_EOF            = errors.New("raw eof.")
	ERR_SERIALIZATION_FAILED = errors.New("tx Serialization failed.")
	ERR_TX_INVALID           = errors.New("invalid transaction")

	ERR_TX_PROPTY_OVERFLOW = errors.New("the propty in txData  exceeds the upper limit")

	ERR_SINK_EOF = errors.New("sink eof.")
)
