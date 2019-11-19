package extstorages

import (
	"testing"

	"github.com/sixexorg/magnetic-ring/mock"
)

func TestBlockHashStore(t *testing.T) {

	hashes, gasSum, err := store.GetBlockHashSpan(mock.LeagueAddress1, 2, 4)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	for k, v := range hashes {
		t.Logf("key:%d ,val:%s\n", k, v.String())
	}
	t.Log("gasSum:", gasSum.Uint64())
}

func TestGetBlock(t *testing.T) {
	blk3, err := store.GetBlock(mock.LeagueAddress1, 3)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log(blk3.Header, blk3.EnergyUsed, blk3.TxHashes)
}
