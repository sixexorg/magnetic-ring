package storages

import (
	"testing"

	"math/big"

	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

func TestAccountCache(t *testing.T) {
	accahe, err := NewAccountCache()
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}

	as := &states.AccountState{
		Address: mock.Address_1,
		Data: &states.Account{
			Nonce:       5,
			Balance:     big.NewInt(20),
			EnergyBalance: big.NewInt(200),
			BonusHeight: 0,
		}}

	accahe.AddState(as)

	state := accahe.GetState(mock.Address_1)
	if state != nil {
		t.Log(state)
	}
	as2 := &states.AccountState{
		Address: mock.Address_1,
		Data: &states.Account{
			Nonce:       7,
			Balance:     big.NewInt(230),
			EnergyBalance: big.NewInt(300),
			BonusHeight: 0,
		}}
	accahe.AddState(as2)
	state2 := accahe.GetState(mock.Address_1)
	if state2 != nil {
		t.Log(state2)
	}
}
