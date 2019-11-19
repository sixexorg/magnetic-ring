package errors

import "errors"

var (
	ERR_TXPOOL_TXNOTFOUND         = errors.New("transaction not found")
	ERR_TXPOOL_LOWFEE             = errors.New("tx fee is too low")
	ERR_TXPOOL_UNINIT             = errors.New("txpool has not init")
	ERR_TXPOOL_OUTOFMAX           = errors.New("server is busy,retry please")
	ERR_TXPOOL_UNDERPRICED        = errors.New("ErrUnderpriced")
	ERR_TXPOOL_REPLACEUNDERPRICED = errors.New("Replace Underpriced")
	ERR_TX_SIGN = errors.New("signature validate error")
)
