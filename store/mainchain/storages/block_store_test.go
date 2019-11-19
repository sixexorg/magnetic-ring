/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package storages

import (
	"crypto/sha256"
	"testing"

	"fmt"
	"os"

	"time"

	"bytes"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
)

var (
	testBlockStore   *BlockStore
	testAccountStore *AccountStore
)

func TestMain(m *testing.M) {
	var err error
	testBlockDir := "test/block"
	testBlockStore, err = NewBlockStore(testBlockDir, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewBlockStore error %s\n", err)
		return
	}
	testStateDir := "test/account"
	testAccountStore, err = NewAccoutStore(testStateDir, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewAccoutStore error %s\n", err)
		return
	}
	m.Run()
	err = testBlockStore.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testBlockStore.Close error %s\n", err)
		return
	}
	err = os.RemoveAll("./test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.RemoveAll error %s\n", err)
		return
	}
}
func TestVersion(t *testing.T) {
	testBlockStore.NewBatch()
	version := byte(1)
	err := testBlockStore.SaveVersion(version)
	if err != nil {
		t.Errorf("SaveVersion error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	v, err := testBlockStore.GetVersion()
	if err != nil {
		t.Errorf("GetVersion error %s", err)
		return
	}
	if version != v {
		t.Errorf("TestVersion failed version %d != %d", v, version)
		return
	}
}

func TestCurrentBlock(t *testing.T) {
	blockHash := common.Hash(sha256.Sum256([]byte("123456789")))
	t.Logf("%x", blockHash)
	blockHeight := uint64(1)
	testBlockStore.NewBatch()
	err := testBlockStore.SaveCurrentBlock(blockHeight, blockHash)
	if err != nil {
		t.Errorf("SaveCurrentBlock error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, height, err := testBlockStore.GetCurrentBlock()
	if hash != blockHash {
		t.Errorf("TestCurrentBlock BlockHash %x != %x", hash, blockHash)
		return
	}
	if height != blockHeight {
		t.Errorf("TestCurrentBlock BlockHeight %x != %x", height, blockHeight)
		return
	}
}

func TestBlockHash(t *testing.T) {
	blockHash := common.Hash(sha256.Sum256([]byte("123456789")))
	blockHeight := uint64(1)
	testBlockStore.NewBatch()
	testBlockStore.SaveBlockHash(blockHeight, blockHash)
	blockHash = common.Hash(sha256.Sum256([]byte("234567890")))
	blockHeight = uint64(2)
	testBlockStore.SaveBlockHash(blockHeight, blockHash)
	err := testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}
	hash, err := testBlockStore.GetBlockHash(blockHeight)
	if err != nil {
		t.Errorf("GetBlockHash error %s", err)
		return
	}
	if hash != blockHash {
		t.Errorf("TestBlockHash failed BlockHash %x != %x", hash, blockHash)
		return
	}
}

func TestSaveTransaction(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	tx := mock.Mock_Tx_CreateLeague
	err := tx.Serialize(buff)
	if err != nil {
		t.Fail()
		t.Errorf("Serialize error %s", err)
		return
	}

	txHash := mock.Mock_Tx_CreateLeague.Hash()

	exist, err := testBlockStore.ContainTransaction(txHash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if exist {
		t.Errorf("TestSaveTransaction ContainTransaction should be false.")
		return
	}

	testBlockStore.NewBatch()
	err = testBlockStore.SaveTransaction(tx, 1)
	if err != nil {
		t.Errorf("SaveTransaction error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	tx1, height, err := testBlockStore.GetTransaction(txHash)
	if err != nil {
		t.Errorf("GetTransaction error %s", err)
		return
	}
	if 1 != height {
		t.Errorf("TestSaveTransaction failed BlockHeight %d != %d", height, 1)
		return
	}
	if tx1.TxType != tx.TxType {
		t.Errorf("TestSaveTransaction failed TxType %d != %d", tx1.TxType, tx.TxType)
		return
	}
	tx1Hash := tx1.Hash()
	if txHash != tx1Hash {
		t.Errorf("TestSaveTransaction failed TxHash %x != %x", tx1Hash, txHash)
		return
	}

	exist, err = testBlockStore.ContainTransaction(txHash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestSaveTransaction ContainTransaction should be true.")
		return
	}
}

func TestHeaderIndexList(t *testing.T) {
	testBlockStore.NewBatch()
	startHeight := uint64(0)
	size := uint64(100)
	indexMap := make(map[uint64]common.Hash, size)
	indexList := make([]common.Hash, 0)
	for i := startHeight; i < size; i++ {
		hash := common.Hash(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	err := testBlockStore.SaveHeaderIndexList(startHeight, indexList)
	if err != nil {
		t.Errorf("SaveHeaderIndexList error %s", err)
		return
	}
	startHeight = uint64(100)
	indexMap = make(map[uint64]common.Hash, size)
	for i := startHeight; i < size; i++ {
		hash := common.Hash(sha256.Sum256([]byte(fmt.Sprintf("%v", i))))
		indexMap[i] = hash
		indexList = append(indexList, hash)
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	totalMap, err := testBlockStore.GetHeaderIndexList()
	if err != nil {
		t.Errorf("GetHeaderIndexList error %s", err)
		return
	}

	for height, hash := range indexList {
		h, ok := totalMap[uint64(height)]
		if !ok {
			t.Errorf("TestHeaderIndexList failed height:%d hash not exist", height)
			return
		}
		if hash != h {
			t.Errorf("TestHeaderIndexList failed height:%d hash %x != %x", height, hash, h)
			return
		}
	}
}

func TestSaveHeader(t *testing.T) {
	//bookkeeper := mock.Mock_Address_1
	header := &types.Header{
		Version:          123,
		PrevBlockHash:    common.Hash{},
		TxRoot:           common.Hash{},
		Timestamp:        uint64(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
		Height:           uint64(1),
		ConsensusData:    123456789,
		ConsensusPayload: []byte("123456789"),
		//NextBookkeeper: bookkeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{},
	}
	blockHash := block.Hash()

	testBlockStore.NewBatch()
	err := testBlockStore.SaveHeader(block)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	h, err := testBlockStore.GetHeader(blockHash)
	if err != nil {
		t.Errorf("GetHeader error %s", err)
		return
	}

	headerHash := h.Hash()
	if blockHash != headerHash {
		t.Errorf("TestSaveHeader failed HeaderHash \r\n %x \r\n %x", headerHash, blockHash)
		return
	}

	if header.Height != h.Height {
		t.Errorf("TestSaveHeader failed Height %d \r\n %d", h.Height, header.Height)
		return
	}
	fmt.Println(string(h.ConsensusPayload))
}

func TestBlock(t *testing.T) {

	header := &types.Header{
		Version:        123,
		PrevBlockHash:  common.Hash{},
		TxRoot:         common.Hash{},
		Timestamp:      uint64(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix()),
		Height:         2,
		ConsensusData:  1234567890,
		NextBookkeeper: mock.Mock_Address_1,
	}
	buff := bytes.NewBuffer(nil)
	tx := mock.Mock_Tx_CreateLeague
	tx2 := mock.Mock_Tx
	err := tx.Serialize(buff)
	block := &types.Block{
		Header:       header,
		Transactions: []*types.Transaction{tx, tx2},
		Sigs: &types.SigData{
			TimeoutSigs: [][]byte{{1, 2, 3}, {4, 5, 6}},
			FailerSigs:  [][]byte{{1, 2, 3}, {4, 5, 6}},
			ProcSigs:    [][]byte{{1, 2, 3}, {4, 5, 6}},
		},
	}
	blockHash := block.Hash()
	tx1Hash := tx.Hash()

	testBlockStore.NewBatch()

	err = testBlockStore.SaveBlock(block)
	if err != nil {
		t.Errorf("SaveHeader error %s", err)
		return
	}
	err = testBlockStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo error %s", err)
		return
	}

	b, err := testBlockStore.GetBlock(blockHash)
	if err != nil {
		t.Errorf("GetBlock error %s", err)
		return
	}

	hash := b.Hash()
	if hash != blockHash {
		t.Errorf("TestBlock failed BlockHash %x != %x ", hash, blockHash)
		return
	}
	exist, err := testBlockStore.ContainTransaction(tx1Hash)
	if err != nil {
		t.Errorf("ContainTransaction error %s", err)
		return
	}
	if !exist {
		t.Errorf("TestBlock failed transaction %x should exist", tx1Hash)
		return
	}

	if len(block.Transactions) != len(b.Transactions) {
		t.Errorf("TestBlock failed Transaction size %d != %d ", len(b.Transactions), len(block.Transactions))
		return
	}
	if b.Transactions[0].Hash() != tx1Hash {
		t.Errorf("TestBlock failed transaction1 hash %x != %x", b.Transactions[0].Hash(), tx1Hash)
		return
	}
	assert.Equal(t, block.Sigs, b.Sigs)
}

func TestGetTx(t *testing.T) {
	dbdir := "/Users/rennbon/magnetict/node4/magneticchain/block"
	bkstore, err := NewBlockStore(dbdir, false)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	for i := uint64(1); i < 20; i++ {
		bkHash, err := bkstore.GetBlockHash(i)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}
		bk, err := bkstore.GetBlock(bkHash)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}
		fmt.Println(bk.Header.Height, bk.Transactions.Len())

	}

	/*	txHash, _ := common.StringToHash("yvt4vx3ge2sf1azggbf3vj9vaheebfrc2wf8ej33agznw4znxvyh====")

		tx, height, err := bkstore.GetTransaction(txHash)
		if err != nil {
			t.Error(err)
			t.Fail()
			return
		}
		t.Log(tx, height)*/
}
