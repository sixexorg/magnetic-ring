package miansync

import (
	// "fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	ledger "github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/mainchain/validation"

	"fmt"
	"reflect"

	"github.com/sixexorg/magnetic-ring/log"
	p2pComm "github.com/sixexorg/magnetic-ring/p2pserver/common"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"
	"github.com/sixexorg/magnetic-ring/p2pserver/sync/p2pserprotocol"
	"github.com/sixexorg/magnetic-ring/radar/mainchain"
)

const (
	SYNC_MAX_HEADER_FORWARD_SIZE = 500        //keep CurrentHeaderHeight - CurrentBlockHeight <= SYNC_MAX_HEADER_FORWARD_SIZE
	SYNC_MAX_FLIGHT_HEADER_SIZE  = 1          //Number of headers on flight
	SYNC_MAX_FLIGHT_BLOCK_SIZE   = 50         //Number of blocks on flight
	SYNC_MAX_BLOCK_CACHE_SIZE    = 500        //Cache size of block wait to commit to ledger
	SYNC_HEADER_REQUEST_TIMEOUT  = 2          //s, Request header timeout time. If header haven't receive after SYNC_HEADER_REQUEST_TIMEOUT second, retry
	SYNC_BLOCK_REQUEST_TIMEOUT   = 2          //s, Request block timeout time. If block haven't received after SYNC_BLOCK_REQUEST_TIMEOUT second, retry
	SYNC_NEXT_BLOCK_TIMES        = 3          //Request times of next height block
	SYNC_NEXT_BLOCKS_HEIGHT      = 2          //for current block height plus next
	SYNC_NODE_RECORD_SPEED_CNT   = 3          //Record speed count for accuracy
	SYNC_NODE_RECORD_TIME_CNT    = 3          //Record request time  for accuracy
	SYNC_NODE_SPEED_INIT         = 100 * 1024 //Init a big speed (100MB/s) for every node in first round
	SYNC_MAX_ERROR_RESP_TIMES    = 5          //Max error headers/blocks response times, if reaches, delete it
	SYNC_MAX_HEIGHT_OFFSET       = 5          //Offset of the max height and current height
)

//NodeWeight record some params of node, using for sort
type NodeWeight struct {
	id           uint64    //NodeID
	speed        []float32 //Record node request-response speed, using for calc the avg speed, unit kB/s
	timeoutCnt   int       //Node response timeout count
	errorRespCnt int       //Node response error data count
	reqTime      []int64   //Record request time, using for calc the avg req time interval, unit millisecond
}

//NewNodeWeight new a nodeweight
func NewNodeWeight(id uint64) *NodeWeight {
	s := make([]float32, 0, SYNC_NODE_RECORD_SPEED_CNT)
	for i := 0; i < SYNC_NODE_RECORD_SPEED_CNT; i++ {
		s = append(s, float32(SYNC_NODE_SPEED_INIT))
	}
	r := make([]int64, 0, SYNC_NODE_RECORD_TIME_CNT)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	for i := 0; i < SYNC_NODE_RECORD_TIME_CNT; i++ {
		r = append(r, now)
	}
	return &NodeWeight{
		id:           id,
		speed:        s,
		timeoutCnt:   0,
		errorRespCnt: 0,
		reqTime:      r,
	}
}

//AddTimeoutCnt incre timeout count
func (that *NodeWeight) AddTimeoutCnt() {
	that.timeoutCnt++
}

//AddErrorRespCnt incre receive error header/block count
func (that *NodeWeight) AddErrorRespCnt() {
	that.errorRespCnt++
}

//GetErrorRespCnt get the error response count
func (that *NodeWeight) GetErrorRespCnt() int {
	return that.errorRespCnt
}

//AppendNewReqTime append new request time
func (that *NodeWeight) AppendNewReqtime() {
	copy(that.reqTime[0:SYNC_NODE_RECORD_TIME_CNT-1], that.reqTime[1:])
	that.reqTime[SYNC_NODE_RECORD_TIME_CNT-1] = time.Now().UnixNano() / int64(time.Millisecond)
}

//addNewSpeed apend the new speed to tail, remove the oldest one
func (that *NodeWeight) AppendNewSpeed(s float32) {
	copy(that.speed[0:SYNC_NODE_RECORD_SPEED_CNT-1], that.speed[1:])
	that.speed[SYNC_NODE_RECORD_SPEED_CNT-1] = s
}

//Weight calculate node's weight for sort. Highest weight node will be accessed first for next request.
func (that *NodeWeight) Weight() float32 {
	avgSpeed := float32(0.0)
	for _, s := range that.speed {
		avgSpeed += s
	}
	avgSpeed = avgSpeed / float32(len(that.speed))

	avgInterval := float32(0.0)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	for _, t := range that.reqTime {
		avgInterval += float32(now - t)
	}
	avgInterval = avgInterval / float32(len(that.reqTime))
	w := avgSpeed + avgInterval
	return w
}

//NodeWeights implement sorting
type NodeWeights []*NodeWeight

func (nws NodeWeights) Len() int {
	return len(nws)
}

func (nws NodeWeights) Swap(i, j int) {
	nws[i], nws[j] = nws[j], nws[i]
}
func (nws NodeWeights) Less(i, j int) bool {
	ni := nws[i]
	nj := nws[j]
	return ni.Weight() < nj.Weight() && ni.errorRespCnt >= nj.errorRespCnt && ni.timeoutCnt >= nj.timeoutCnt
}

//SyncFlightInfo record the info of fight object(header or block)
type SyncFlightInfo struct {
	Height      uint64         //BlockHeight of HeaderHeight
	nodeId      uint64         //The current node to send msg
	startTime   time.Time      //Request start time
	failedNodes map[uint64]int //Map nodeId => timeout times
	totalFailed int            //Total timeout times
	lock        sync.RWMutex
}

//NewSyncFlightInfo return a new SyncFlightInfo instance
func NewSyncFlightInfo(height uint64, nodeId uint64) *SyncFlightInfo {
	return &SyncFlightInfo{
		Height:      height,
		nodeId:      nodeId,
		startTime:   time.Now(),
		failedNodes: make(map[uint64]int, 0),
	}
}

//GetNodeId return current node id for sending msg
func (that *SyncFlightInfo) GetNodeId() uint64 {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return that.nodeId
}

//SetNodeId set a new node id
func (that *SyncFlightInfo) SetNodeId(nodeId uint64) {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.nodeId = nodeId
}

//MarkFailedNode mark node failed, after request timeout
func (that *SyncFlightInfo) MarkFailedNode() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.failedNodes[that.nodeId] += 1
	that.totalFailed++
}

//GetFailedTimes return failed times of a node
func (that *SyncFlightInfo) GetFailedTimes(nodeId uint64) int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	times, ok := that.failedNodes[nodeId]
	if !ok {
		return 0
	}
	return times
}

//GetTotalFailedTimes return the total failed times of request
func (that *SyncFlightInfo) GetTotalFailedTimes() int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return that.totalFailed
}

//ResetStartTime
func (that *SyncFlightInfo) ResetStartTime() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.startTime = time.Now()
}

//GetStartTime return the start time of request
func (that *SyncFlightInfo) GetStartTime() time.Time {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return that.startTime
}

//BlockInfo is used for saving block information in cache
type BlockInfo struct {
	nodeID uint64
	block  *types.Block
}

//BlockSyncMgr is the manager class to deal with block sync
type BlockSyncMgr struct {
	flightBlocks   map[common.Hash][]*SyncFlightInfo //Map BlockHash => []SyncFlightInfo, using for manager all of those block flights
	flightHeaders  map[uint64]*SyncFlightInfo        //Map HeaderHeight => SyncFlightInfo, using for manager all of those header flights
	blocksCache    map[uint64]*BlockInfo             //Map BlockHash => BlockInfo, using for cache the blocks receive from net, and waiting for commit to ledger
	server         p2pserprotocol.SyncP2PSer         //Pointer to the local node
	syncBlockLock  bool                              //Help to avoid send block sync request duplicate
	syncHeaderLock bool                              //Help to avoid send header sync request duplicate
	saveBlockLock  bool                              //Help to avoid saving block concurrently
	exitCh         chan interface{}                  //ExitCh to receive exit signal
	ledger         *ledger.LedgerStoreImp            //ledger
	radar          *mainchain.LeagueConsumers        //ladar
	lock           sync.RWMutex                      //lock
	nodeWeights    map[uint64]*NodeWeight            //Map NodeID => NodeStatus, using for getNextNode
	syncStatus     bool                              //lock or not sync header info
	lockBlkHeight  uint64                            //should lock block height not save
}

//NewBlockSyncMgr return a BlockSyncMgr instance
func NewBlockSyncMgr(server p2pserprotocol.SyncP2PSer) *BlockSyncMgr {
	return &BlockSyncMgr{
		flightBlocks:  make(map[common.Hash][]*SyncFlightInfo, 0),
		flightHeaders: make(map[uint64]*SyncFlightInfo, 0),
		blocksCache:   make(map[uint64]*BlockInfo, 0),
		server:        server,
		ledger:        server.GetLedger(),
		radar:         mainchain.GetLeagueConsumersInstance(),
		exitCh:        make(chan interface{}, 1),
		nodeWeights:   make(map[uint64]*NodeWeight, 0),
	}
}

//Start to sync
func (that *BlockSyncMgr) Start() {
	go that.sync()
	go that.printHeight()
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-that.exitCh:
			return
		case <-ticker.C:
			go that.checkTimeout()
			go that.sync()
			go that.saveBlock()
		}
	}
}

func (that *BlockSyncMgr) printHeight() {
	t := time.NewTicker(p2pComm.DEFAULT_GEN_BLOCK_TIME * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			hh := that.ledger.GetCurrentHeaderHeight()
			log.Info("main_sync", "CurrentBlockHeight", that.ledger.GetCurrentBlockHeight( /*that.orgID*/), "GetCurrentHeaderHeight", hh)
		case <-that.exitCh:
			return
		}
	}
}

func (that *BlockSyncMgr) checkTimeout() {
	now := time.Now()
	headerTimeoutFlights := make(map[uint64]*SyncFlightInfo, 0)
	blockTimeoutFlights := make(map[common.Hash][]*SyncFlightInfo, 0)
	that.lock.RLock()
	for height, flightInfo := range that.flightHeaders {
		if int(now.Sub(flightInfo.startTime).Seconds()) >= SYNC_HEADER_REQUEST_TIMEOUT {
			headerTimeoutFlights[height] = flightInfo
		}
	}
	for blockHash, flightInfos := range that.flightBlocks {
		for _, flightInfo := range flightInfos {
			if int(now.Sub(flightInfo.startTime).Seconds()) >= SYNC_BLOCK_REQUEST_TIMEOUT {
				blockTimeoutFlights[blockHash] = append(blockTimeoutFlights[blockHash], flightInfo)
			}
		}
	}
	that.lock.RUnlock()

	curHeaderHeight := that.ledger.GetCurrentHeaderHeight()
	curBlockHeight := that.ledger.GetCurrentBlockHeight()

	for height, flightInfo := range headerTimeoutFlights {
		that.addTimeoutCnt(flightInfo.GetNodeId())
		if height <= curHeaderHeight {
			that.delFlightHeader(height)
			continue
		}
		flightInfo.ResetStartTime()
		flightInfo.MarkFailedNode()
		log.Trace("[p2p]checkTimeout main sync headers", "from id", flightInfo.GetNodeId(), "height", height, "after", SYNC_HEADER_REQUEST_TIMEOUT, "Times", flightInfo.GetTotalFailedTimes())
		reqNode := that.getNodeWithMinFailedTimes(flightInfo, curBlockHeight)
		if reqNode == nil {
			break
		}
		flightInfo.SetNodeId(reqNode.GetID())

		headerHash := that.ledger.GetCurrentHeaderHash()
		var nuladd common.Address
		msg := msgpack.NewHeadersReq(headerHash, nuladd, p2pComm.SYNC_DATA_MAIN, curHeaderHeight)
		// msg := msgpack.NewHeadersReq(headerHash,nuladd,p2pComm.SYNC_DATA_MAIN)
		err := that.server.Send(reqNode, msg, false)
		if err != nil {
			log.Warn("[p2p]checkTimeout failed to send a new headersReq", "err", err)
		} else {
			that.appendReqTime(reqNode.GetID())
		}
	}

	for blockHash, flightInfos := range blockTimeoutFlights {
		for _, flightInfo := range flightInfos {
			that.addTimeoutCnt(flightInfo.GetNodeId())
			if flightInfo.Height <= curBlockHeight {
				that.delFlightBlock(blockHash)
				continue
			}
			flightInfo.ResetStartTime()
			flightInfo.MarkFailedNode()
			log.Trace("[p2p]checkTimeout sync", "height", flightInfo.Height, "blockHash", blockHash, "after", SYNC_BLOCK_REQUEST_TIMEOUT, "times", flightInfo.GetTotalFailedTimes())
			reqNode := that.getNodeWithMinFailedTimes(flightInfo, curBlockHeight)
			if reqNode == nil {
				break
			}
			flightInfo.SetNodeId(reqNode.GetID())

			var nuladd common.Address
			msg := msgpack.NewBlkDataReq(blockHash, nuladd, p2pComm.SYNC_DATA_MAIN)
			err := that.server.Send(reqNode, msg, false)
			if err != nil {
				log.Warn("[p2p]checkTimeout reqNode", "ID", reqNode.GetID(), "err", err)
				continue
			} else {
				that.appendReqTime(reqNode.GetID())
			}
		}
	}
}

func (that *BlockSyncMgr) sync() {
	that.syncHeader()
	that.syncBlock()
}

func (that *BlockSyncMgr) syncHeader() {
	if !that.server.ReachMinConnection() {
		return
	}
	if that.tryGetSyncHeaderLock() {
		return
	}
	defer that.releaseSyncHeaderLock()

	if that.getFlightHeaderCount() >= SYNC_MAX_FLIGHT_HEADER_SIZE {
		return
	}
	curBlockHeight := that.ledger.GetCurrentBlockHeight()

	curHeaderHeight := that.ledger.GetCurrentHeaderHeight()
	//Waiting for block catch up header
	if curHeaderHeight-curBlockHeight >= uint64(SYNC_MAX_HEADER_FORWARD_SIZE) {
		return
	}
	NextHeaderId := curHeaderHeight + 1
	reqNode := that.getNextNode(NextHeaderId)
	if reqNode == nil {
		return
	}
	that.addFlightHeader(reqNode.GetID(), NextHeaderId)

	headerHash := that.ledger.GetCurrentHeaderHash()
	// fmt.Println(" ********* stopHeight:",headerHash)
	var nuladd common.Address
	msg := msgpack.NewHeadersReq(headerHash, nuladd, p2pComm.SYNC_DATA_MAIN, curHeaderHeight)
	// msg := msgpack.NewHeadersReq(headerHash,nuladd,p2pComm.SYNC_DATA_MAIN)
	err := that.server.Send(reqNode, msg, false)
	if err != nil {
		log.Warn("[p2p]syncHeader failed to send a new headersReq")
	} else {
		that.appendReqTime(reqNode.GetID())
	}

	log.Info("主链--> Header sync request ", "height", NextHeaderId)
}

func (that *BlockSyncMgr) MainGetBlockHashByHeight(nextBlockHeight uint64) common.Hash {
	nextBlockHeaderHash := that.ledger.GetBlockHeaderHashByHeight(nextBlockHeight)
	empty := common.Hash{}
	if nextBlockHeaderHash != empty {
		return nextBlockHeaderHash
	}
	nextBlockHash, _ := that.ledger.GetBlockHashByHeight(nextBlockHeight)
	return nextBlockHash
}

func (that *BlockSyncMgr) syncBlock() {
	if that.tryGetSyncBlockLock() {
		return
	}
	defer that.releaseSyncBlockLock()

	availCount := SYNC_MAX_FLIGHT_BLOCK_SIZE - that.getFlightBlockCount()
	if availCount <= 0 {
		return
	}
	curBlockHeight := that.ledger.GetCurrentBlockHeight()
	curHeaderHeight := that.ledger.GetCurrentHeaderHeight()
	count := int(curHeaderHeight - curBlockHeight)
	if count <= 0 {
		return
	}
	if count > availCount {
		count = availCount
	}
	cacheCap := SYNC_MAX_BLOCK_CACHE_SIZE - that.getBlockCacheSize()
	if count > cacheCap {
		count = cacheCap
	}

	counter := 1
	i := uint64(0)
	reqTimes := 1
	for {
		if counter > count {
			break
		}
		i++
		nextBlockHeight := curBlockHeight + i
		nextBlockHash := that.MainGetBlockHashByHeight(uint64(nextBlockHeight))
		empty := common.Hash{}
		if nextBlockHash == empty {
			return
		}
		if that.isBlockOnFlight(nextBlockHash) {
			if nextBlockHeight <= curBlockHeight+SYNC_NEXT_BLOCKS_HEIGHT {
				//request more nodes for next block height
				reqTimes = SYNC_NEXT_BLOCK_TIMES
			} else {
				continue
			}
		}
		if that.isInBlockCache(nextBlockHeight) {
			continue
		}
		if nextBlockHeight <= curBlockHeight+SYNC_NEXT_BLOCKS_HEIGHT {
			reqTimes = SYNC_NEXT_BLOCK_TIMES
		}
		for t := 0; t < reqTimes; t++ {
			reqNode := that.getNextNode(nextBlockHeight)
			if reqNode == nil {
				return
			}
			that.addFlightBlock(reqNode.GetID(), nextBlockHeight, nextBlockHash)
			var nuladd common.Address
			msg := msgpack.NewBlkDataReq(nextBlockHash, nuladd, p2pComm.SYNC_DATA_MAIN)
			err := that.server.Send(reqNode, msg, false)
			if err != nil {
				log.Warn("[p2p]syncBlock", "Height", nextBlockHeight, "err", err)
				return
			} else {
				that.appendReqTime(reqNode.GetID())
			}
		}
		counter++
		reqTimes = 1
	}
}

//OnHeaderReceive receive header from net
func (that *BlockSyncMgr) OnHeaderReceive(fromID uint64, headers []*types.Header) {
	if len(headers) == 0 {
		return
	}
	log.Info("Header receive", "Height", headers[0].Height, "Height", headers[len(headers)-1].Height)
	height := headers[0].Height
	curHeaderHeight := that.ledger.GetCurrentHeaderHeight()

	//Means another gorountinue is adding header
	if height <= curHeaderHeight {
		return
	}
	if !that.isHeaderOnFlight(height) {
		return
	}
	err := that.ledger.AddHeaders(headers)
	that.delFlightHeader(height)
	if err != nil {
		that.addErrorRespCnt(fromID)
		n := that.getNodeWeight(fromID)
		if n != nil && n.GetErrorRespCnt() >= SYNC_MAX_ERROR_RESP_TIMES {
			that.delNode(fromID)
		}
		log.Warn("[p2p]OnHeaderReceive AddHeaders", "err", err)
		return
	}
	that.syncHeader()
}

//OnBlockReceive receive block from net
func (that *BlockSyncMgr) OnBlockReceive(fromID uint64, blockSize uint32, block *types.Block) {
	height := uint32(block.Header.Height)
	blockHash := block.Hash()
	log.Trace("[p2p]OnBlockReceive", "height", height)
	/*	fmt.Println("*******************************************[p2p]OnBlockReceive", "height", height)
		for k, v := range block.Transactions {
			fmt.Println("num:", k, ",hash:", v.Hash().String(), ",from:", v.TxData.From.ToString(), ",nonce:", v.TxData.Froms.Tis[0].Nonce)
		}*/
	that.lock.Lock()
	flightInfos := that.flightBlocks[blockHash]
	that.lock.Unlock()

	for _, flightInfo := range flightInfos {
		if flightInfo.GetNodeId() == fromID {
			t := (time.Now().UnixNano() - flightInfo.GetStartTime().UnixNano()) / int64(time.Millisecond)
			s := float32(blockSize) / float32(t) * 1000.0 / 1024.0
			that.addNewSpeed(fromID, s)
			break
		}
	}

	that.delFlightBlock(blockHash)
	curHeaderHeight := uint32(that.ledger.GetCurrentHeaderHeight())
	nextHeader := curHeaderHeight + 1
	if height > nextHeader {
		return
	}
	curBlockHeight := uint32(that.ledger.GetCurrentBlockHeight())
	if height <= curBlockHeight {
		return
	}

	that.addBlockCache(fromID, block)
	go that.saveBlock()
	that.syncBlock()
}

//OnAddNode to node list when a new node added
func (that *BlockSyncMgr) OnAddNode(nodeId uint64) {
	log.Debug("[p2p]OnAddNode", "nodeId", nodeId)
	that.lock.Lock()
	defer that.lock.Unlock()
	w := NewNodeWeight(nodeId)
	that.nodeWeights[nodeId] = w
}

//OnDelNode remove from node list. When the node disconnect
func (that *BlockSyncMgr) OnDelNode(nodeId uint64) {
	that.delNode(nodeId)
	log.Info("OnDelNode", "nodeId", nodeId)
}

//delNode remove from node list
func (that *BlockSyncMgr) delNode(nodeId uint64) {
	that.lock.Lock()
	defer that.lock.Unlock()
	delete(that.nodeWeights, nodeId)
	log.Info("delNode", "nodeId", nodeId)
	if len(that.nodeWeights) == 0 {
		log.Warn("no sync nodes")
	}
	log.Info("OnDelNode", "nodeId", nodeId)
}

func (that *BlockSyncMgr) tryGetSyncHeaderLock() bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	if that.syncHeaderLock {
		return true
	}
	that.syncHeaderLock = true
	return false
}

func (that *BlockSyncMgr) releaseSyncHeaderLock() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.syncHeaderLock = false
}

func (that *BlockSyncMgr) tryGetSyncBlockLock() bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	if that.syncBlockLock {
		return true
	}
	that.syncBlockLock = true
	return false
}

func (that *BlockSyncMgr) releaseSyncBlockLock() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.syncBlockLock = false
}

func (that *BlockSyncMgr) addBlockCache(nodeID uint64, block *types.Block) bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	blockInfo := &BlockInfo{
		nodeID: nodeID,
		block:  block,
	}
	that.blocksCache[block.Header.Height] = blockInfo
	return true
}

func (that *BlockSyncMgr) getBlockCache(blockHeight uint64) (uint64, *types.Block) {
	that.lock.RLock()
	defer that.lock.RUnlock()
	blockInfo, ok := that.blocksCache[blockHeight]
	if !ok {
		return 0, nil
	}
	return blockInfo.nodeID, blockInfo.block
}

func (that *BlockSyncMgr) delBlockCache(blockHeight uint64) {
	that.lock.Lock()
	defer that.lock.Unlock()
	delete(that.blocksCache, blockHeight)
}

// used
func (that *BlockSyncMgr) tryGetSaveBlockLock() bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	if that.saveBlockLock {
		return true
	}
	that.saveBlockLock = true
	return false
}

func (that *BlockSyncMgr) releaseSaveBlockLock() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.saveBlockLock = false
}

func (that *BlockSyncMgr) saveBlock() {
	if that.tryGetSaveBlockLock() {
		return
	}
	defer that.releaseSaveBlockLock()
	curBlockHeight := that.ledger.GetCurrentBlockHeight()
	nextBlockHeight := curBlockHeight + 1
	that.lock.Lock()
	for height := range that.blocksCache {
		if uint64(height) <= curBlockHeight {
			delete(that.blocksCache, height)
		}
	}
	that.lock.Unlock()
	for {
		fromID, nextBlock := that.getBlockCache(nextBlockHeight)
		if nextBlock == nil {
			return
		}
		if that.lockBlkHeight == nextBlockHeight && that.syncStatus == true {
			fmt.Printf("[p2p][block_sync]saveBlock block save lock at:%v\n", that.lockBlkHeight)
			return
		}

		blockInfo, err := validation.ValidateBlock(nextBlock, that.ledger)
		if err == nil {
			if blockInfo.ObjTxs.Len() > 0 {
				err = that.checkLeaguesBlock(blockInfo.ObjTxs)
				if err != nil {
					log.Error("[p2p][block_sync] checkLeaguesBlock error", "err", err)
					return
				}
			}
			err = that.ledger.SaveAll(blockInfo)
			fmt.Printf("[p2p][block_sync] save received block data: hash=%s,height=%d,err=%v\n", blockInfo.Block.Hash(), blockInfo.Block.Header.Height, err)
		}
		that.delBlockCache(nextBlockHeight)
		if err != nil { //flag001
			that.addErrorRespCnt(fromID)
			n := that.getNodeWeight(fromID)
			if n != nil && n.GetErrorRespCnt() >= SYNC_MAX_ERROR_RESP_TIMES {
				that.delNode(fromID)
			}
			log.Warn("[p2p]saveBlock", "Height", nextBlockHeight, "err", err)
			reqNode := that.getNextNode(nextBlockHeight)
			if reqNode == nil {
				return
			}
			that.addFlightBlock(reqNode.GetID(), nextBlockHeight, nextBlock.Hash())
			var nuladd common.Address
			msg := msgpack.NewBlkDataReq(nextBlock.Hash(), nuladd, p2pComm.SYNC_DATA_MAIN)
			err := that.server.Send(reqNode, msg, false)
			if err != nil {
				log.Warn("[p2p]require new block", "err", err)
				return
			} else {
				that.appendReqTime(reqNode.GetID())
			}
			return
		}
		nextBlockHeight++
		that.pingOutsyncNodes(nextBlockHeight - 1)
	}
}

func (that *BlockSyncMgr) checkLeaguesBlock(objTxs types.Transactions) error {
	t := time.Now()
	fmt.Printf("[p2p][block_sync] checkLeaguesBlock [id=%v]begin:%v\n", t.Nanosecond(), t.String())
	defer fmt.Printf("[p2p][block_sync] checkLeaguesBlock [id=%v]end:%v\n", t.Nanosecond(), time.Now().String())

	lsp := mainchain.NewLeagueStatePipe()
	leaguesNum := that.radar.CheckLeaguesResult(objTxs, lsp)
	for leaguesNum > 0 {
		//// [test] for print node info
		//for leagueId,value := range that.radar.ConvertTxsToObjTxForCheck(objTxs)  {
		//	fmt.Printf("[p2p][block_sync]ConvertTxsToObjTxForCheck,leagueId=%v,value=%v\n", leagueId,value)
		//	for key,value := range  mainchain.GetNodeLHCacheInstance().GetNodeLeagueHeight(leagueId) {
		//		fmt.Printf("[p2p][block_sync]nodeId:%v,nodeHeight:%v,nodePeerHandle:%v\n",key,value,that.server.GetNodeFromDiscoverID(key))
		//	}
		//}
		select {
		case sign := <-lsp.StateSignal:
			switch sign.(type) {
			case *mainchain.LeagueNeed:
				leagueBlockNeed := sign.(*mainchain.LeagueNeed)
				fmt.Println("[p2p][block_sync] LeagueNeedBlock:", leagueBlockNeed.LeagueId.ToString(), leagueBlockNeed.BlockHeight)
				//nodeLH := mainchain.GetNodeLHCacheInstance().GetNodeLeagueHeight(leagueBlockNeed.LeagueId)
				//for nodeIdStr, height := range nodeLH {
				//	if height < leagueBlockNeed.BlockHeight {
				//		continue
				//	}
				//	reqNode := that.server.GetNodeFromDiscoverID(nodeIdStr)
				//	if reqNode == nil {
				//		fmt.Printf("[p2p][block_sync] server.GetNodeFromDiscoverID is nil, nodeIdStr=%v\n",nodeIdStr)
				//		continue
				//	}
				//	fmt.Printf("[p2p][block_sync]nodeID%v,height:%v\n", nodeIdStr, height)
				//	extDataReq := &p2pComm.ExtDataResponse{
				//		IsReq:true,
				//		LeagueId: leagueBlockNeed.LeagueId,
				//		Height: leagueBlockNeed.BlockHeight,
				//		Data:nil,
				//	}
				//	msg := msgpack.NewExtMsg(extDataReq)
				//	err := that.server.Send(reqNode, msg, false)
				//	if err != nil {
				//		log.Warn("[p2p]require ext block msg", "err", err)
				//		continue
				//	} else {
				//		that.appendReqTime(reqNode.GetID())
				//	}
				//}
				extDataReq := &p2pComm.ExtDataRequest{
					LeagueId: leagueBlockNeed.LeagueId,
					Height:   leagueBlockNeed.BlockHeight,
				}
				err := that.server.Xmit(extDataReq)
				if nil != err {
					log.Warn("[p2p]error xmit message", "err", err.Error(), "Message", reflect.TypeOf(extDataReq))
				}
			case *mainchain.LeagueExec:
				fmt.Println("[p2p][block_sync] LeagueExec")
			case *mainchain.LeagueErr:
				// reget block info
				leagueErr := sign.(*mainchain.LeagueErr)
				fmt.Println("[p2p][block_sync] LeagueErr:", leagueErr)
			}
		case addr, ok := <-lsp.Successed:
			if ok {
				leaguesNum--
				if leaguesNum <= 0 {
					fmt.Println("[p2p][block_sync] Check all OK")
					that.radar.CommitLastHeight()
				}
				fmt.Printf("[p2p][block_sync] radar.CheckLeaguesResult,leaguesAddr[%v]\n", addr.ToString())
			}
			//leaguesNum--
			//if leaguesNum <= 0 {
			//	fmt.Println("[p2p][block_sync] Check all OK")
			//	that.radar.CommitLastHeight()
			//}
			//fmt.Printf("[p2p][block_sync] radar.CheckLeaguesResult,leaguesNum[%v]\n", leaguesNum)
			//case time.After(time.Second * 3):
			//	leaguesNum = -1
		}
	}

	return nil
}

func (that *BlockSyncMgr) isInBlockCache(blockHeight uint64) bool {
	that.lock.RLock()
	defer that.lock.RUnlock()
	_, ok := that.blocksCache[blockHeight]
	return ok
}

func (that *BlockSyncMgr) getBlockCacheSize() int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return len(that.blocksCache)
}

func (that *BlockSyncMgr) addFlightHeader(nodeId uint64, height uint64) {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.flightHeaders[height] = NewSyncFlightInfo(height, nodeId)
}

func (that *BlockSyncMgr) getFlightHeader(height uint64) *SyncFlightInfo {
	that.lock.RLock()
	defer that.lock.RUnlock()
	info, ok := that.flightHeaders[height]
	if !ok {
		return nil
	}
	return info
}

func (that *BlockSyncMgr) delFlightHeader(height uint64) bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	_, ok := that.flightHeaders[height]
	if !ok {
		return false
	}
	delete(that.flightHeaders, height)
	return true
}

func (that *BlockSyncMgr) getFlightHeaderCount() int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return len(that.flightHeaders)
}

// used
func (that *BlockSyncMgr) isHeaderOnFlight(height uint64) bool {
	flightInfo := that.getFlightHeader(height)
	return flightInfo != nil
}

func (that *BlockSyncMgr) addFlightBlock(nodeId uint64, height uint64, blockHash common.Hash) {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.flightBlocks[blockHash] = append(that.flightBlocks[blockHash], NewSyncFlightInfo(height, nodeId))
}

func (that *BlockSyncMgr) getFlightBlock(blockHash common.Hash) []*SyncFlightInfo {
	that.lock.RLock()
	defer that.lock.RUnlock()
	info, ok := that.flightBlocks[blockHash]
	if !ok {
		return nil
	}
	return info
}

func (that *BlockSyncMgr) delFlightBlock(blockHash common.Hash) bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	_, ok := that.flightBlocks[blockHash]
	if !ok {
		return false
	}
	delete(that.flightBlocks, blockHash)
	return true
}

func (that *BlockSyncMgr) getFlightBlockCount() int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	cnt := 0
	for hash := range that.flightBlocks {
		cnt += len(that.flightBlocks[hash])
	}
	return cnt
}

func (that *BlockSyncMgr) isBlockOnFlight(blockHash common.Hash) bool {
	flightInfos := that.getFlightBlock(blockHash)
	if len(flightInfos) != 0 {
		return true
	}
	return false
}

func (that *BlockSyncMgr) getNextNode(nextBlockHeight uint64) *peer.Peer {
	weights := that.getAllNodeWeights()
	sort.Sort(sort.Reverse(weights))
	nodelist := make([]uint64, 0)
	for _, n := range weights {
		nodelist = append(nodelist, n.id)
	}
	nextNodeIndex := 0
	triedNode := make(map[uint64]bool, 0)
	for {
		var nextNodeId uint64
		nextNodeIndex, nextNodeId = getNextNodeId(nextNodeIndex, nodelist)
		if nextNodeId == 0 {
			return nil
		}
		_, ok := triedNode[nextNodeId]
		if ok {
			return nil
		}
		triedNode[nextNodeId] = true
		n := that.server.GetNode(nextNodeId)
		if n == nil {
			continue
		}
		if n.GetSyncState() != p2pComm.ESTABLISH {
			continue
		}
		nodeBlockHeight := n.GetHeight()
		if nextBlockHeight <= nodeBlockHeight {
			return n
		}
	}
}

// used
// 1. Poll to find a node higher than the current height
// 2. If the node has not requested the height yet, use the node
// 3. Otherwise find the node with the fewest failures
func (that *BlockSyncMgr) getNodeWithMinFailedTimes(flightInfo *SyncFlightInfo, curBlockHeight uint64) *peer.Peer {
	var minFailedTimes = math.MaxInt64
	var minFailedTimesNode *peer.Peer
	triedNode := make(map[uint64]bool, 0)
	for {
		nextNode := that.getNextNode(curBlockHeight + 1)
		if nextNode == nil {
			return nil
		}
		failedTimes := flightInfo.GetFailedTimes(nextNode.GetID())
		if failedTimes == 0 {
			return nextNode
		}
		_, ok := triedNode[nextNode.GetID()]
		if ok {
			return minFailedTimesNode
		}
		triedNode[nextNode.GetID()] = true
		if failedTimes < minFailedTimes {
			minFailedTimes = failedTimes
			minFailedTimesNode = nextNode
		}
	}
}

//Stop to sync
func (that *BlockSyncMgr) Close() {
	close(that.exitCh)
}

//Stop to sync
func (that *BlockSyncMgr) SetSyncStatus(status bool, lockHeight uint64) {
	that.syncStatus = status
	that.lockBlkHeight = lockHeight
}

//getNodeWeight get nodeweight by id // used
func (that *BlockSyncMgr) getNodeWeight(nodeId uint64) *NodeWeight {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return that.nodeWeights[nodeId]
}

//getAllNodeWeights get all nodeweight and return a slice
func (that *BlockSyncMgr) getAllNodeWeights() NodeWeights {
	that.lock.RLock()
	defer that.lock.RUnlock()
	weights := make(NodeWeights, 0, len(that.nodeWeights))
	for _, w := range that.nodeWeights {
		weights = append(weights, w)
	}
	return weights
}

//addTimeoutCnt incre a node's timeout count // used
func (that *BlockSyncMgr) addTimeoutCnt(nodeId uint64) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AddTimeoutCnt()
	}
}

//addErrorRespCnt incre a node's error resp count // used
func (that *BlockSyncMgr) addErrorRespCnt(nodeId uint64) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AddErrorRespCnt()
	}
}

//appendReqTime append a node's request time // used
func (that *BlockSyncMgr) appendReqTime(nodeId uint64) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AppendNewReqtime()
	}
}

//addNewSpeed apend the new speed to tail, remove the oldest one
func (that *BlockSyncMgr) addNewSpeed(nodeId uint64, speed float32) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AppendNewSpeed(speed)
	}
}

//pingOutsyncNodes send ping msg to lower height nodes for syncing
func (that *BlockSyncMgr) pingOutsyncNodes(curHeight uint64) {
	peers := make([]*peer.Peer, 0)
	that.lock.RLock()
	maxHeight := curHeight
	for id := range that.nodeWeights {
		peer := that.server.GetNode(id)
		if peer == nil {
			continue
		}
		peerHeight := peer.GetHeight()
		if peerHeight >= maxHeight {
			maxHeight = peerHeight
		}
		if peerHeight < curHeight {
			peers = append(peers, peer)
		}
	}
	that.lock.RUnlock()
	if curHeight > maxHeight-SYNC_MAX_HEIGHT_OFFSET && len(peers) > 0 {
		that.server.PingTo(peers, false)
	}
}

//Using polling for load balance
func getNextNodeId(nextNodeIndex int, nodeList []uint64) (int, uint64) {
	num := len(nodeList)
	if num == 0 {
		return 0, 0
	}
	if nextNodeIndex >= num {
		nextNodeIndex = 0
	}
	index := nextNodeIndex
	nextNodeIndex++
	return nextNodeIndex, nodeList[index]
}
