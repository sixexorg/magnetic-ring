package vm_storages

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/db"
	"github.com/sixexorg/magnetic-ring/vm/cvm"
)

type LValueStore struct {
	dbDir string
	store *db.LevelDBStore
}

func NewLValueStore() (*LValueStore, error) {
	var err error
	store, err := db.NewLevelDBStore("")
	if err != nil {
		return nil, err
	}
	lValueStore := &LValueStore{
		dbDir: "",
		store: store,
	}
	return lValueStore, nil
}

func (this *LValueStore) Save(lv cvm.LValue, address common.Address) error {
	data := cvm.LVSerialize(lv)
	return this.store.Put(address[:], data)
}

func (this *LValueStore) Get(address common.Address) (cvm.LValue, error) {
	data, err := this.store.Get(address[:])
	if err != nil {
		return nil, err
	}
	return cvm.LVDeserialize(data), nil
}
