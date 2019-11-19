package states

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/sixexorg/magnetic-ring/common"
)

func TestAccountState_Hash(t *testing.T) {
	as := &AccountState{
		Address: common.Address{1, 2, 3},
		Height:  20,
		Data: &Account{
			Nonce:       1,
			Balance:     big.NewInt(20),
			EnergyBalance: big.NewInt(30),
			BonusHeight: 20,
		},
	}
	fmt.Println(as.Hash().String())

}
