package storages_test

import (
	"os"
	"testing"

	"math/big"
	"time"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

func TestLedger(t *testing.T) {
	var err error
	dbdir := "test/ledger"
	defer os.RemoveAll(dbdir)
	ledger, err := storages.NewLedgerStore(dbdir)
	if err != nil {
		return
	}
	address_1, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	as := &states.AccountState{
		Address: address_1,
		Height:  1,
		Data: &states.Account{
			Nonce:       1,
			Balance:     big.NewInt(0).SetUint64(1000),
			EnergyBalance: big.NewInt(0).SetUint64(1000),
		},
	}
	sts := states.AccountStates{as}

	txs := make(types.Transactions, 0, 2)
	//todo hello world msg
	block := &types.Block{
		Header: &types.Header{
			Height:        1,
			Version:       0x01,
			PrevBlockHash: common.Hash{},
			LeagueRoot:    common.Hash{},
			ReceiptsRoot:  common.Hash{},
			StateRoot:     sts.GetHashRoot(),
			TxRoot:        txs.GetHashRoot(),
			Timestamp:     uint64(time.Now().Unix()),
		},
		Transactions: txs,
	}
	blockInfo := &storages.BlockInfo{
		Block:         block,
		AccountStates: sts,
	}
	err = ledger.SaveAll(blockInfo)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	block1, err := ledger.GetBlockByHeight(1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	fmt.Println(block1.Header.Height)
	fmt.Println(ledger.GetCurrentBlockHeight(), ledger.GetCurrentBlockHash())
	blockHash, err := ledger.GetBlockHashByHeight(1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	fmt.Println("blockHash:", blockHash)
	//fmt.Println(ledger.GetAccountStore().GetAccount(address_1).Data.Balance.Uint64())
}

func TestAddHeaders(t *testing.T) {
	var err error
	dbdir := "test/ledger"
	defer os.RemoveAll(dbdir)
	ledger, err := storages.NewLedgerStore(dbdir)
	if err != nil {
		return
	}
	address_1, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	as := &states.AccountState{
		Address: address_1,
		Height:  1,
		Data: &states.Account{
			Nonce:       1,
			Balance:     big.NewInt(0).SetUint64(1000),
			EnergyBalance: big.NewInt(0).SetUint64(1000),
		},
	}
	sts := states.AccountStates{as}

	txs := make(types.Transactions, 0, 2)
	//todo hello world msg

	Header := &types.Header{
		Height:        1,
		Version:       0x01,
		PrevBlockHash: common.Hash{},
		LeagueRoot:    common.Hash{},
		ReceiptsRoot:  common.Hash{},
		StateRoot:     sts.GetHashRoot(),
		TxRoot:        txs.GetHashRoot(),
		Timestamp:     uint64(time.Now().Unix()),
	}
	err = ledger.AddHeaders([]*types.Header{Header})
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	blockHash, err := ledger.GetBlockHashByHeight(1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log(blockHash)
}
