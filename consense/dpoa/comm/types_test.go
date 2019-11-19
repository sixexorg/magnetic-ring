package comm

import (
	"testing"
	"time"

	"math/big"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

func TestBlock(t *testing.T) {

	hash1, _ := common.ParseHashFromBytes([]byte{1, 2, 3, 4, 5})
	sd := &types.SigData{
		TimeoutSigs: [][]byte{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}},
		FailerSigs:  [][]byte{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}},
		ProcSigs:    [][]byte{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}},
	}
	txs := types.Transactions{
		{
			Version: types.TxVersion,
			TxType:  types.JoinLeague,
			TxData: &types.TxData{
				Nonce:    1,
				From:     common.Address{1, 2, 3},
				Account:  common.Address{1, 2, 3},
				LeagueId: common.Address{3, 4, 5},
				Fee:      big.NewInt(123),
				MinBox:   20,
			},
		},
	}
	block := &types.Block{
		Header: &types.Header{
			Version:          0x01,
			PrevBlockHash:    hash1,
			BlockRoot:        hash1,
			TxRoot:           hash1,
			LeagueRoot:       hash1,
			StateRoot:        hash1,
			ReceiptsRoot:     hash1,
			Timestamp:        uint64(time.Now().Unix()),
			Height:           50,
			ConsensusPayload: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		},
		Transactions: txs,
		Sigs:         sd,
	}
	blk := &Block{
		Block: block,
		Info: &VbftBlockInfo{
			View:               1,
			Miner:              "miner",
			Proposer:           2,
			VrfValue:           []byte{1, 2, 3},
			VrfProof:           []byte{4, 5, 6},
			LastConfigBlockNum: 9999,
			NewChainConfig: &ChainConfig{
				Peers: []string{"p", "e", "e", "r", "s"},
			},
		},
	}
	blk.Info.NewChainConfig = nil
	buff, err := blk.Serialize()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	blkGet := &Block{}
	err = blkGet.Deserialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, blk.Block.Transactions.Len(), blkGet.Block.Transactions.Len())
	assert.Equal(t, blk.Block.Transactions[0].TxData, blkGet.Block.Transactions[0].TxData)
	assert.Equal(t, blk.Block.Hash(), blkGet.Block.Hash())
	assert.Equal(t, blk.Info, blkGet.Info)
	assert.Equal(t, blk.Block.Sigs, blkGet.Block.Sigs)
}
