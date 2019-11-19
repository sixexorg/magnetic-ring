package storages

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type LeagueBirthCertStore struct {
	dbDir string
	store *db.LevelDBStore
}

func NewLeagueBirthCertStore(dbDir string) (*LeagueBirthCertStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	model := &LeagueBirthCertStore{
		dbDir: dbDir,
		store: store,
	}
	return model, nil
}
func (this *LeagueBirthCertStore) NewBatch() {
	this.store.NewBatch()
}
func (this *LeagueBirthCertStore) CommitTo() error {
	return this.store.BatchCommit()
}
func (this *LeagueBirthCertStore) BatchPutBirthCert(leagueId common.Address, txHash common.Hash) {
	key := this.getKey(leagueId)
	this.store.BatchPut(key, txHash.ToBytes())
}
func (this *LeagueBirthCertStore) GetBirthCert(leagueId common.Address) (common.Hash, error) {
	key := this.getKey(leagueId)
	val, err := this.store.Get(key)
	if err != nil {
		return common.Hash{}, err
	}
	return common.ParseHashFromBytes(val)
}
func (this *LeagueBirthCertStore) getKey(leagueId common.Address) []byte {
	buff := make([]byte, 1+common.AddrLength)
	buff[0] = byte(mcom.LEAGUE_BIRTH_CERT)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	return buff
}
