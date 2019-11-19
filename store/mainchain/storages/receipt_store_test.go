package storages_test

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

func TestReceiptStore(t *testing.T) {
	dbdir := "test"
	defer os.RemoveAll(dbdir)
	store, _ := storages.NewReceiptStore(dbdir, false)
	txHash, _ := common.StringToHash("6s3q52rn2fhv38ssk14zkx4zsk9rctv3ugw7tapanu25ucskb18h====")
	receipt := &types.Receipt{
		Status:  true,
		TxHash:  txHash,
		GasUsed: 1000,
	}
	store.NewBatch()
	store.BatchSave(types.Receipts{receipt})
	err := store.CommitTo()
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	receiptGet, err := store.GetReceipt(txHash)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	assert.Equal(t, receipt, receiptGet)
}
