package types_test

import (
	"bytes"
	"testing"

	"time"

	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

func TestBlockLife(t *testing.T) {
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
	hash1, _ := common.ParseHashFromBytes([]byte{1, 2, 3, 4, 5})
	sd := &types.SigData{
		TimeoutSigs: [][]byte{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}},
		FailerSigs:  [][]byte{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}},
		ProcSigs:    [][]byte{{1, 2, 3, 4, 5}, {2, 3, 4, 5, 6}},
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
		Transactions: txs, // types.Transactions{mock.CreateLeagueTx, mock.CreateLeagueTx},
		Sigs:         sd,
	}
	buff := bytes.NewBuffer(nil)
	err := block.Serialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	buff.WriteString("hello world!!!")
	block2 := &types.Block{}
	err = block2.Deserialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log(string(buff.Bytes()))
}
