package storages

import (
	"bytes"

	"encoding/binary"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

//prototype pattern
type LeagueStore struct {
	enableCache bool
	dbDir       string
	store       *db.LevelDBStore
}

func NewLeagueStore(dbDir string) (*LeagueStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	leagueStates := &LeagueStore{
		dbDir: dbDir,
		store: store,
	}
	return leagueStates, nil
}

//NewBatch start a commit batch
func (this *LeagueStore) NewBatch() {
	this.store.NewBatch()
}
func (this *LeagueStore) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *LeagueStore) GetPrev(height uint64, leagueId common.Address) (*states.LeagueState, error) {
	key := this.getKey(height, leagueId)
	//fmt.Println("🈚️ 🈚️ 🈚️ get", hex.EncodeToString(key))
	iter := this.store.NewIterator(nil)
	buff := bytes.NewBuffer(nil)
	flag := true
	if iter.Seek(key) {
		//fmt.Println("🈚️ 🈚️ 🈚️ get 0.5")
		if bytes.Compare(iter.Key(), key) == 0 {
			//fmt.Println("🈚️ 🈚️ 🈚️ get 0.6")
			buff.Write(iter.Key())
			buff.Write(iter.Value())
			flag = false
		}
	}
	if flag && iter.Prev() {
		//fmt.Println("🈚️ 🈚️ 🈚️ get 0.7")
		buff.Write(iter.Key())
		buff.Write(iter.Value())
	}
	iter.Release()
	err := iter.Error()
	//fmt.Println("🈚️ 🈚️ 🈚️ get1")
	if err != nil {
		return nil, err
	}
	//fmt.Println("🈚️ 🈚️ 🈚️ get2")
	if buff.Len() == 0 {
		return nil, errors.ERR_DB_NOT_FOUND
	}
	//fmt.Println("🈚️ 🈚️ 🈚️ get3")
	as := &states.LeagueState{}
	err = as.Deserialize(buff)
	//fmt.Println("🈚️ 🈚️ 🈚️ get4")
	if err != nil {
		return nil, err
	}
	//fmt.Println("🈚️ 🈚️ 🈚️ get5")
	if as.Address.Equals(leagueId) {
		rate, creator, minBox, symbol, private, err := this.GetMetaData(leagueId)
		if err != nil {
			return nil, err
		}
		as.Rate = rate
		as.MinBox = minBox
		as.Creator = creator
		as.Symbol = symbol
		as.Private = private
		return as, nil
	}
	//fmt.Println("🈚️ 🈚️ 🈚️ get6")
	return nil, errors.ERR_DB_NOT_FOUND
}
func (this *LeagueStore) GetMetaData(leagueId common.Address) (rate uint32, creator common.Address, minBox uint64, symbol common.Symbol, private bool, err error) {
	//fmt.Println("🈚️ 🈚️ 🈚️ GetMetaData 1")
	k := this.getMetaKey(leagueId)
	v, err := this.store.Get(k)
	if err != nil {
		return
	}
	//fmt.Println("🈚️ 🈚️ 🈚️ GetMetaData 2")
	sk := sink.NewZeroCopySource(v)
	eof := false
	rate, eof = sk.NextUint32()
	creator, eof = sk.NextAddress()
	minBox, eof = sk.NextUint64()
	symbol, eof = sk.NextSymbol()
	private, _, eof = sk.NextBool()
	if eof {
		err = errors.ERR_SINK_EOF
	}
	//fmt.Println("🈚️ 🈚️ 🈚️ GetMetaData 3")
	return
}

//SaveBlock persist block to store
func (this *LeagueStore) Save(state *states.LeagueState) error {
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
func (this *LeagueStore) SymbolExists(symbol common.Symbol) error {
	bl, err := this.store.Has(this.getSymbolKey(symbol))
	if err != nil || bl {
		return errors.ERR_STATE_SYMBOL_EXISTS
	}
	return nil
}
func (this *LeagueStore) BatchSave(states states.LeagueStates) error {
	for _, v := range states {
		key := v.GetKey()
		//fmt.Println("🈚️ 🈚️ 🈚️ save league ", hex.EncodeToString(key), v.Address.ToString(), v.Data.FrozenBox.Uint64(), v.Rate, v.Height)
		buff := new(bytes.Buffer)
		err := v.Serialize(buff)
		if err != nil {
			return err
		}
		this.store.BatchPut(key, buff.Bytes())
		if v.Data.Nonce == 1 {
			mk := this.getMetaKey(v.Address)
			if bl, _ := this.store.Has(mk); !bl {
				mv := this.getMetaVal(v.Rate, v.Creator, v.MinBox, v.Symbol, v.Private)
				this.store.BatchPut(mk, mv)
				//fmt.Println("🈚️ 🈚️ 🈚️ save first ", hex.EncodeToString(key), v.Address.ToString(), v.Rate)
				sym := this.getSymbolKey(v.Symbol)
				this.store.BatchPut(sym, v.Address[:])
			}
		}
	}
	return nil
}
func (this *LeagueStore) getKey(height uint64, leagueId common.Address) []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(mcom.ST_LEAGUE)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], height)
	return buff
}

func (this *LeagueStore) getMetaKey(leagueId common.Address) []byte {
	buff := make([]byte, 1+common.AddrLength)
	buff[0] = byte(mcom.ST_LEAGUE_META)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	return buff
}

func (this *LeagueStore) getMetaVal(rate uint32, creator common.Address, minBox uint64, symbol common.Symbol, private bool) []byte {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteUint32(rate)
	sk.WriteAddress(creator)
	sk.WriteUint64(minBox)
	sk.WriteSymbol(symbol)
	sk.WriteBool(private)
	return sk.Bytes()
}

func (this *LeagueStore) getSymbolKey(symbol common.Symbol) []byte {
	buff := make([]byte, 1+common.SymbolLen)
	buff[0] = byte(mcom.ST_SYMBOL)
	copy(buff[1:], symbol[:])
	return buff
}
