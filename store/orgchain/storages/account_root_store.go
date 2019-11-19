package storages

import (
	"bytes"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/sixexorg/magnetic-ring/store/db"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
)

//prototype pattern
type AccountRootStore struct {
	dbDir string
	store *db.LevelDBStore
}

func NewAccountRootStore(dbDir string) (*AccountRootStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	accountRootStore := &AccountRootStore{
		dbDir: dbDir,
		store: store,
	}
	return accountRootStore, nil
}

//NewBatch start a commit batch
func (this *AccountRootStore) NewBatch() {
	this.store.NewBatch()
}
func (this *AccountRootStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *AccountRootStore) Get(height uint64) (*states.AccountStateRoot, error) {
	key := states.GetAccountStateRootKey(height)
	v, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	ast := &states.AccountStateRoot{}
	bu := bytes.NewBuffer(key)
	bu.Write(v)
	err = ast.Deserialize(bu)
	if err != nil {
		return nil, err
	}
	return ast, nil
}

func (this *AccountRootStore) GetRange(start, end uint64) ([]*states.AccountStateRoot, error) {
	sh := states.GetAccountStateRootKey(start)
	eh := states.GetAccountStateRootKey(end)
	iter := this.store.NewSeniorIterator(&util.Range{Start: sh, Limit: eh})
	asrs := make([]*states.AccountStateRoot, 0, end-start+1)
	for iter.Next() {
		asr := &states.AccountStateRoot{}
		buff := bytes.NewBuffer(iter.Key())
		buff.Write(iter.Value())
		asr.Deserialize(buff)
		asrs = append(asrs, asr)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}
	return asrs, nil
}

func (this *AccountRootStore) Save(entity *states.AccountStateRoot) error {
	key := entity.GetKey()
	buff := new(bytes.Buffer)
	err := entity.Serialize(buff)
	if err != nil {
		return err
	}
	err = this.store.Put(key, buff.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (this *AccountRootStore) BatchRemove(heights []uint64) {
	for _, v := range heights {
		key := states.GetAccountStateRootKey(v)
		this.store.BatchDelete(key)
	}
}
