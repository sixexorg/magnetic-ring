package storages

import (
	"math/big"

	"encoding/binary"

	"bytes"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/store/db"
	scom "github.com/sixexorg/magnetic-ring/store/orgchain/common"
)

type UTStore struct {
	dbDir string           //Store file path
	store *db.LevelDBStore //Store handler
}

func NewUTStore(dbDir string) (*UTStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	utStore := &UTStore{
		dbDir: dbDir,
		store: store,
	}
	return utStore, nil
}

func (this *UTStore) GetUTByHeight(height uint64) *big.Int {
	key := this.getKey(height)
	iter := this.store.NewIterator(nil)
	buff := bytes.NewBuffer(nil)
	flag := true
	if iter.Seek(key) {
		if bytes.Compare(iter.Key(), key) == 0 {
			buff.Write(iter.Value())
			flag = false
		}
	}
	if flag && iter.Prev() {
		buff.Write(iter.Value())
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return big.NewInt(0)
	}
	cp, err := serialization.ReadComplex(buff)
	if err != nil {
		return big.NewInt(0)
	}
	bg, _ := cp.ComplexToBigInt()
	return bg
}
func (this *UTStore) Save(height uint64, ut *big.Int) error {
	fmt.Printf("⛔️league ut save height:%d ut:%v\n", height, ut)
	key := this.getKey(height)
	sk := sink.NewZeroCopySink(nil)
	cp, _ := sink.BigIntToComplex(ut)
	sk.WriteComplex(cp)
	return this.store.Put(key, sk.Bytes())
}

func (this *UTStore) getKey(height uint64) []byte {
	buff := make([]byte, 9)
	buff[0] = byte(scom.ST_VOTE_STATE)
	binary.LittleEndian.PutUint64(buff[1:], height)
	return buff
}
