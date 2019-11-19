package extstorages

import (
	"bytes"
	"encoding/binary"

	"math/big"

	"io"

	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/store/db"
	scom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
)

type ExternalLeagueBlock struct {
	enableCache bool
	dbDir       string
	store       *db.LevelDBStore
}

func NewExternalLeagueBlock(dbDir string, enableCache bool) (*ExternalLeagueBlock, error) {
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	el := new(ExternalLeagueBlock)

	el.enableCache = enableCache
	el.dbDir = dbDir
	el.store = store
	return el, nil
}

//NewBatch start a commit batch
func (this *ExternalLeagueBlock) NewBatch() {
	this.store.NewBatch()
}
func (this *ExternalLeagueBlock) CommitTo() error {
	return this.store.BatchCommit()
}

func (this *ExternalLeagueBlock) GetBlockHash(leagueId common.Address, height uint64) (common.Hash, error) {
	key := this.getBlockHashKey(leagueId, height)
	value, err := this.store.Get(key)
	if err != nil {
		return common.Hash{}, err
	}
	sk := sink.NewZeroCopySource(value)
	blockHash, eof := sk.NextHash()
	if eof {
		return common.Hash{}, io.ErrUnexpectedEOF
	}
	return blockHash, nil
}

func (this *ExternalLeagueBlock) GetBlock(leagueId common.Address, height uint64) (*extstates.LeagueBlockSimple, error) {
	hash, err := this.GetBlockHash(leagueId, height)
	if err != nil {
		return nil, err
	}
	key := this.getHeaderKey(hash)
	val, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	block := &extstates.LeagueBlockSimple{}
	buff := bytes.NewBuffer(val)
	err = block.Deserialize(buff)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (this *ExternalLeagueBlock) GetBlockHashSpan(leagueId common.Address, left, right uint64) (hashArr common.HashArray, gasUsedSum *big.Int, err error) {
	leftKey := this.getBlockHashKey(leagueId, left)
	rightKey := this.getBlockHashKey(leagueId, right+1)
	iter := this.store.NewSeniorIterator(&util.Range{
		Start: leftKey,
		Limit: rightKey,
	})
	hashes := make(common.HashArray, 0, right-left+1)
	gasSum := big.NewInt(0)
	for iter.Next() {
		value := iter.Value()
		sk := sink.NewZeroCopySource(value)
		blockHash, eof := sk.NextHash()
		if eof {
			return nil, nil, io.ErrUnexpectedEOF
		}
		cp, eof := sk.NextComplex()
		if eof {
			return nil, nil, io.ErrUnexpectedEOF
		}
		gasUsed, err := cp.ComplexToBigInt()
		if err != nil {
			return nil, nil, err
		}
		gasSum.Add(gasSum, gasUsed)
		hashes = append(hashes, blockHash)
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return nil, nil, err
	}
	return hashes, gasSum, nil
}

//SaveBlock persist block to store
func (this *ExternalLeagueBlock) Save(block *extstates.LeagueBlockSimple) error {
	blockHash := block.Header.Hash()
	err := this.saveBlockHash(block.Header.LeagueId, block.Header.Height, blockHash, block.EnergyUsed)
	if err != nil {
		return err
	}
	buff := new(bytes.Buffer)
	err = block.Serialize(buff)
	if err != nil {
		return err
	}
	key := this.getHeaderKey(blockHash)
	this.store.BatchPut(key, buff.Bytes())
	return nil
}
func (this *ExternalLeagueBlock) BatchSave(blocks []*extstates.LeagueBlockSimple) error {
	leagueKey := make(map[common.Address]bool)
	for _, v := range blocks {
		leagueKey[v.Header.LeagueId] = true
	}
	var err error
	for _, v := range blocks {
		blockHash := v.Header.Hash()
		err = this.saveBlockHash(v.Header.LeagueId, v.Header.Height, blockHash, v.EnergyUsed)
		if err != nil {
			return err
		}
		buff := new(bytes.Buffer)
		err = v.Serialize(buff)
		if err != nil {
			return err
		}
		key := this.getHeaderKey(blockHash)
		this.store.BatchPut(key, buff.Bytes())
	}
	return nil
}

func (this *ExternalLeagueBlock) saveBlockHash(leagueId common.Address, height uint64, blockHash common.Hash, gasUsed *big.Int) error {
	key := this.getBlockHashKey(leagueId, height)
	sk := sink.NewZeroCopySink(nil)
	sk.WriteBytes(blockHash[:])
	cp, err := sink.BigIntToComplex(gasUsed)
	if err != nil {
		return err
	}
	sk.WriteComplex(cp)
	this.store.BatchPut(key, sk.Bytes())
	return nil
}
func (this *ExternalLeagueBlock) getHeaderKey(blockHash common.Hash) []byte {
	buff := make([]byte, 1+common.HashLength)
	buff[0] = byte(scom.EXT_LEAGUE_DATA_HEADER)
	copy(buff[1:common.HashLength+1], blockHash[:])
	return buff
}
func (this *ExternalLeagueBlock) getBlockHashKey(leagueId common.Address, height uint64) []byte {
	buff := make([]byte, 1+common.AddrLength+8)
	buff[0] = byte(scom.EXT_LEAGUE_DATA_BLOCK)
	copy(buff[1:common.AddrLength+1], leagueId[:])
	binary.LittleEndian.PutUint64(buff[common.AddrLength+1:], height)
	return buff
}
