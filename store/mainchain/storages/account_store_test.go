package storages

import (
	"testing"

	"math/big"

	"fmt"

	"bytes"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

func TestAccountState(t *testing.T) {
	dbdir := "test/account1"
	store, _ := NewAccoutStore(dbdir, false)
	as := &states.AccountState{
		Address: mock.Address_1,
		Height:  1,
		Data: &states.Account{
			Nonce:       1,
			Balance:     big.NewInt(2000),
			EnergyBalance: big.NewInt(3000),
		},
	}
	buff := bytes.NewBuffer(nil)
	buff.Write(as.GetKey())
	as.Serialize(buff)
	asTmp := &states.AccountState{}
	asTmp.Deserialize(buff)
	assert.Equal(t, as, asTmp)
	err := store.Save(as)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	asNew, err := store.GetPrev(2, mock.Address_1)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	fmt.Println(asNew.Height, asNew.Data.Balance)
}
