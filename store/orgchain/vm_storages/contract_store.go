package vm_storages

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/db"
)

type ContractStore struct {
	dbDir string
	store *db.LevelDBStore
}

func NewContractStore(dbDir string) (*ContractStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	voteStore := &ContractStore{
		dbDir: dbDir,
		store: store,
	}
	return voteStore, nil
}

func (this *ContractStore) Save(contract []byte, address common.Address) error {
	return this.store.Put(address[:], contract)
}

func (this *ContractStore) Get(address common.Address) ([]byte, error) {
	return this.store.Get(address[:])
}
