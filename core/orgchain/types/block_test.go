package types_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
)

func TestBlockLife(t *testing.T) {
	hash1, _ := common.ParseHashFromBytes([]byte{1, 2, 3, 4, 5})
	block := &types.Block{
		Header: &types.Header{
			Version:       0x01,
			PrevBlockHash: hash1,
			BlockRoot:     hash1,
			TxRoot:        hash1,
			StateRoot:     hash1,
			ReceiptsRoot:  hash1,
			Timestamp:     uint64(time.Now().Unix()),
			Height:        50,
		},
		Transactions: types.Transactions{mock.OrgTx1},
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
