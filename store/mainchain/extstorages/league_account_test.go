package extstorages

import (
	"testing"

	"os"

	"fmt"

	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
)

func TestGetAccountByHeight(t *testing.T) {
	account, err := accountStore.GetAccountByHeight(mock.Address_1, mock.LeagueAddress1, 1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log(account)
}

func TestGetPrev(t *testing.T) {
	dbDir := "./testAccount"
	os.RemoveAll(dbDir)
	leagueId := common.Address{1, 2, 3}
	account := common.Address{3, 4, 5}
	store, _ := NewExternalLeague(dbDir, false)

	las := &extstates.LeagueAccountState{
		Address:  account,
		LeagueId: leagueId,
		Height:   30,
		Data: &extstates.Account{
			Nonce:       2,
			Balance:     big.NewInt(20000),
			EnergyBalance: big.NewInt(23333),
			BonusHeight: 0,
		},
	}
	err := store.Save(las)
	if err != nil {
		t.Error(err)
		return
	}
	ass, err := store.GetPrev(50, account, leagueId)
	fmt.Println(ass, err)
}
