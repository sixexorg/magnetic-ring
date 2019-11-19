package storages

import (
	"math/big"

	"bytes"

	"fmt"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/db"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

var (
	zero = big.NewInt(0)
)

//prototype pattern
type AccountStore struct {
	dbDir string
	store *db.LevelDBStore
}

func NewAccountStore(dbDir string) (*AccountStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	accountStore := &AccountStore{
		dbDir: dbDir,
		store: store,
	}
	return accountStore, nil
}

//NewBatch start a commit batch
func (this *AccountStore) NewBatch() {
	this.store.NewBatch()
}
func (this *AccountStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *AccountStore) GetByHeight(height uint64, account common.Address) (storelaw.AccountStater, error) {
	key := states.GetAccountStateKey(account, height)
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
func (this *AccountStore) GetPrev(height uint64, account, leagueId common.Address) (storelaw.AccountStater, error) {
	key := states.GetAccountStateKey(account, height)
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
func (this *AccountStore) GetAccountRange(start, end uint64, account common.Address) (storelaw.AccountStaters, error) {
	s := states.GetAccountStateKey(account, start)
	l := states.GetAccountStateKey(account, end)
	iter := this.store.NewSeniorIterator(&util.Range{Start: s, Limit: l})
	ass := storelaw.AccountStaters{}
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

//SaveBlock persist block to store
func (this *AccountStore) Save(state storelaw.AccountStater) error {
	key := state.GetKey()
	buff := new(bytes.Buffer)
	err := state.Serialize(buff)
	if err != nil {
		return err
	}
	err = this.store.Put(key, buff.Bytes())
	if err != nil {
		return err

	}
	return nil
}
func (this *AccountStore) BatchSave(states storelaw.AccountStaters) error {
	for _, v := range states {
		fmt.Println("⭕️ accountSave ", v.Balance().Uint64(), v.Energy().Uint64())
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

func (this *AccountStore) BatchRemove(height uint64, addresses []common.Address) {
	for _, v := range addresses {
		key := states.GetAccountStateKey(v, height)
		this.store.BatchDelete(key)
	}
}
