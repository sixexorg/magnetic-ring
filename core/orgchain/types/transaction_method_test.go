package types_test

import (
	"bytes"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
)

func TestTransactionLife(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	err := mock.Block2.Transactions[3].Serialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	tx := &types.Transaction{}
	err = tx.Deserialize(buff)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, mock.Block2.Transactions[3].Hash(), tx.Hash())
}

func TestTransactionDeepCopy(t *testing.T) {
	txCopy := &types.Block{}
	err := common.DeepCopy(txCopy, mock.Block2)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, txCopy.Hash(), mock.Block2.Hash())
}
