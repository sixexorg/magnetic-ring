package dpoa

import (
	"fmt"
	"sync"

	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/node"
	"github.com/sixexorg/magnetic-ring/radar/mainchain"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/mainchain/validation"
)

type BlockStore struct {
	sync.RWMutex

	db            *storages.LedgerStoreImp
	cacheLen      uint64
	pendingBlocks map[uint64]*comm.Block
	epochEnd      uint64
	//curStars      []string
	//earthNode     string
	accountStr string
	gesisBlk   *comm.Block
	failersMap map[uint64]map[string]struct{}
	//epochNotify     chan struct{}
	mainRadar *mainchain.LeagueConsumers
	p2pActor  *actor.PID
	curHeight uint64
}

func NewBlockStore(db *storages.LedgerStoreImp, acc string, p2pActor *actor.PID) (*BlockStore, error) {
	//db.
	bk, _ := db.GetBlockByHeight(1)
	gensis, _ := comm.InitVbftBlock(bk)
	//fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$@@@@@@@@@@@@@@@@@@@@@@@@@@", gensis.Block.Header.Height, len(gensis.getVrfValue()),hex.EncodeToString(gensis.getVrfValue()))
	return &BlockStore{
		//earthNode:     "04d2db562f13d94fd31d5d500152cac0bfd1692b9fc1185f2fbea712dbd34f7e6c65ce05303ee3a4ce772e0513c75e95a3f3dcc97ea45e22cfebbe3a658de4a493",
		db:            db,
		cacheLen:      500,
		gesisBlk:      gensis,
		accountStr:    acc,
		epochEnd:      1,
		pendingBlocks: make(map[uint64]*comm.Block),
		failersMap:    make(map[uint64]map[string]struct{}),
		mainRadar:     mainchain.GetLeagueConsumersInstance(),
		p2pActor:      p2pActor,
		curHeight:     db.GetCurrentBlockHeight(),
		//curStars: []string{
		//	//"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		//	//"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		//	//"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f",
		//	//"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		//	//"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		//	//"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f",
		//	"04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
		//	"044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6",
		//	"0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f"},
	}, nil
}

func (pool *BlockStore) getGeisisBlock() (*comm.Block, error) {
	pool.RLock()
	defer pool.RUnlock()

	return pool.gesisBlk, nil
}

func (pool *BlockStore) getSealedBlock(blockNum uint64) (*comm.Block, error) {
	pool.RLock()
	defer pool.RUnlock()
	if blk, present := pool.pendingBlocks[blockNum]; present {
		return blk, nil
	}
	block, err := pool.db.GetBlockByHeight(blockNum)
	if err != nil {
		return nil, err
	}
	return comm.InitVbftBlock(block)
}

func (pool *BlockStore) getLatestBlock() (*comm.Block, error) {
	pool.RLock()
	defer pool.RUnlock()
	blockNum := pool.db.GetCurrentBlockHeight()
	if blk, present := pool.pendingBlocks[blockNum]; present {
		return blk, nil
	}
	block, err := pool.db.GetBlockByHeight(blockNum)
	if err != nil {
		return nil, err
	}

	return comm.InitVbftBlock(block)
}

func (pool *BlockStore) getLatestBlockNumber() uint64 {
	pool.RLock()
	defer pool.RUnlock()

	return pool.db.GetCurrentBlockHeight()
}

func (pool *BlockStore) GetBlock(blknum uint64) (*comm.Block, error) {
	pool.RLock()
	defer pool.RUnlock()

	return pool.getSealedBlock(blknum)
}

// 持久化落地
func (pool *BlockStore) setBlockSealed(block *comm.Block) error {
	pool.Lock()
	defer pool.Unlock()

	log.Info("setBlockSealed", "block", block)

	if block == nil {
		return fmt.Errorf("try add nil block")
	}

	if block.GetBlockNum() <= pool.db.GetCurrentBlockHeight() {
		log.Warn("chain store adding chained block(%d, %d)", block.GetBlockNum(), pool.getLatestBlockNumber())
		return nil
	}
	//
	if block.Block.Header == nil {
		panic("nil block header")
	}
	log.Info("func dpoa store setBlockSealed", "blockHeight", block.Block.Header.Height, "txlen", block.Block.Transactions.Len())
	//if err := pool.addSignaturesToBlockLocked(block, forEmpty); err != nil {
	//	return fmt.Errorf("failed to add sig to block: %s", err)
	//}
	//
	//var sealed *comm.Block
	//if !forEmpty {
	//	// remove empty block
	//	sealed = &comm.Block{
	//		Block: block.Block,
	//		Info:  block.Info,
	//	}
	//} else {
	//	// replace with empty block
	//	sealed = &comm.Block{
	//		Block: block.EmptyBlock,
	//		Info:  block.Info,
	//	}
	//}
	log.Info("func dpoa setBlockSealed 01", "blockHeight", block.Block.Header.Height, "txlen", block.Block.Transactions.Len())
	pool.pendingBlocks[block.GetBlockNum()] = block
	//receipts types.Receipts,
	//	accountStates states.AccountStates,
	//	leagueStates states.LeagueStates,
	//	leagueMemmaps map[common.Address]map[common.Address]*states.LeagueMember
	blkNum := pool.db.GetCurrentBlockHeight() + 1
	log.Info("func dpoa setBlockSealed 02", "realBlockHeight", pool.db.GetCurrentBlockHeight())
	//fmt.Println("====================+!@@@@@@@@@@@@@@@@@@", block.Block.Header.Height, block.Block.Header.PrevBlockHash.String(), block.Block.Header.Timestamp)
	for {
		blk, present := pool.pendingBlocks[blkNum]
		log.Info("func dpoa setBlockSealed 03", "isnil", blk == nil, "present", present)
		if blk, present := pool.pendingBlocks[blkNum]; blk != nil && present {
			log.Info("func dpoa setBlockSealed 04", "blockHeight", blk.Block.Header.Height, "txlen", blk.Block.Transactions.Len())
			blkInfo, err := validation.ValidateBlock(blk.Block, pool.db)
			if err != nil {
				log.Error("func dpoa setBlockSealed 05 validateBlock failed", "err", err)
				return err
			}
			if blkInfo.ObjTxs.Len() > 0 {
				fmt.Println("🔯 consensus CheckLeaguesWithBoolResponse start")
				err = pool.mainRadar.CheckLeaguesWithBoolResponse(blkInfo.ObjTxs, time.Second*5)
				if err != nil {
					log.Error("func dpoa setBlockSealed 06 validateBlock failed ", "err", err)
					fmt.Println("🔯 consensus save failed", err, " 🌐 🔐 p2p sync next height unlocked")
					return err
				}
			}
			err = pool.db.SaveAll(blkInfo)
			if err != nil && blkNum > pool.getLatestBlockNumber() {
				return fmt.Errorf("ledger add blk (%d, %d) failed: %s", blkNum, pool.db.GetCurrentBlockHeight(), err)
			}
			if blkNum != pool.db.GetCurrentBlockHeight() {
				log.Error("chain store added chained block (%d, %d): %s",
					blkNum, pool.db.GetCurrentBlockHeight(), err)
			}

			delete(pool.pendingBlocks, blkNum)
			blkNum++
		} else {
			break
		}
	}
	return nil
}

func (pool *BlockStore) onBlockSealed(blockNum uint64) {
	pool.Lock()
	defer pool.Unlock()

	if blockNum-pool.cacheLen > 0 {
		toFree := make([]uint64, 0)
		for blkNum, _ := range pool.pendingBlocks {
			if blkNum < blockNum-pool.cacheLen {
				toFree = append(toFree, blkNum)
			}
		}
		for _, blkNum := range toFree {
			delete(pool.pendingBlocks, blkNum)
		}
	}

	var blkData *comm.Block
	if v, ok := pool.pendingBlocks[blockNum]; ok {
		blkData = v
	} else {
		block, err := pool.db.GetBlockByHeight(blockNum)
		if err != nil {
			return
		}
		blkData, err = comm.InitVbftBlock(block)
		if err != nil {
			log.Error("------------------BlockStore onBlockSealed InitVbftBlock", "err", err)
		}
		pool.pendingBlocks[blockNum] = blkData
		log.Info("func dpoa store onBlockSealed", "blockHeight", blkData.Block.Header.Height, "txlen", blkData.Block.Transactions.Len())
	}
	//fmt.Println("------->>>>>>>>>>20190110onBlockSealed persist block", "time", time.Now().String(), "miner", blkData.GetProposer(), "blocknum", blkData.GetBlockNum(), "hash", blkData.Block.Hash().String(), "txs",len(blkData.Block.Transactions))
	log.Info("------->>>>>>>>>>20190110onBlockSealed persist block", "miner", blkData.GetProposer(), "blocknum", blkData.GetBlockNum(), "hash", blkData.Block.Hash().String())
	if blkData.GetProposer() == pool.earthNode() {
		//log.Info("BlockStore setBlockSealed update epochBegin", "pre", pool.epochEnd, "cur", blockNum)
		pool.epochEnd = blockNum
		//if pool.earthNode != pool.accountStr {
		//	if pool.epochNotify != nil {
		//		select {
		//		case <-pool.epochNotify:
		//		default:
		//			close(pool.epochNotify)
		//		}
		//	}
		//}
	}

}

func (pool *BlockStore) EpochBegin() uint64 {
	pool.RLock()
	defer pool.RUnlock()

	return pool.epochEnd
}

func (pool *BlockStore) GetCurStars() []string {
	return node.GurStars()
}

func (pool *BlockStore) isEarth() bool {
	pool.RLock()
	defer pool.RUnlock()

	return pool.earthNode() == pool.accountStr
}

func (pool *BlockStore) earthNode() string {
	return node.CurEarth()
}

func (pool *BlockStore) computeFairs(preEpbegin, epEnd uint64, curStars []string) []string {
	//fmt.Println("BlockStore computeFairs----------------------------->>>", preEpbegin, epEnd)
	failers := make([]string, 0)
	defer func() {
		//fmt.Println("BlockStore computeFairs----------------------------->>>end", failers)
	}()
	if preEpbegin == 0 {
		return failers
	}
	m, _ := CalcStellar(float64(len(curStars)))
	for i := epEnd - 1; i > preEpbegin; i-- {
		if blk, err := pool.db.GetBlockByHeight(i); err == nil {
			curBlock, _ := comm.InitVbftBlock(blk)
			if i == preEpbegin+1 {
				for j := 0; j < int(curBlock.GetViews()); j++ {
					failers = append(failers, curStars[j*m:(j+1)*m]...)
				}
			} else {
				pre, _ := pool.db.GetBlockByHeight(i - 1)
				preBlock, _ := comm.InitVbftBlock(pre)
				for j := preBlock.GetViews() + 1; j < curBlock.GetViews(); j++ {
					failers = append(failers, curStars[int(j)*m:(int(j)+1)*m]...)
				}
			}
			for _, pubKey := range curStars[int(curBlock.GetViews())*m : int(curBlock.GetViews()+1)*m] {
				//fmt.Println("BlockStore computeFairs1======>>>------>>>", m, pubKey, curBlock.getViews())
				var isFound bool = false
				for _, data := range curBlock.Block.Sigs.ProcSigs {
					p, _ := GetNode(curStars, int(byte2Int(data[0:2])))
					//fmt.Println("BlockStore computeFairs======>>>------>>>", p, byte2Int(data[0:2]))
					if pubKey == p {
						isFound = true
						break
					}
				}
				if !isFound {
					failers = append(failers, pubKey)
				}
			}
		}
	}

	return failers
}

func (pool *BlockStore) preEpheight(epBegin uint64) uint64 {
	for i := epBegin; i > 0; i-- {
		block, _ := pool.db.GetBlockByHeight(i)
		bk, _ := comm.InitVbftBlock(block)
		if bk.GetProposer() == pool.earthNode() {
			return bk.GetBlockNum()
		}
	}
	return 0
}

func (pool *BlockStore) inFailers(epEnd uint64, pub string) bool {
	pool.Lock()
	defer pool.Unlock()
	//fmt.Println("BlockStore inFailers----------------------------->>>", pool.epochEnd, epEnd, pub)

	if _, ok := pool.failersMap[epEnd]; !ok {
		if pool.epochEnd < epEnd {
			return false
		} else if pool.epochEnd >= epEnd {
			curStars := pool.GetCurStars()
			pool.failersMap[epEnd] = make(map[string]struct{})
			for _, pub := range pool.computeFairs(pool.preEpheight(epEnd-1), epEnd, curStars) {
				pool.failersMap[epEnd][pub] = struct{}{}
			}
		}
	}
	/*for k, _ := range pool.failersMap[epEnd] {
		fmt.Println("BlockStore inFailers>>>", epEnd, k)
	}*/
	if _, ok := pool.failersMap[epEnd][pub]; !ok {
		return false
	}
	return true
}

func (pool *BlockStore) close() {
}

func (pool *BlockStore) sealBlock(block *comm.Block) error {
	log.Info("func dpoa store sealBlock", "blockHeight", block.Block.Header.Height, "txlen", block.Block.Transactions.Len())
	sealedBlkNum := block.GetBlockNum()
	if sealedBlkNum < pool.getLatestBlockNumber()+1 {
		// we already in future round
		log.Error("late seal of , current blkNum", "sealedBlkNum", sealedBlkNum, "getLatestBlockNumber", pool.getLatestBlockNumber()+1)
		return nil
	} else if sealedBlkNum > pool.getLatestBlockNumber()+1 {
		// we have lost sync, restarting syncing
		//self.restartSyncing()
		return fmt.Errorf("future seal of %d, current blknum: %d", sealedBlkNum, pool.getLatestBlockNumber()+1)
	}

	if err := pool.setBlockSealed(block); err != nil {
		return fmt.Errorf("failed to seal block: %s", err)
	}


	blk, _ := pool.getSealedBlock(sealedBlkNum)
	h := blk.Block.Hash()
	prevBlkHash := block.GetPrevBlockHash()
	log.Info("func dpoa store sealBlock server sealed block proposer prevhash hash", "accountStr", pool.accountStr,
		"sealedBlkNum", sealedBlkNum, "proposer", block.GetProposer(), "prevhash", prevBlkHash, "hash", h)

	return nil
}

func (pool *BlockStore) nonConsensusNode() bool {
	for _, v := range pool.GetCurStars() {
		if pool.accountStr == v {
			return false
		}
	}
	return true
}
