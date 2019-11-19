package storages

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

type PublicLedger interface {
	GetCurrentHeaderHash() common.Hash
	GetCurrentBlock() (uint64, common.Hash)
	GetCurrentBlockHash() common.Hash
	GetCurrentBlockHeight() uint64
	TotalDifficulty() *big.Int
	GetDifficultyByBlockHash(hash common.Hash) (*big.Int, error)
	RollbackToHeight(height uint64) error
	SaveAll(block *types.Block, accountStates storelaw.AccountStaters) error
}
