package orgchain

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
)

type TxReq struct {
	Tx       *types.Transaction
	RespChan chan *TxResp
}

type TxResp struct {
	Hash common.Hash `json:"hash"`
	Desc string `json:"desc"`
}

type MustPackTxsReq struct {
	Txs []*types.Transaction
}
