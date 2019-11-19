package storages

import (
	"bytes"
	"encoding/binary"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/db"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

// write to db by height
type VoteStore struct {
	dbDir        string
	store        *db.LevelDBStore
	votePrefix   byte
	recordPrefix byte
}

func NewVoteStore(dbDir string, votePrefix, recordPrefix byte) (*VoteStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	voteStore := &VoteStore{
		dbDir:        dbDir,
		store:        store,
		votePrefix:   votePrefix,
		recordPrefix: recordPrefix,
	}
	return voteStore, nil
}

func (this *VoteStore) NewBatch() {
	this.store.NewBatch()
}
func (this *VoteStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *VoteStore) GetVoteState(voteId common.Hash, height uint64) (*storelaw.VoteState, error) {
	fmt.Printf("⭕️ vote store get voteId:%s height:%d\n", voteId.String(), height)
	key := this.getKey(voteId, height)
	iter := this.store.NewIterator(nil)
	flag := true
	buff := bytes.NewBuffer(nil)
	if iter.Seek(key) {
		if bytes.Compare(iter.Key(), key) == 0 {
			buff.Write(iter.Key()[1:])
			buff.Write(iter.Value())
			flag = false
		}
	}
	if flag && iter.Prev() {
		buff.Write(iter.Key()[1:])
		buff.Write(iter.Value())
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}
	vs := &storelaw.VoteState{}
	err = vs.Deserialize(buff)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(vs.VoteId[:], voteId[:]) {
		return vs, nil
	}
	return nil, errors.ERR_DB_NOT_FOUND
}

func (this *VoteStore) SaveVotes(vs []*storelaw.VoteState) {
	for _, v := range vs {
		fmt.Printf("⭕️ vote store save voteId:%s height:%d\n", v.VoteId.String(), v.Height)
		key := this.getKey(v.VoteId, v.Height)
		val := v.GetVal()
		this.store.BatchPut(key, val)
	}
}

func (this *VoteStore) AlreadyVoted(voteId common.Hash, account common.Address) bool {
	key := this.getAccountKey(voteId, account)
	_, err := this.store.Get(key)
	if err != nil {
		return false
	}
	return true
}
func (this *VoteStore) SaveAccountVoted(avs []*storelaw.AccountVoted, height uint64) {
	for _, v := range avs {
		for _, vi := range v.Accounts {
			key := this.getAccountKey(v.VoteId, vi)

			val := make([]byte, 8)
			binary.LittleEndian.PutUint64(val, height)
			this.store.BatchPut(key, val)
		}
	}
}

func (this *VoteStore) getKey(voteId common.Hash, height uint64) []byte {
	buff := make([]byte, 1+common.HashLength+8)
	buff[0] = this.votePrefix
	copy(buff[1:common.HashLength+1], voteId[:])
	binary.LittleEndian.PutUint64(buff[common.HashLength+1:], height)
	return buff
}
func (this *VoteStore) getAccountKey(voteId common.Hash, account common.Address) []byte {
	buff := make([]byte, 1+common.HashLength+common.AddrLength)
	buff[0] = this.recordPrefix
	copy(buff[1:common.HashLength+1], voteId[:])
	copy(buff[common.HashLength+1:], account[:])
	return buff
}
