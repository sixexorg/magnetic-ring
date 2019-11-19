package extstorages

import (
	"encoding/binary"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/db"
	mcom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

//Redundancy of the leagues data
type ExtAccIndexStore struct {
	enableCache bool
	dbDir       string
	store       *db.LevelDBStore
}

func NewExtAccIndexStore(dbDir string, enableCache bool) (*ExtAccIndexStore, error) {
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	el := new(ExtAccIndexStore)
	el.enableCache = enableCache
	el.dbDir = dbDir
	el.store = store
	return el, nil
}

func (this *ExtAccIndexStore) Save(height uint64, leagueId common.Address, addrs []common.Address, hashes common.HashArray) error {
	key := this.getKey(height, leagueId)
	cp, err := sink.AddressesToComplex(addrs)
	if err != nil {
		return err
	}
	cp2, err := sink.HashArrayToComplex(hashes)
	if err != nil {
		return err
	}
	sk := sink.NewZeroCopySink(nil)
	sk.WriteComplex(cp)
	sk.WriteComplex(cp2)
	return this.store.Put(key, sk.Bytes())
}
func (this *ExtAccIndexStore) Get(height uint64, leagueId common.Address) ([]common.Address, common.HashArray, error) {
	key := this.getKey(height, leagueId)
	val, err := this.store.Get(key)
	if err != nil {
		if err == errors.ERR_DB_NOT_FOUND {
			return []common.Address{}, common.HashArray{}, nil
		}
		return nil, nil, err
	}
	sk := sink.NewZeroCopySource(val)
	cp, eof := sk.NextComplex()
	if eof {
		return nil, nil, errors.ERR_SINK_EOF
	}
	cp2, eof := sk.NextComplex()
	if eof {
		return nil, nil, errors.ERR_SINK_EOF
	}
	addrs, err := cp.ComplexToTxAddresses()
	if err != nil {
		return nil, nil, err
	}
	hashes, err := cp2.ComplexToHashArray()
	if err != nil {
		return nil, nil, err
	}
	return addrs, hashes, nil
}

func (tshi *ExtAccIndexStore) getKey(height uint64, leagueId common.Address) []byte {
	buff := make([]byte, 1+8+common.AddrLength)
	buff[0] = byte(mcom.EXT_LEAGUE_ACCOUNT_INDEX)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	binary.LittleEndian.PutUint64(buff[1+common.AddrLength:], height)
	return buff
}
