package genesis

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

func NewGensisState(tx *maintypes.Transaction) *states.AccountState {
	if tx.TxType != maintypes.CreateLeague {
		return nil
	}
	balance := big.NewInt(0)
	account := &states.AccountState{
		Address: tx.TxData.From,
		Height:  1,
		Data: &states.Account{
			Nonce:       1,
			Balance:     balance.Mul(tx.TxData.MetaBox, big.NewInt(int64(tx.TxData.Rate))),
			EnergyBalance: big.NewInt(20000000000),
		},
	}
	return account
}
func GensisBlock(tx *maintypes.Transaction, createTime uint64) (*types.Block, storelaw.AccountStaters) {
	leagueId := maintypes.ToLeagueAddress(tx)
	leagueId = leagueId
	state := NewGensisState(tx)
	sts := storelaw.AccountStaters{state}

	block := &types.Block{
		Header: &types.Header{
			LeagueId:      leagueId,
			PrevBlockHash: common.Hash{},
			Version:       0x01,
			BlockRoot:     common.Hash{},
			TxRoot:        common.Hash{},
			Difficulty:    big.NewInt(10),
			Height:        1,
			Timestamp:     createTime,
			ReceiptsRoot:  common.Hash{},
			StateRoot:     sts.GetHashRoot(),
		},
	}
	return block, sts
}
