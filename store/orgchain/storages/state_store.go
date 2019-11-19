package storages
import (
	"fmt"
	
	"bytes"
	
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/serialization"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/merkle"
	"github.com/sixexorg/magnetic-ring/store/db"
	scom "github.com/sixexorg/magnetic-ring/store/orgchain/common"
)

//StateStore saving the data of ledger states. Like balance of account, and the execution result of smart contract
type StateStore struct {
	dbDir           string                    //Store file path
	store           *db.LevelDBStore          //Store handler
	merklePath      string                    //Merkle tree store path
	merkleTree      *merkle.CompactMerkleTree //Merkle tree of block root
	merkleHashStore merkle.HashStore
}

//NewStateStore return state store instance
func NewStateStore(dbDir, merklePath string) (*StateStore, error) {
	var err error
	store, err := db.NewLevelDBStore(dbDir)
	if err != nil {
		return nil, err
	}
	stateStore := &StateStore{
		dbDir:      dbDir,
		store:      store,
		merklePath: merklePath,
	}
	_, height, err := stateStore.GetCurrentBlock()
	if err != nil && err != errors.ERR_DB_NOT_FOUND {
		return nil, fmt.Errorf("GetCurrentBlock error %s", err)
	}
	err = stateStore.init(height)
	if err != nil {
		return nil, fmt.Errorf("init error %s", err)
	}
	return stateStore, nil
}

//NewBatch start new commit batch
func (self *StateStore) NewBatch() {
	self.store.NewBatch()
}

func (self *StateStore) init(currBlockHeight uint64) error {
	treeSize, hashes, err := self.GetMerkleTree()
	if err != nil && err != errors.ERR_DB_NOT_FOUND {
		return err
	}
	if treeSize > 0 && treeSize != currBlockHeight+1 {
		return fmt.Errorf("merkle tree size is inconsistent with blockheight: %d", currBlockHeight+1)
	}
	self.merkleHashStore, err = merkle.NewFileHashStore(self.merklePath, treeSize)
	if err != nil {
		return fmt.Errorf("merkle store is inconsistent with ChainStore. persistence will be disabled")
	}
	self.merkleTree = merkle.NewTree(treeSize, hashes, self.merkleHashStore)
	return nil
}

//GetMerkleTree return merkle tree size an tree node
func (self *StateStore) GetMerkleTree() (uint64, []common.Hash, error) {
	key := self.getMerkleTreeKey()
	data, err := self.store.Get(key)
	if err != nil {
		return 0, nil, err
	}
	value := bytes.NewBuffer(data)
	treeSize, err := serialization.ReadUint64(value)
	if err != nil {
		return 0, nil, err
	}
	hashCount := (len(data) - 8) / common.HashLength
	hashes := make([]common.Hash, 0, hashCount)
	for i := 0; i < hashCount; i++ {
		var hash = new(common.Hash)
		err = hash.Deserialize(value)
		if err != nil {
			return 0, nil, err
		}
		hashes = append(hashes, *hash)
	}
	return treeSize, hashes, nil
}

//AddMerkleTreeRoot add a new tree root
func (self *StateStore) AddMerkleTreeRoot(txRoot common.Hash) error {
	key := self.getMerkleTreeKey()
	self.merkleTree.AppendHash(txRoot)
	err := self.merkleHashStore.Flush()
	if err != nil {
		return err
	}
	treeSize := self.merkleTree.TreeSize()
	hashes := self.merkleTree.Hashes()
	value := bytes.NewBuffer(make([]byte, 0, 8+len(hashes)*common.HashLength))
	err = serialization.WriteUint64(value, treeSize)
	if err != nil {
		return err
	}
	for _, hash := range hashes {
		err = hash.Serialize(value)
		if err != nil {
			return err
		}
	}
	self.store.BatchPut(key, value.Bytes())
	return nil
}

//GetMerkleProof return merkle proof of block
func (self *StateStore) GetMerkleProof(proofHeight, rootHeight uint64) ([]common.Hash, error) {
	return self.merkleTree.InclusionProof(proofHeight, rootHeight+1)
}

//CommitTo commit state batch to state store
func (self *StateStore) CommitTo() error {
	return self.store.BatchCommit()
}

//GetCurrentBlock return current block height and current hash in state store
func (self *StateStore) GetCurrentBlock() (common.Hash, uint64, error) {
	key := self.getCurrentBlockKey()
	data, err := self.store.Get(key)
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

//SaveCurrentBlock persist current block to state store
func (self *StateStore) SaveCurrentBlock(height uint64, blockHash common.Hash) error {
	key := self.getCurrentBlockKey()
	value := bytes.NewBuffer(nil)
	blockHash.Serialize(value)
	serialization.WriteUint64(value, height)
	self.store.BatchPut(key, value.Bytes())
	return nil
}

func (self *StateStore) getCurrentBlockKey() []byte {
	return []byte{byte(scom.SYS_CURRENT_BLOCK)}
}

func (self *StateStore) GetBlockRootWithNewTxRoot(txRoot common.Hash) common.Hash {
	return self.merkleTree.GetRootWithNewLeaf(txRoot)
}

func (self *StateStore) getMerkleTreeKey() []byte {
	return []byte{byte(scom.SYS_BLOCK_MERKLE_TREE)}
}

//ClearAll clear all data in state store
func (self *StateStore) ClearAll() error {
	self.store.NewBatch()
	iter := self.store.NewIterator(nil)
	for iter.Next() {
		self.store.BatchDelete(iter.Key())
	}
	iter.Release()
	return self.store.BatchCommit()
}

//Close state store
func (self *StateStore) Close() error {
	return self.store.Close()
}
