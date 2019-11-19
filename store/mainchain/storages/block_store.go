package storages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sixexorg/magnetic-ring/common/sink"

	"encoding/hex"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/node"
	"github.com/sixexorg/magnetic-ring/store/db"
	scom "github.com/sixexorg/magnetic-ring/store/mainchain/common"
)

type BlockStore struct {
	enableCache bool
	dbDir       string
	cache       *BlockCache
	store       *db.LevelDBStore
}

func NewBlockStore(dbDir string, enableCache bool) (*BlockStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	blockStore := &BlockStore{
		dbDir: dbDir,
		store: store,
	}
	return blockStore, nil
}

//NewBatch start a commit batch
func (this *BlockStore) NewBatch() {
	this.store.NewBatch()
}

//SaveBlock persist block to store
func (this *BlockStore) SaveBlock(block *types.Block) error {
	if this.enableCache {
		this.cache.AddBlock(block)
	}
	blockHeight := block.Header.Height
	err := this.SaveHeader(block) // head+txhash set
	if err != nil {
		return fmt.Errorf("SaveHeader error %s", err)
	}
	pubStrs := make([]string, 0, 3)

	for _, v := range block.Transactions {
		if v.TxType == types.AuthX { // The current star nodes defined by Genesis will generate transactions.
			pubStrs = append(pubStrs, hex.EncodeToString(v.TxData.NodePub))
		}
		err = this.SaveTransaction(v, blockHeight)
		if err != nil {
			return fmt.Errorf("SaveTransaction block height %d tx %s err %s", blockHeight, v.Hash().String(), err)
		}
	}
	node.PushStars(pubStrs)
	this.saveSigData(block.Sigs, blockHeight)
	return nil
}

//ContainBlock return the block specified by block hash save in store
func (this *BlockStore) ContainBlock(blockHash common.Hash) (bool, error) {
	if this.enableCache {
		if this.cache.ContainBlock(blockHash) {
			return true, nil
		}
	}
	key := this.getHeaderKey(blockHash)
	_, err := this.store.Get(key)
	if err != nil {
		if err == errors.ERR_DB_NOT_FOUND {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

//GetBlock return block by block hash
func (this *BlockStore) GetBlock(blockHash common.Hash) (*types.Block, error) {
	var block *types.Block
	if this.enableCache {
		block = this.cache.GetBlock(blockHash)
		if block != nil {
			return block, nil
		}
	}
	header, txHashes, err := this.loadHeaderWithTx(blockHash)
	if err != nil {
		return nil, err
	}
	txList := make([]*types.Transaction, 0, len(txHashes))
	for _, v := range txHashes {
		txHash := v
		tx, _, err := this.GetTransaction(txHash)
		if err != nil {
			return nil, fmt.Errorf("GetTransaction %s error %s", txHash.String(), err)
		}
		if tx == nil {
			return nil, fmt.Errorf("cannot get transaction %s", txHash.String())
		}
		tx.TransactionHash = tx.Hash()
		txList = append(txList, tx)
	}
	sig, err := this.GetSigData(header.Height)
	if err != nil {
		return nil, fmt.Errorf("GetSigData %d error %s", header.Height, err)
	}
	block = &types.Block{
		Header:       header,
		Transactions: txList,
		Sigs:         sig,
	}
	return block, nil
}

//021 62078654

func (this *BlockStore) loadHeaderWithTx(blockHash common.Hash) (*types.Header, []common.Hash, error) {
	key := this.getHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		return nil, nil, err
	}
	reader := bytes.NewBuffer(value)
	header := new(types.Header)
	err = header.Deserialize(reader)
	if err != nil {
		return nil, nil, err
	}
	txSize, err := serialization.ReadUint32(reader)
	if err != nil {
		return nil, nil, err
	}
	txHashes := make([]common.Hash, 0, int(txSize))
	for i := uint32(0); i < txSize; i++ {
		txHash := common.Hash{}
		err = txHash.Deserialize(reader)
		if err != nil {
			return nil, nil, err
		}
		txHashes = append(txHashes, txHash)
	}
	return header, txHashes, nil
}
func (this *BlockStore) RemoveHeaderWithTx(blockHash common.Hash) error {
	key := this.getHeaderKey(blockHash)
	err := this.store.Delete(key)
	return err
}

//SaveHeader persist block header to store
func (this *BlockStore) SaveHeader(block *types.Block) error {
	blockHash := block.Hash()

	key := this.getHeaderKey(blockHash)
	value := bytes.NewBuffer(nil)
	err := block.Header.Serialize(value)
	if err != nil {
		return err
	}
	serialization.WriteUint32(value, uint32(block.Transactions.Len()))
	for _, v := range block.Transactions {
		txHash := v.Hash()
		err := txHash.Serialize(value)
		if err != nil {
			return err
		}
	}
	this.store.BatchPut(key, value.Bytes())
	return nil
}

//GetHeader return the header specified by block hash
func (this *BlockStore) GetHeader(blockHash common.Hash) (*types.Header, error) {
	if this.enableCache {
		block := this.cache.GetBlock(blockHash)
		if block != nil {
			return block.Header, nil
		}
	}
	return this.loadHeader(blockHash)
}

func (this *BlockStore) loadHeader(blockHash common.Hash) (*types.Header, error) {
	key := this.getHeaderKey(blockHash)
	value, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewBuffer(value)
	header := new(types.Header)
	err = header.Deserialize(reader)
	if err != nil {
		return nil, err
	}
	return header, nil
}

//GetCurrentBlock return the current block hash and current block height
func (this *BlockStore) GetCurrentBlock() (common.Hash, uint64, error) {
	key := this.getCurrentBlockKey()
	data, err := this.store.Get(key)
	if err != nil {
		return common.Hash{}, 0, err
	}
	reader := bytes.NewReader(data)
	blockHash := common.Hash{}
	err = blockHash.Deserialize(reader)
	if err != nil {
		return common.Hash{}, 0, err
	}
	height, err := serialization.ReadUint64(reader)
	if err != nil {
		return common.Hash{}, 0, err
	}
	return blockHash, height, nil
}

//SaveCurrentBlock persist the current block height and current block hash to store
func (this *BlockStore) SaveCurrentBlock(height uint64, blockHash common.Hash) error {
	key := this.getCurrentBlockKey()
	value := bytes.NewBuffer(nil)
	blockHash.Serialize(value)
	serialization.WriteUint64(value, height)
	this.store.BatchPut(key, value.Bytes())
	return nil
}

//GetHeaderIndexList return the head index store in header index list
func (this *BlockStore) GetHeaderIndexList() (map[uint64]common.Hash, error) {
	result := make(map[uint64]common.Hash)
	iter := this.store.NewIterator([]byte{byte(scom.IX_HEADER_HASH_LIST)})
	defer iter.Release()
	for iter.Next() {
		startCount, err := this.getStartHeightByHeaderIndexKey(iter.Key())
		if err != nil {
			return nil, fmt.Errorf("getStartHeightByHeaderIndexKey error %s", err)
		}
		reader := bytes.NewReader(iter.Value())
		count, err := serialization.ReadUint32(reader)
		if err != nil {
			return nil, fmt.Errorf("serialization.ReadUint32 count error %s", err)
		}
		for i := uint32(0); i < count; i++ {
			height := startCount + uint64(i)
			blockHash := common.Hash{}
			err = blockHash.Deserialize(reader)
			if err != nil {
				return nil, fmt.Errorf("blockHash.Deserialize error %s", err)
			}
			result[height] = blockHash
		}
	}
	return result, nil
}

//SaveHeaderIndexList persist header index list to store
//func (this *BlockStore) SaveHeaderIndexList(startIndex uint64, indexList []common.Hash) error {
//	indexKey := this.getHeaderIndexListKey(startIndex)
//	indexSize := uint32(len(indexList))
//	value := bytes.NewBuffer(nil)
//	serialization.WriteUint32(value, indexSize)
//	for _, hash := range indexList {
//		hash.Serialize(value)
//	}
//	this.store.BatchPut(indexKey, value.Bytes())
//	return nil
//}

//GetBlockHash return block hash by block height
func (this *BlockStore) GetBlockHash(height uint64) (common.Hash, error) {
	key := this.getBlockHashKey(height)
	value, err := this.store.Get(key)
	if err != nil {
		return common.Hash{}, err
	}
	blockHash, err := common.ParseHashFromBytes(value)
	if err != nil {
		return common.Hash{}, err
	}
	return blockHash, nil
}

//SaveBlockHash persist block height and block hash to store
func (this *BlockStore) SaveBlockHash(height uint64, blockHash common.Hash) {
	key := this.getBlockHashKey(height)
	this.store.BatchPut(key, blockHash.ToBytes())
}

//SaveTransaction persist transaction to store
func (this *BlockStore) SaveTransaction(tx *types.Transaction, height uint64) error {
	if this.enableCache {
		this.cache.AddTransaction(tx, height)
	}
	return this.putTransaction(tx, height)
}

func (this *BlockStore) saveSigData(sig *types.SigData, height uint64) error {
	if this.enableCache {
		this.cache.AddSigData(sig, height)
	}
	return this.putSigData(sig, height)
}
func (this *BlockStore) putSigData(sig *types.SigData, height uint64) error {
	key := this.getSigDataKey(height)
	buff := bytes.NewBuffer(nil)
	err := sig.Serialize(buff)
	if err != nil {
		return err
	}
	this.store.BatchPut(key, buff.Bytes())
	return nil
}

func (this *BlockStore) putTransaction(tx *types.Transaction, height uint64) error {
	if tx.TxType == types.RaiseUT {
		fmt.Println("🔆 league vote save on mainchain")
	}
	txHash := tx.Hash()
	key := this.getTransactionKey(txHash)
	value := bytes.NewBuffer(nil)
	serialization.WriteUint64(value, height)
	err := tx.Serialize(value)
	if err != nil {
		return err
	}
	this.store.BatchPut(key, value.Bytes())
	return nil
}
func (this *BlockStore) GetSigData(height uint64) (*types.SigData, error) {
	if this.enableCache {
		sig := this.cache.GetSigData(height)
		if sig != nil {
			return sig, nil
		}
	}
	return this.loadSigData(height)
}

//GetTransaction return transaction by transaction hash
func (this *BlockStore) GetTransaction(txHash common.Hash) (*types.Transaction, uint64, error) {
	if this.enableCache {
		tx, height := this.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}
	return this.loadTransaction(txHash)
}
func (this *BlockStore) loadSigData(height uint64) (*types.SigData, error) {
	key := this.getSigDataKey(height)
	value, err := this.store.Get(key)
	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer(value)
	sig := &types.SigData{}
	if err = sig.Deserialize(buff); err != nil {
		return nil, err
	}
	return sig, nil
	/*
		source := sink.NewZeroCopySource(value)
		sig := &types.SigData{}
		if err = sig.Deserialization(source); err != nil {
			return nil, err
		}
		return sig, nil*/
}
func (this *BlockStore) loadTransaction(txHash common.Hash) (*types.Transaction, uint64, error) {
	key := this.getTransactionKey(txHash)
	var tx *types.Transaction
	var height uint64
	if this.enableCache {
		tx, height = this.cache.GetTransaction(txHash)
		if tx != nil {
			return tx, height, nil
		}
	}
	value, err := this.store.Get(key)
	if err != nil {
		return nil, 0, err
	}

	source := sink.NewZeroCopySource(value)
	var eof bool
	height, eof = source.NextUint64()
	if eof {
		return nil, 0, io.ErrUnexpectedEOF
	}
	tx = new(types.Transaction)
	err = tx.Deserialization(source)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction deserialize error %s", err)
	}
	return tx, height, nil
}

//IsContainTransaction return whether the transaction is in store
func (this *BlockStore) ContainTransaction(txHash common.Hash) (bool, error) {
	key := this.getTransactionKey(txHash)

	if this.enableCache {
		if this.cache.ContainTransaction(txHash) {
			return true, nil
		}
	}
	_, err := this.store.Get(key)
	if err != nil {
		if err == errors.ERR_DB_NOT_FOUND {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

//GetVersion return the version of store
func (this *BlockStore) GetVersion() (byte, error) {
	key := this.getVersionKey()
	value, err := this.store.Get(key)
	if err != nil {
		return 0, err
	}
	reader := bytes.NewReader(value)
	return reader.ReadByte()
}

//SaveVersion persist version to store
func (this *BlockStore) SaveVersion(ver byte) error {
	key := this.getVersionKey()
	return this.store.Put(key, []byte{ver})
}

//ClearAll clear all the data of block store
func (this *BlockStore) ClearAll() error {
	this.NewBatch()
	iter := this.store.NewIterator(nil)
	for iter.Next() {
		this.store.BatchDelete(iter.Key())
	}
	iter.Release()
	return this.CommitTo()
}

//CommitTo commit the batch to store
func (this *BlockStore) CommitTo() error {
	return this.store.BatchCommit()
}

//Close block store
func (this *BlockStore) Close() error {
	return this.store.Close()
}

func (this *BlockStore) getTransactionKey(txHash common.Hash) []byte {
	key := bytes.NewBuffer(nil)
	key.WriteByte(byte(scom.DATA_TRANSACTION))
	txHash.Serialize(key)
	return key.Bytes()
}

func (this *BlockStore) getHeaderKey(blockHash common.Hash) []byte {
	data := blockHash.ToBytes()
	key := make([]byte, 1+len(data))
	key[0] = byte(scom.DATA_HEADER)
	copy(key[1:], data)
	return key
}

func (this *BlockStore) getBlockHashKey(height uint64) []byte {
	key := make([]byte, 9, 9)
	key[0] = byte(scom.DATA_BLOCK)
	binary.LittleEndian.PutUint64(key[1:], height)
	return key
}

func (this *BlockStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (this *BlockStore) getBlockMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_BLOCK_MERKLE_TREE)}
}

func (this *BlockStore) getVersionKey() []byte {
	return []byte{byte(scom.SYS_VERSION)}
}

//func (this *BlockStore) getHeaderIndexListKey(startHeight uint64) []byte {
//	key := bytes.NewBuffer(nil)
//	key.WriteByte(byte(scom.IX_HEADER_HASH_LIST))
//	serialization.WriteUint64(key, startHeight)
//	return key.Bytes()
//}

func (this *BlockStore) getStartHeightByHeaderIndexKey(key []byte) (uint64, error) {
	reader := bytes.NewReader(key[1:])
	height, err := serialization.ReadUint64(reader)
	if err != nil {
		return 0, err
	}
	return height, nil
}
func (this *BlockStore) getSigDataKey(height uint64) []byte {
	key := make([]byte, 9)
	key[0] = byte(scom.DATA_SIGDATA)
	binary.LittleEndian.PutUint64(key[1:], height)
	return key
}
