package storages

import (
	"bytes"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

//prototype pattern
type ReceiptStore struct {
	enableCache bool
	dbDir       string
	store       *db.LevelDBStore
}

func NewReceiptStore(dbDir string, enableCache bool) (*ReceiptStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	receiptStore := &ReceiptStore{
		dbDir:       dbDir,
		store:       store,
		enableCache: enableCache,
	}
	return receiptStore, nil
}

func (this *ReceiptStore) NewBatch() {
	this.store.NewBatch()
}
func (this *ReceiptStore) CommitTo() error {
	return this.store.BatchCommit()
}
func (this *ReceiptStore) BatchSave(receipts types.Receipts) {
	for _, v := range receipts {
		sk := sink.NewZeroCopySink(nil)
		sk.WriteBool(v.Status)
		sk.WriteUint64(v.GasUsed)
		this.store.Put(this.getKey(v.TxHash), sk.Bytes())
	}
}
func (this *ReceiptStore) GetReceipt(txHash common.Hash) (*types.Receipt, error) {
	key := this.getKey(txHash)
	val, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	receipt := &types.Receipt{}
	buff := bytes.NewBuffer(key[1:])
	buff.Write(val)
	err = receipt.Deserialize(buff)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (this *ReceiptStore) getKey(txHash common.Hash) []byte {
	buff := make([]byte, 1+common.HashLength)
	buff[0] = byte(mcom.ST_RECEIPT)
	copy(buff[1:common.HashLength+1], txHash[:])
	return buff
}
