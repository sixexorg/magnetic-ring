package extstates

import (
	"math/big"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
)

func TestAccountState_Hash(t *testing.T) {
	add := common.Address{1, 2, 3}
	h := uint64(20)
	nonce := uint64(1)
	bal := big.NewInt(20)
	energy := big.NewInt(2000)
	bonus := uint64(20)
	as := &states.AccountState{
		Address: add,
		Height:  h,
		Data: &states.Account{
			Nonce:       nonce,
			Balance:     bal,
			EnergyBalance: energy,
			BonusHeight: bonus,
		},
	}
	las := LeagueAccountState{
		Address: add,
		Height:  h,
		Data: &Account{
			Nonce:       nonce,
			Balance:     bal,
			EnergyBalance: energy,
			BonusHeight: bonus,
		},
	}
	assert.Equal(t, as.Hash(), las.Hash())

}
