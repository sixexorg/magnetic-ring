package extstorages

import (
	"bytes"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/store/db"
	scom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type ExtFullTXStore struct {
	dbDir string
	store *db.LevelDBStore
}

func NewExtFullTX(dbDir string) (*ExtFullTXStore, error) {
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	el := new(ExtFullTXStore)

	el.dbDir = dbDir
	el.store = store
	return el, nil
}
func (this *ExtFullTXStore) GetTx(leagueId common.Address, hash common.Hash) (tx *types.Transaction, height uint64, err error) {
	key := this.getKey(leagueId, hash)
	val, err := this.store.Get(key)
	if err != nil {
		return nil, 0, err
	}
	tx = &types.Transaction{}
	buff := bytes.NewBuffer(val)
	err = tx.Deserialize(buff)
	if err != nil {
		return nil, 0, err
	}
	height, err = serialization.ReadUint64(buff)
	if err != nil {
		return nil, 0, err
	}
	return tx, height, nil
}

func (this *ExtFullTXStore) SaveTxs(leagueId common.Address, height uint64, txs types.Transactions) error {
	this.store.NewBatch()
	for _, v := range txs {
		key := this.getKey(leagueId, v.Hash())
		sk := sink.NewZeroCopySink(nil)
		err := v.Serialization(sk)
		if err != nil {
			return err
		}
		sk.WriteUint64(height)
		this.store.BatchPut(key, sk.Bytes())
	}
	err := this.store.BatchCommit()
	return err
}

func (this *ExtFullTXStore) getKey(leagueId common.Address, hash common.Hash) []byte {
	buff := make([]byte, 1+common.AddrLength+common.HashLength)
	buff[0] = byte(scom.EXT_VOTE_TX)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	copy(buff[common.AddrLength+1:], hash[:])
	return buff
}
