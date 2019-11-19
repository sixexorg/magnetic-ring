package storages

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

const (
	BLOCK_CAHE_SIZE        = 10    //Block cache size
	TRANSACTION_CACHE_SIZE = 10000 //Transaction cache size
)

//Value of transaction cache
type TransactionCacheaValue struct {
	Tx     *types.Transaction
	Height uint64
}

//BlockCache with block cache and transaction hash
type BlockCache struct {
	blockCache       *lru.ARCCache
	transactionCache *lru.ARCCache
	sigCache         *lru.ARCCache
}

//NewBlockCache return BlockCache instance
func NewBlockCache() (*BlockCache, error) {
	blockCache, err := lru.NewARC(BLOCK_CAHE_SIZE)
	if err != nil {
		return nil, fmt.Errorf("NewARC block error %s", err)
	}
	transactionCache, err := lru.NewARC(TRANSACTION_CACHE_SIZE)
	if err != nil {
		return nil, fmt.Errorf("NewARC header error %s", err)
	}
	sigCache, err := lru.NewARC(BLOCK_CAHE_SIZE)
	if err != nil {
		return nil, fmt.Errorf("NewARC sig error %s", err)
	}
	return &BlockCache{
		blockCache:       blockCache,
		transactionCache: transactionCache,
		sigCache:         sigCache,
	}, nil
}

//AddBlock to cache
func (this *BlockCache) AddBlock(block *types.Block) {
	blockHash := block.Hash()
	this.blockCache.Add(string(blockHash.ToBytes()), block)
}

//GetBlock return block by block hash from cache
func (this *BlockCache) GetBlock(blockHash common.Hash) *types.Block {
	block, ok := this.blockCache.Get(string(blockHash.ToBytes()))
	if !ok {
		return nil
	}
	return block.(*types.Block)
}

//ContainBlock retuen whether block is in cache
func (this *BlockCache) ContainBlock(blockHash common.Hash) bool {
	return this.blockCache.Contains(string(blockHash.ToBytes()))
}

//AddTransaction add transaction to block cache
func (this *BlockCache) AddTransaction(tx *types.Transaction, height uint64) {
	txHash := tx.Hash()
	this.transactionCache.Add(string(txHash.ToBytes()), &TransactionCacheaValue{
		Tx:     tx,
		Height: height,
	})
}
func (this *BlockCache) AddSigData(sig *types.SigData, height uint64) {
	this.sigCache.Add(string(height), sig)
}
func (this *BlockCache) GetSigData(height uint64) (sig *types.SigData) {
	value, ok := this.sigCache.Get(string(height))
	if !ok {
		return nil
	}
	return value.(*types.SigData)
}

//GetTransaction return transaction by transaction hash from cache
func (this *BlockCache) GetTransaction(txHash common.Hash) (*types.Transaction, uint64) {
	value, ok := this.transactionCache.Get(string(txHash.ToBytes()))
	if !ok {
		return nil, 0
	}
	txValue := value.(*TransactionCacheaValue)
	return txValue.Tx, txValue.Height
}

//ContainTransaction return whether transaction is in cache
func (this *BlockCache) ContainTransaction(txHash common.Hash) bool {
	return this.transactionCache.Contains(string(txHash.ToBytes()))
}
