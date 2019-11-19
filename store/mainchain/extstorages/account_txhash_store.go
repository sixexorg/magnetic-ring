package extstorages

import (
	"encoding/binary"

	"bytes"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type AccountHashStore struct {
	dbDir string
	store *db.LevelDBStore
}

type AccountHash struct {
	TxHash common.Hash
	Index  uint32
}

func NewAccountHashStore(dbDir string) (*AccountHashStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	accoutHashStore := &AccountHashStore{
		dbDir: dbDir,
		store: store,
	}
	return accoutHashStore, nil
}

func (this *AccountHashStore) BatchSave(account common.Address, txHashes common.HashArray) {
	idx := this.getIndex(account)
	for _, v := range txHashes {
		key := this.getKey(account, idx)
		this.store.BatchPut(key, v.ToBytes())
		idx++
	}
}
func (this *AccountHashStore) GetRangeHashes(account common.Address, pageSize uint32, prevIndex uint32, esc bool) []*AccountHash {
	var start, end uint32

	if esc {
		start = prevIndex + 1
		end = prevIndex + pageSize
	} else {
		end = prevIndex - 1
		if prevIndex < pageSize {
			end = 0
		} else {
			start = prevIndex - pageSize
		}
	}
	prefixS := this.getKey(account, start)
	prefixE := this.getKey(account, end)

	iter := this.store.NewSeniorIterator(&util.Range{Start: prefixS, Limit: prefixE})
	ahs := make([]*AccountHash, 0, pageSize)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		index := key[1+common.AddrLength:]
		txHash, _ := common.ParseHashFromBytes(value)
		ahs = append(ahs, &AccountHash{Index: binary.LittleEndian.Uint32(index), TxHash: txHash})
	}
	return ahs
}

func (this *AccountHashStore) getIndex(address common.Address) uint32 {
	prefix := bytes.NewBuffer(nil)
	prefix.WriteByte(byte(mcom.ACCOUNT_HASHES))
	prefix.Write(address[:])
	iter := this.store.NewIterator(prefix.Bytes())
	if iter.Last() {
		key := iter.Key()
		l := len(key)
		idxbyte := key[l-4 : l]
		idx := binary.LittleEndian.Uint32(idxbyte)
		return idx
	}
	return 0
}

func (this *AccountHashStore) getKey(address common.Address, index uint32) []byte {
	buff := make([]byte, 1+common.AddrLength+4)
	buff[0] = byte(mcom.ACCOUNT_HASHES)
	copy(buff[1:common.AddrLength+1], address[:])
	binary.LittleEndian.PutUint32(buff[common.AddrLength+1:], index)
	return buff
}
