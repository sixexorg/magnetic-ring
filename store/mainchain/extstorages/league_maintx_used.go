package extstorages

import (
	"bytes"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/store/db"
	dbcommon "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type MainTxUsed struct {
	dbDir string
	store *db.LevelDBStore
}

func NewMainTxUsed(dbDir string) (*MainTxUsed, error) {
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	el := &MainTxUsed{
		dbDir: dbDir,
		store: store,
	}
	return el, nil
}

//NewBatch start a commit batch
func (this *MainTxUsed) NewBatch() {
	this.store.NewBatch()
}
func (this *MainTxUsed) CommitTo() error {
	return this.store.BatchCommit()
}
func (this *MainTxUsed) Exist(txHash common.Hash) bool {
	key := this.getKey(txHash)
	bl, err := this.store.Has(key)
	if err != nil {
		return false
	}
	return bl
}

func (this *MainTxUsed) BatchSave(txHashes common.HashArray, height uint64) error {
	buff := bytes.NewBuffer(nil)
	err := serialization.WriteUint64(buff, height)
	if err != nil {
		return err
	}
	for _, v := range txHashes {
		key := this.getKey(v)
		this.store.BatchPut(key, buff.Bytes())
	}
	return nil
}

func (this *MainTxUsed) getKey(txHash common.Hash) []byte {
	buff := make([]byte, 1+common.HashLength)
	buff[0] = byte(dbcommon.EXT_LEAGUE_MAIN_USED)
	copy(buff[1:common.AddrLength+1], txHash[:])
	return buff
}
