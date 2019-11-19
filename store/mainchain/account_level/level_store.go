package account_level

import (
	"encoding/binary"

	"bytes"

	"fmt"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

//prototype pattern
type LevelStore struct {
	dbDir string
	store *db.LevelDBStore
	//cache *AccountCache
}

func NewLevelStore(dbDir string) (*LevelStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	levelStore := &LevelStore{
		dbDir: dbDir,
		store: store,
	}
	return levelStore, nil
}

func (this *LevelStore) SaveLevels(height uint64, lvls map[common.Address]EasyLevel) error {
	this.store.NewBatch()
	for k, v := range lvls {
		fmt.Printf("📈 📈 📈 levelUp account:%s level:%d\n", k.ToString(), v)
		key := this.getKey(height, k)
		sk := sink.NewZeroCopySink(nil)
		sk.WriteUint8(uint8(v))
		this.store.BatchPut(key, sk.Bytes())
	}
	err := this.store.BatchCommit()
	return err
}
func (this *LevelStore) GetAccountLvlRange(start, end uint64, account common.Address) []common.FTreer {
	s := this.getKey(start, account)
	l := this.getKey(end, account)
	iter := this.store.NewSeniorIterator(&util.Range{Start: s, Limit: l})
	lvls := []common.FTreer{}
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		height := binary.LittleEndian.Uint64(key[1+common.AddrLength:])
		lvl := EasyLevel(val[0])
		lvls = append(lvls, &HeightLevel{Height: height, Lv: lvl})
	}
	iter.Release()
	return lvls
}
func (this *LevelStore) GetAccountLevel(height uint64, account common.Address) EasyLevel {
	key := this.getKey(height, account)
	iter := this.store.NewIterator(nil)
	buff := bytes.NewBuffer(nil)
	flag := true
	if iter.Seek(key) {
		if bytes.Compare(iter.Key(), key) == 0 {
			buff.Write(iter.Key())
			buff.Write(iter.Value())
			flag = false
		}
	}
	if flag && iter.Prev() {
		buff.Write(iter.Key())
		buff.Write(iter.Value())
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return lv0
	}
	if buff.Len() == 0 {
		return lv0
	}

	buff.Next(1)
	accBytes, err := serialization.ReadBytes(buff, common.AddrLength)
	if err != nil {
		return lv0
	}
	if !bytes.Equal(accBytes, account[:]) {
		return lv0
	}
	buff.Next(8)
	lvl, err := serialization.ReadUint8(buff)
	if err != nil {
		return lv0
	}
	return EasyLevel(lvl)
}
func (this *LevelStore) getKey(height uint64, account common.Address) []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(mcom.ST_ACCOUNT)
	copy(buff[1:common.AddrLength+1], account[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], height)
	return buff
}
