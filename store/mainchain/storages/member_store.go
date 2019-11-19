package storages

import (
	"bytes"

	"encoding/binary"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

//prototype pattern
type MemberStore struct {
	enableCache bool
	dbDir       string
	store       *db.LevelDBStore
}

func NewMemberStore(dbDir string) (*MemberStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	memberStore := &MemberStore{
		dbDir: dbDir,
		store: store,
	}
	return memberStore, nil
}

//NewBatch start a commit batch
func (this *MemberStore) NewBatch() {
	this.store.NewBatch()
}
func (this *MemberStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *MemberStore) GetPrev(height uint64, leagueId, account common.Address) (*states.LeagueMember, error) {
	key := this.getKey(height, leagueId, account)
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
		return nil, errors.ERR_DB_NOT_FOUND
	}
	as := &states.LeagueMember{}
	err = as.Deserialize(buff)
	if err != nil {
		return nil, err
	}
	if as.LeagueId.Equals(leagueId) && as.Data.Account.Equals(account) {
		return as, nil
	}
	return nil, errors.ERR_DB_NOT_FOUND
}

//SaveBlock persist block to store
func (this *MemberStore) Save(member *states.LeagueMember) error {
	key := this.getKey(member.Height, member.LeagueId, member.Data.Account)
	err := this.store.Put(key, []byte{byte(member.Data.Status)})
	if err != nil {
		return err

	}
	return nil
}

func (this *MemberStore) BatchSave(lmstates states.LeagueMembers) error {
	for _, v := range lmstates {
		key := this.getKey(v.Height, v.LeagueId, v.Data.Account)
		this.store.BatchPut(key, []byte{byte(v.Data.Status)})
	}
	return nil
}
func (this *MemberStore) getKey(height uint64, leagueId, account common.Address) []byte {
	buff := make([]byte, 1+common.AddrLength*2+8)
	buff[0] = byte(mcom.ST_MEMBER)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	copy(buff[common.AddrLength+1:common.AddrLength*2+1], account[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength*2+1:], height)
	return buff
}
