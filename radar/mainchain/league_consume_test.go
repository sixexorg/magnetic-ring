package mainchain

import (
	"os"
	"testing"
	"time"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstorages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

type MockTell struct {
}

func (MockTell) Tell(message interface{}) {

}
func mockTellInit() (bactor.Teller, error) {
	return &MockTell{}, nil
}
func ValidateBlock(block *orgtypes.Block, ledgerStore storelaw.Ledger4Validation) (blkInfo *storelaw.OrgBlockInfo, err error) {
	return &storelaw.OrgBlockInfo{
		Block: block,
	}, nil
}
func TestEntireProcess(t *testing.T) {
	dbdir := "test"
	defer os.RemoveAll(dbdir)
	stroer, err := extstorages.NewLedgerStore(dbdir)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	adpter := &ConsumerAdapter{
		FuncGenesis: func(address common.Address) (*types.Transaction, uint64, error) {
			return mock.CreateLeagueTx, 0, nil
		},
		FuncBlk: func(height uint64) (*types.Header, error) {
			return &types.Header{Timestamp: 1000}, nil
		},
		FuncTx:        mock.GetTransaction,
		FuncValidate:  ValidateBlock,
		Ledger:        stroer,
		P2pTlrInit:    mockTellInit,
		P2pTlr:        MockTell{},
		TxPoolTlrInit: mockTellInit,
		TxPoolTlr:     MockTell{},
	}
	lgc := NewLeagueConsumers(adpter)
	lgc.ReceiveBlock(mock.Blocks[0])
	lgc.ReceiveBlock(mock.Blocks[1])
	//lgc.Identify()
	time.Sleep(2 * time.Second)
	lgc.Settlement()
	t.Log("------------part1--------------")
	for _, v := range lgc.txs {
		t.Logf("leagueId:%s\nstartheight:%d\nendheight:%d\nfee:%s\nblockroot:%s\n",
			v.TxData.LeagueId.ToString(),
			v.TxData.StartHeight,
			v.TxData.EndHeight,
			v.TxData.Energy.String(),
			v.TxData.BlockRoot.String())
	}
	leagueHMap := make(map[common.Address]uint64)
	for k, _ := range lgc.consumers {
		leagueHMap[k] = 2
		break
	}
	mainTxs, err := lgc.generateMainTx(leagueHMap)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log("------------part2--------------")
	for _, v := range mainTxs {
		t.Logf("leagueId:%s\nstartheight:%d\nendheight:%d\nfee:%s\nblockroot:%s\n",
			v.TxData.LeagueId.ToString(),
			v.TxData.StartHeight,
			v.TxData.EndHeight,
			v.TxData.Energy.String(),
			v.TxData.BlockRoot.String())
	}
	if lgc.txs.Len() > 0 {
		lsp := NewLeagueStatePipe()
		lgc.CheckLeaguesResult(lgc.txs, lsp)
		leagueCount := len(lgc.consumers)
		count := 0
		for {
			select {
			case sign := <-lsp.StateSignal:
				switch msg := sign.(type) {
				case *LeagueNeed:
					t.Log("LeagueNeed", msg)
				case *LeagueExec:
					t.Log("LeagueExec", msg)
				case *LeagueErr:
					t.Log("LeagueErr", msg)
				}
			case <-lsp.Successed:
				count++
				if count == leagueCount {
					t.Log("Check OK")
					goto END

				}
			}
		}
	END:
		lsp.Kill()
	}
	for k, _ := range leagueHMap {
		t.Logf("leagueId:%s,lastHeight:%d\n", k.ToString(), lgc.consumers[k].getLastHeight())
	}
}
func TestIntercept(t *testing.T) {
	h1 := [5]uint64{100, 50, 20, 6, 8}
	h2 := [5]uint64{200, 300, 400, 500, 600}
	h3 := [5]uint64{900, 20, 3000, 250, 76}
	h4 := [5]uint64{88, 99, 33, 77, 66}

	m1 := make(map[common.Address]uint64)
	m1[mock.Address_1] = h1[0]
	m1[mock.Address_2] = h2[0]
	m1[mock.Address_3] = h3[0]
	m1[mock.Address_4] = h4[0]

	m2 := make(map[common.Address]uint64)
	m2[mock.Address_1] = h1[1]
	m2[mock.Address_2] = h2[1]
	m2[mock.Address_3] = h3[1]
	m2[mock.Address_4] = h4[1]

	m3 := make(map[common.Address]uint64)
	m3[mock.Address_1] = h1[2]
	m3[mock.Address_2] = h2[2]
	m3[mock.Address_3] = h3[2]
	m3[mock.Address_4] = h4[2]

	m4 := make(map[common.Address]uint64)
	m4[mock.Address_1] = h1[3]
	m4[mock.Address_2] = h2[3]
	m4[mock.Address_3] = h3[3]
	m4[mock.Address_4] = h4[3]

	m5 := make(map[common.Address]uint64)
	m5[mock.Address_1] = h1[4]
	m5[mock.Address_2] = h2[4]
	m5[mock.Address_3] = h3[4]
	m5[mock.Address_4] = h4[4]

	// map 会随机,当前排列可能出现2种中位数
	//20, 200, 20, 33
	//6,200,250,77

	m := make(map[common.Address]uint64)
	m[mock.Address_1] = 100 //20 = 20  //6 = 6
	m[mock.Address_2] = 50  //200 = 50 //200 = 50
	m[mock.Address_3] = 30  //20 =20   //250 = 30
	m[mock.Address_4] = 200 //33 =33   //77 = 77

	ls := &LeagueConsumers{}
	result := ls.interceptHeight(m, m1, m2, m3, m4, m5)
	real := make(map[common.Address]uint64)
	if result[mock.Address_1] == 20 {
		real[mock.Address_1] = min(m[mock.Address_1], 20)
		real[mock.Address_2] = min(m[mock.Address_2], 200)
		real[mock.Address_3] = min(m[mock.Address_3], 20)
		real[mock.Address_4] = min(m[mock.Address_4], 33)

	} else {
		real[mock.Address_1] = min(m[mock.Address_1], 6)
		real[mock.Address_2] = min(m[mock.Address_2], 200)
		real[mock.Address_3] = min(m[mock.Address_3], 250)
		real[mock.Address_4] = min(m[mock.Address_4], 77)
	}
	assert.Equal(t, result[mock.Address_1], real[mock.Address_1])
	assert.Equal(t, result[mock.Address_2], real[mock.Address_2])
	assert.Equal(t, result[mock.Address_3], real[mock.Address_3])
	assert.Equal(t, result[mock.Address_4], real[mock.Address_4])
}

func min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}
func TestInitRadar(t *testing.T) {
	adpter := &ConsumerAdapter{
		P2pTlrInit:    mockTellInit,
		P2pTlr:        MockTell{},
		TxPoolTlrInit: mockTellInit,
		TxPoolTlr:     MockTell{},
	}
	adpter.initTeller()
	fmt.Println(adpter.TlrReady)
	adpter.initTeller()
	fmt.Println(adpter.TlrReady)
}
