package extstates

import (
	"sort"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
)

func Block2LeagueBlock(block *types.Block) *LeagueBlockSimple {
	if block == nil {
		return nil
	}
	hashes := make(common.HashArray, 0, block.Transactions.Len())
	for _, v := range block.Transactions {
		hashes = append(hashes, v.Hash())
	}
	sort.Sort(hashes)
	lbs := &LeagueBlockSimple{
		Header:   block.Header,
		TxHashes: hashes,
	}
	return lbs
}
