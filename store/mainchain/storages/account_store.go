package storages

import (
	"math/big"

	"bytes"

	"encoding/binary"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

var (
	zero = big.NewInt(0)
)

//prototype pattern
type AccountStore struct {
	enableCache bool
	dbDir       string
	store       *db.LevelDBStore
	//cache *AccountCache
}

func NewAccoutStore(dbDir string, enableCache bool) (*AccountStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	accountStore := &AccountStore{
		dbDir:       dbDir,
		store:       store,
		enableCache: enableCache,
	}
	/*	if enableCache {
		cache, err := NewAccountCache()
		if err != nil {
			return nil, err
		}
		accountStore.cache = cache
	}*/
	return accountStore, nil
}

//NewBatch start a commit batch
func (this *AccountStore) NewBatch() {
	this.store.NewBatch()
}
func (this *AccountStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *AccountStore) GetByHeight(height uint64, account common.Address) (*states.AccountState, error) {
	key := this.getKey(account, height)
	v, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	bf := bytes.NewBuffer(key)
	bf.Write(v)
	as := &states.AccountState{}
	err = as.Deserialize(bf)
	if err != nil {
		return nil, err
	}
	return as, nil
}
func (this *AccountStore) GetAccountRange(start, end uint64, account common.Address) (states.AccountStates, error) {
	s := this.getKey(account, start)
	l := this.getKey(account, end)
	iter := this.store.NewSeniorIterator(&util.Range{Start: s, Limit: l})
	ass := states.AccountStates{}
	bf := bytes.NewBuffer(nil)
	var err error
	for iter.Next() {
		bf.Write(iter.Key())
		bf.Write(iter.Value())
		as := &states.AccountState{}
		err = as.Deserialize(bf)
		if err != nil {
			return nil, err
		}
		ass = append(ass, as)
		bf.Reset()
	}
	iter.Release()
	return ass, nil
}
func (this *AccountStore) GetPrev(height uint64, account common.Address) (*states.AccountState, error) {
	key := this.getKey(account, height)
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
		return nil, err
	}
	if buff.Len() == 0 {
		return &states.AccountState{
			Address: account,
			Data: &states.Account{
				Balance:     big.NewInt(0),
				EnergyBalance: big.NewInt(0),
			},
		}, nil
	}
	as := &states.AccountState{}
	err = as.Deserialize(buff)
	if err != nil {
		return nil, err
	}

	if as.Address.Equals(account) {
		return as, nil
	}
	return &states.AccountState{
		Address: account,
		Data: &states.Account{
			Balance:     big.NewInt(0),
			EnergyBalance: big.NewInt(0),
		},
	}, nil
}

//SaveBlock persist block to store
func (this *AccountStore) Save(account *states.AccountState) error {
	key := account.GetKey()
	buff := new(bytes.Buffer)
	err := account.Serialize(buff)
	if err != nil {
		return err
	}
	err = this.store.Put(key, buff.Bytes())
	if err != nil {
		return err

	}
	return nil
}
func (this *AccountStore) BatchSave(states states.AccountStates) error {
	for _, v := range states {
		key := v.GetKey()
		buff := new(bytes.Buffer)
		err := v.Serialize(buff)
		if err != nil {
			return err
		}
		this.store.BatchPut(key, buff.Bytes())

	}
	return nil
}

func (this *AccountStore) getKey(account common.Address, height uint64) []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(mcom.ST_ACCOUNT)
	copy(buff[1:common.AddrLength+1], account[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], height)
	return buff
}
