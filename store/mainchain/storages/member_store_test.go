package storages

import (
	"os"
	"testing"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

func TestMemberStore(t *testing.T) {
	dbDir := "./testMember"
	os.RemoveAll(dbDir)
	store, _ := NewMemberStore(dbDir)
	leagueId := common.Address{1, 2, 3}
	account := common.Address{2, 3, 4}
	sls := states.LeagueMembers{
		&states.LeagueMember{
			LeagueId: leagueId,
			Height:   10,
			Data: &states.LeagueAccount{
				Account: account,
				Status:  states.LAS_Apply,
			},
		},
	}
	store.NewBatch()
	store.BatchSave(sls)
	err := store.CommitTo()
	if err != nil {
		t.Error(err)
		return
	}
	lm, err := store.GetPrev(15, leagueId, account)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(lm.LeagueId.ToString(), lm.Height, lm.Data.Status, lm.Data.Account.ToString())
}
