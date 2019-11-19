package storages_test

import (
	"os"
	"testing"

	"math/big"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

func TestLeagueStore(t *testing.T) {
	dbDir := "./testLeague"
	os.RemoveAll(dbDir)
	store, _ := storages.NewLeagueStore(dbDir)
	leagueId := common.Address{1, 2, 3}
	la := &states.LeagueState{
		Address: leagueId,
		Height:  1,
		MinBox:  1000, // not stored in the db
		Rate:    20000,
		Private: false,
		Data: &states.League{
			Nonce:      1,
			FrozenBox:  big.NewInt(1000),
			MemberRoot: common.Hash{2, 2, 2},
		},
	}
	/*err := store.Save(la)
	if err != nil {
		t.Error(err)
		return
	}*/
	store.NewBatch()
	err := store.BatchSave(states.LeagueStates{la})
	if err != nil {
		t.Error(err)
		return
	}
	if err = store.CommitTo(); err != nil {
		t.Error(err)
		return
	}
	sl, err := store.GetPrev(2, leagueId)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(sl.Rate, sl.Data.FrozenBox)

	rate, creator, minBox, _, _, err := store.GetMetaData(leagueId)
	fmt.Println(rate, creator, minBox, err)
}
