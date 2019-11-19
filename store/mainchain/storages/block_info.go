package storages

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

type LeagueKV struct {
	K common.Address
	V common.Hash
}

type BlockInfo struct {
	Block         *types.Block
	Receipts      types.Receipts
	AccountStates states.AccountStates
	LeagueStates  states.LeagueStates
	Members       states.LeagueMembers
	LeagueKVs     []*LeagueKV
	FeeSum        *big.Int
	ObjTxs        types.Transactions
	GasDestroy    *big.Int
}

func (this *BlockInfo) Height() uint64 {
	return this.Block.Header.Height
}
