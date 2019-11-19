package mainchain

import (
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

type TxReq struct {
	Tx       *types.Transaction
	RespChan chan *TxResp
}

type TxResp struct {
	Hash string
	Desc string
}

type MustPackTxsReq struct {
	Txs []*types.Transaction
}
