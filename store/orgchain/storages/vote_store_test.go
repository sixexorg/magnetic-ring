package storages

import (
	"math/big"
	"os"
	"testing"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

func TestVoteStore(t *testing.T) {
	dbDir := "./testVote"
	os.RemoveAll(dbDir)
	store, _ := NewVoteStore(dbDir, 0x01, 0x02)
	voteId := common.Hash{1, 2, 3}
	sv := &storelaw.VoteState{
		VoteId:     voteId,
		Height:     10,
		Agree:      big.NewInt(10),
		Against:    big.NewInt(20),
		Abstention: big.NewInt(15),
	}
	store.NewBatch()
	store.SaveVotes([]*storelaw.VoteState{sv})
	err := store.CommitTo()
	if err != nil {
		t.Error(err)
		return
	}
	vs2, err := store.GetVoteState(voteId, 30)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(vs2)
}
