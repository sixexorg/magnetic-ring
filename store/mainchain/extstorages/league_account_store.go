package extstorages

import (
	"bytes"

	"math/big"

	"encoding/binary"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

//Redundancy of the leagues data
type ExternalLeague struct {
	enableCache bool
	dbDir       string
	store       *db.LevelDBStore
}

func NewExternalLeague(dbDir string, enableCache bool) (*ExternalLeague, error) {
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	el := new(ExternalLeague)

	el.enableCache = enableCache
	el.dbDir = dbDir
	el.store = store
	return el, nil
}

//NewBatch start a commit batch
func (this *ExternalLeague) NewBatch() {
	this.store.NewBatch()
}
func (this *ExternalLeague) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *ExternalLeague) GetAccountByHeight(account, leagueId common.Address, height uint64) (*extstates.Account, error) {
	key := this.getKey(leagueId, account, height)
	//fmt.Println("🛢 🛢 🛢 🛢 🛢 🛢 get ", hex.EncodeToString(key))
	val, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	acc, err := this.deserializeAccount(val)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
func (this *ExternalLeague) GetPrev(height uint64, account, leagueId common.Address) (storelaw.AccountStater, error) {
	key := this.getKey(leagueId, account, height)
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
		return &extstates.LeagueAccountState{
			Address:  account,
			LeagueId: leagueId,
			Height:   height,
			Data: &extstates.Account{
				Balance:     big.NewInt(0),
				EnergyBalance: big.NewInt(0),
			},
		}, nil
	}
	as := &extstates.LeagueAccountState{}
	err = as.Deserialize(buff)
	if err != nil {
		return nil, err
	}
	if as.Address.Equals(account) && as.LeagueId.Equals(leagueId) {
		return as, nil
	}
	return &extstates.LeagueAccountState{
		Address:  account,
		LeagueId: leagueId,
		Height:   height,
		Data: &extstates.Account{
			Balance:     big.NewInt(0),
			EnergyBalance: big.NewInt(0),
		},
	}, nil
}
func (this *ExternalLeague) GetAccountRange(start, end uint64, account, leagueId common.Address) (storelaw.AccountStaters, error) {
	s := this.getKey(leagueId, account, start)
	l := this.getKey(leagueId, account, end)
	iter := this.store.NewSeniorIterator(&util.Range{Start: s, Limit: l})
	ass := storelaw.AccountStaters{}
	bf := bytes.NewBuffer(nil)
	var err error
	for iter.Next() {
		bf.Write(iter.Key())
		bf.Write(iter.Value())
		as := &extstates.LeagueAccountState{}
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
func (this *ExternalLeague) Save(state storelaw.AccountStater) error {
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
func (this *ExternalLeague) BatchSave(states storelaw.AccountStaters) error {
	for _, v := range states {
		key := v.GetKey()
		//fmt.Println("🛢 🛢 🛢 🛢 🛢 🛢 save ", hex.EncodeToString(key))
		buff := new(bytes.Buffer)
		err := v.Serialize(buff)
		if err != nil {
			return err
		}
		this.store.BatchPut(key, buff.Bytes())
	}
	return nil
}

func (this *ExternalLeague) getKey(leagueId, account common.Address, height uint64) []byte {
	buff := make([]byte, 1+2*common.AddrLength+8)
	buff[0] = byte(mcom.EXT_LEAGUE_ACCOUNT)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	copy(buff[common.AddrLength+1:common.AddrLength*2+1], account[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength*2+1:], height)
	return buff
}

func (this *ExternalLeague) deserializeAccount(val []byte) (*extstates.Account, error) {
	source := sink.NewZeroCopySource(val)
	account := &extstates.Account{}
	eof := false
	var err error
	account.Nonce, eof = source.NextUint64()
	bal, eof := source.NextComplex()
	account.Balance, err = bal.ComplexToBigInt()
	if err != nil {
		return nil, err
	}
	bbl, eof := source.NextComplex()
	account.EnergyBalance, err = bbl.ComplexToBigInt()
	if err != nil {
		return nil, err
	}
	account.BonusHeight, eof = source.NextUint64()
	if eof {
		return nil, errors.ERR_TXRAW_EOF
	}
	return account, nil
}
