package orgchain

import (
	"testing"

	"os"

	"log"

	"fmt"

	"path"

	"github.com/magiconair/properties/assert"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
	mainstorages "github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/orgchain/genesis"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

func TestEntireProcess(t *testing.T) {
	dbdir := "test"
	defer os.RemoveAll(dbdir)
	ledger, _ := storages.NewLedgerStore(path.Join(dbdir, "org"))
	mainLedger, _ := mainstorages.NewLedgerStore(path.Join(dbdir, "main"))
	tx := mock.CreateLeagueTx
	mainBlock := &maintypes.Block{
		Header: &maintypes.Header{
			Height:    1,
			Timestamp: mock.OrgTimeSpan,
		},
		Transactions: maintypes.Transactions{tx},
	}
	mainLedger.SaveBlockForMockTest(mainBlock)

	radar := NewRadar()
	txCh := make(chan *maintypes.Transaction)
	go func() { txCh <- tx }()
	block, _, _, _, _, sts := radar.MonitorGenesis(txCh, mainstorages.GetLightLedger(), tx.TxData.From)
	state := genesis.NewGensisState(mock.CreateLeagueTx)
	sts2 := states.AccountStates{state}
	stroot := sts2.GetHashRoot()
	mock.Block1.Header.StateRoot = stroot
	assert.Equal(t, stroot, sts.GetHashRoot())

	blocknew := mock.Block1
	fmt.Println(blocknew.Header.LeagueId.ToString())
	if block.Hash() != blocknew.Hash() {
		t.Fail()
		t.Error("diffent between mock and executed")
		t.Log("\nblock1:", block.Hash().String(), "\nblock2", mock.Block1.Hash().String())
	}
	//err := accountStore.Save(state)
	err := ledger.SaveAccount(sts[0])
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}

	blk1 := &storelaw.OrgBlockInfo{
		Block: block,
	}
	blk2 := &storelaw.OrgBlockInfo{
		Block: mock.Block2,
	}
	ledger.SaveAll(blk1)
	ledger.SaveAll(blk2)

	//there can do a for loop here
	re := radar.RoutingMainTx(mock.ConfirmTx)
	select {
	case orgTx := <-re.ORGTX:
		t.Log(orgTx.Hash().String())
		return
	case mcc := <-re.MCC:
		blocks, err := radar.GetBlocks(mcc.Start, mcc.End)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}
		result := radar.VerifyMainAndLeague(blocks)
		if !result {
			t.Log("failed")
			t.Fail()
			return
		}
	case err := <-re.Err:
		log.Println(err.Error())
		return
	}

}
