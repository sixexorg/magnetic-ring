package resp

import "github.com/sixexorg/magnetic-ring/core/mainchain/types"

type RespGetTxByHash struct {
	Tx *types.Transaction
}
