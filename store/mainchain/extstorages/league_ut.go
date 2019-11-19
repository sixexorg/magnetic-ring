package extstorages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/store/db"
	scom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type ExtUTStore struct {
	dbDir string           //Store file path
	store *db.LevelDBStore //Store handler
}

func NewExtUTStore(dbDir string) (*ExtUTStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	utStore := &ExtUTStore{
		dbDir: dbDir,
		store: store,
	}
	return utStore, nil
}

func (this *ExtUTStore) GetUTByHeight(height uint64, leagueId common.Address) *big.Int {
	key := this.getKey(height, leagueId)
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
func (this *ExtUTStore) Save(height uint64, leagueId common.Address, ut *big.Int) error {
	fmt.Printf("⛔️leagueext ut save height:%d ut:%v\n", height, ut)
	key := this.getKey(height, leagueId)
	sk := sink.NewZeroCopySink(nil)
	cp, _ := sink.BigIntToComplex(ut)
	sk.WriteComplex(cp)
	return this.store.Put(key, sk.Bytes())
}

func (this *ExtUTStore) getKey(height uint64, leagueId common.Address) []byte {
	buff := make([]byte, 9+common.AddrLength)
	buff[0] = byte(scom.EXT_UT)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], height)
	return buff
}
