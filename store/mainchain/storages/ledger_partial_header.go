package storages

import (
	"fmt"
	"sort"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

//GetCurrentHeaderHeight return the current header height.
//In block sync states, Header height is usually higher than block height that is has already committed to storage
func (this *LedgerStoreImp) GetCurrentHeaderHeightbak() uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return 0
	}
	return uint64(size)
}

func (this *LedgerStoreImp) GetCurrentHeaderHeight() uint64 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.currBlockHeight
}

//GetCurrentHeaderHash return the current header hash. The current header means the latest header.
func (this *LedgerStoreImp) GetCurrentHeaderHash() common.Hash {
	this.lock.RLock()
	defer this.lock.RUnlock()
	size := len(this.headerIndex)
	if size == 0 {
		return common.Hash{}
	}
	return this.headerIndex[uint64(size)]
}

func (this *LedgerStoreImp) addHeaderCache(header *types.Header) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerCache[header.Hash()] = header
}

func (this *LedgerStoreImp) delHeaderCache(blockHash common.Hash) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.headerCache, blockHash)
}

func (this *LedgerStoreImp) getHeaderCache(blockHash common.Hash) *types.Header {
	this.lock.RLock()
	defer this.lock.RUnlock()
	header, ok := this.headerCache[blockHash]
	if !ok {
		return nil
	}
	return header
}

//GetHeaderByHash return the block header by block hash
func (this *LedgerStoreImp) GetHeaderByHash(blockHash common.Hash) (*types.Header, error) {
	header := this.getHeaderCache(blockHash)
	if header != nil {
		return header, nil
	}
	return this.blockStore.GetHeader(blockHash)
}
func (this *LedgerStoreImp) setHeaderIndex(height uint64, blockHash common.Hash) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.headerIndex[height] = blockHash
}

func (this *LedgerStoreImp) getHeaderIndex(height uint64) common.Hash {
	this.lock.RLock()
	defer this.lock.RUnlock()
	blockHash, ok := this.headerIndex[height]
	if !ok {
		return common.Hash{}
	}
	return blockHash
}
func (this *LedgerStoreImp) AddHeader(header *types.Header) error {
	nextHeaderHeight := this.GetCurrentHeaderHeight() + 1
	if header.Height != nextHeaderHeight {
		return fmt.Errorf("header height %d not equal next header height %d", header.Height, nextHeaderHeight)
	}
	err := this.verifyHeader(header)
	if err != nil {
		return fmt.Errorf("verifyHeader error %s", err)
	}
	this.addHeaderCache(header)
	this.setHeaderIndex(header.Height, header.Hash())
	return nil
}

func (this *LedgerStoreImp) AddHeaders(headers []*types.Header) error {
	sort.Slice(headers, func(i, j int) bool {
		return headers[i].Height < headers[j].Height
	})
	var err error
	for _, header := range headers {
		err = this.AddHeader(header)
		if err != nil {
			return err
		}
	}
	return nil
}
//func (this *LedgerStoreImp) saveHeaderIndexList() error {
//	this.lock.RLock()
//	storeCount := this.storedIndexCount
//	currHeight := this.currBlockHeight
//	if currHeight-storeCount < HEADER_INDEX_BATCH_SIZE {
//		this.lock.RUnlock()
//		return nil
//	}
//
//	headerList := make([]common.Hash, HEADER_INDEX_BATCH_SIZE)
//	for i := uint64(0); i < HEADER_INDEX_BATCH_SIZE; i++ {
//		height := storeCount + i
//		headerList[i] = this.headerIndex[height]
//	}
//	this.lock.RUnlock()
//	err := this.blockStore.SaveHeaderIndexList(storeCount, headerList)
//	if err != nil {
//		return fmt.Errorf("SaveHeaderIndexList start %d error %s", storeCount, err)
//	}
//
//	this.lock.Lock()
//	this.storedIndexCount += HEADER_INDEX_BATCH_SIZE
//	this.lock.Unlock()
//	return nil
//}
