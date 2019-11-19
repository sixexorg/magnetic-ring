package org

import (
	"github.com/sixexorg/magnetic-ring/p2pserver/temp"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/log"
	p2pComm "github.com/sixexorg/magnetic-ring/p2pserver/common"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"
	"github.com/sixexorg/magnetic-ring/p2pserver/sync/p2pserprotocol"
	ledger "github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"github.com/sixexorg/magnetic-ring/store/orgchain/validation"
)

const (
	SYNC_MAX_HEADER_FORWARD_SIZE = 5000       //keep CurrentHeaderHeight - CurrentBlockHeight <= SYNC_MAX_HEADER_FORWARD_SIZE
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
	Height      uint32         //BlockHeight of HeaderHeight
	nodeId      uint64         //The current node to send msg
	startTime   time.Time      //Request start time
	failedNodes map[uint64]int //Map nodeId => timeout times
	totalFailed int            //Total timeout times
	lock        sync.RWMutex
}

//NewSyncFlightInfo return a new SyncFlightInfo instance
func NewSyncFlightInfo(height uint32, nodeId uint64) *SyncFlightInfo {
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

//OrgBlockSyncMgr is the manager class to deal with block sync
type OrgBlockSyncMgr struct {
	flightBlocks   map[common.Hash][]*SyncFlightInfo //Map BlockHash => []SyncFlightInfo, using for manager all of those block flights
	flightHeaders  map[uint32]*SyncFlightInfo        //Map HeaderHeight => SyncFlightInfo, using for manager all of those header flights
	blocksCache    map[uint32]*BlockInfo             //Map BlockHash => BlockInfo, using for cache the blocks receive from net, and waiting for commit to ledger
	server         p2pserprotocol.SyncP2PSer         //Pointer to the local node
	syncBlockLock  bool                              //Help to avoid send block sync request duplicate
	syncHeaderLock bool                              //Help to avoid send header sync request duplicate
	saveBlockLock  bool                              //Help to avoid saving block concurrently
	exitCh         chan interface{}                  //ExitCh to receive exit signal
	ledger         *ledger.LedgerStoreImp            //ledger
	lock           sync.RWMutex                      //lock
	nodeWeights    map[uint64]*NodeWeight            //Map NodeID => NodeStatus, using for getNextNode
	// org
	orgID common.Address
}

//NewOrgBlockSyncMgr return a OrgBlockSyncMgr instance
func NewOrgBlockSyncMgr(server p2pserprotocol.SyncP2PSer, orgID common.Address) *OrgBlockSyncMgr {
	return &OrgBlockSyncMgr{
		flightBlocks:  make(map[common.Hash][]*SyncFlightInfo, 0),
		flightHeaders: make(map[uint32]*SyncFlightInfo, 0),
		blocksCache:   make(map[uint32]*BlockInfo, 0),
		server:        server,
		// ledger:        server.GetLedger(),
		ledger: temp.GetLedger(orgID),

		exitCh:      make(chan interface{}, 1),
		nodeWeights: make(map[uint64]*NodeWeight, 0),
		orgID:       orgID,
	}
}

//Start to sync
func (that *OrgBlockSyncMgr) Start() {
	getledger := func() {
		if that.ledger == nil {
			that.ledger = temp.GetLedger(that.orgID)
		}
	}
	for true {
		getledger()
		if that.ledger != nil {
			break
		}
		time.Sleep(time.Second)
		//fmt.Println(" ****** for true for true for true ...... ")
	}

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


func (that *OrgBlockSyncMgr) printHeight() {
	t := time.NewTicker(p2pComm.DEFAULT_GEN_BLOCK_TIME * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			log.Info(" **** orgID", "orgID", that.orgID, "CurrentBlockHeight", that.ledger.GetCurrentBlockHeight( /*that.orgID*/ ))
			//fmt.Println(" ******* CurrentHeaderHeight:", that.ledger.GetCurrentHeaderHeight())
		case <-that.exitCh:
			return
		}
	}
}

func (that *OrgBlockSyncMgr) checkTimeout() {
	now := time.Now()
	headerTimeoutFlights := make(map[uint32]*SyncFlightInfo, 0)
	blockTimeoutFlights := make(map[common.Hash][]*SyncFlightInfo, 0)
	that.lock.RLock()
	//
	for height, flightInfo := range that.flightHeaders {
		//
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

	curHeaderHeight := that.ledger.GetCurrentHeaderHeight( /*that.orgID*/ )
	curBlockHeight := that.ledger.GetCurrentBlockHeight( /*that.orgID*/ )

	//Header request processing after timeout
	for height, flightInfo := range headerTimeoutFlights {
		that.addTimeoutCnt(flightInfo.GetNodeId())
		if height <= uint32(curHeaderHeight) {
			that.delFlightHeader(height)
			continue
		}
		flightInfo.ResetStartTime()
		flightInfo.MarkFailedNode()
		log.Trace("[p2p]checkTimeout org sync headers from", "id", flightInfo.GetNodeId(),
			"height", height, "after", SYNC_HEADER_REQUEST_TIMEOUT, "Times", flightInfo.GetTotalFailedTimes())
		reqNode := that.getNodeWithMinFailedTimes(flightInfo, uint32(curBlockHeight))
		if reqNode == nil {
			break
		}
		flightInfo.SetNodeId(reqNode.GetID())

		headerHash := that.ledger.GetCurrentHeaderHash( /*that.orgID*/ )
		msg := msgpack.NewHeadersReq(headerHash, that.orgID, p2pComm.SYNC_DATA_ORG)
		err := that.server.Send(reqNode, msg, false)
		if err != nil {
			log.Warn("[p2p]checkTimeout failed to send a new headersReq", "err", err)
		} else {
			that.appendReqTime(reqNode.GetID())
		}
	}
	//
	for blockHash, flightInfos := range blockTimeoutFlights {
		for _, flightInfo := range flightInfos {
			that.addTimeoutCnt(flightInfo.GetNodeId())
			if flightInfo.Height <= uint32(curBlockHeight) {
				that.delFlightBlock(blockHash)
				continue
			}
			flightInfo.ResetStartTime()
			flightInfo.MarkFailedNode()
			log.Trace("[p2p]checkTimeout sync", "height", flightInfo.Height, "block", blockHash, "after", SYNC_BLOCK_REQUEST_TIMEOUT, "times", flightInfo.GetTotalFailedTimes())
			reqNode := that.getNodeWithMinFailedTimes(flightInfo, uint32(curBlockHeight))
			if reqNode == nil {
				break
			}
			flightInfo.SetNodeId(reqNode.GetID())

			msg := msgpack.NewBlkDataReq(blockHash, that.orgID, p2pComm.SYNC_DATA_ORG)
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

func (that *OrgBlockSyncMgr) sync() {
	that.syncHeader()
	that.syncBlock()
}

func (that *OrgBlockSyncMgr) reachMinConnection() bool {

	minCount := p2pComm.ORG_MIN_NODE_NUM

	that.lock.Lock()
	defer that.lock.Unlock()

	return len(that.nodeWeights) >= minCount
}

func (that *OrgBlockSyncMgr) syncHeader() {

	if !that.reachMinConnection() {
		log.Info("[OrgBlockSyncMgr] orgid:%v Wait for minimum connection", "orgID", that.orgID)
		return
	}

	if that.tryGetSyncHeaderLock() {
		return
	}
	defer that.releaseSyncHeaderLock()


	if that.getFlightHeaderCount() >= SYNC_MAX_FLIGHT_HEADER_SIZE {
		return
	}
	curBlockHeight := that.ledger.GetCurrentBlockHeight( /*that.orgID*/ )

	curHeaderHeight := that.ledger.GetCurrentHeaderHeight( /*that.orgID*/ )
	//Waiting for block catch up header
	if curHeaderHeight-curBlockHeight >= SYNC_MAX_HEADER_FORWARD_SIZE {
		return
	}
	NextHeaderId := uint32(curHeaderHeight) + 1
	reqNode := that.getNextNode(NextHeaderId)
	if reqNode == nil {
		return
	}
	that.addFlightHeader(reqNode.GetID(), NextHeaderId)

	headerHash := that.ledger.GetCurrentHeaderHash( /*that.orgID*/ )

	testHeight := that.ledger.GetCurrentHeaderHeight()
	msg := msgpack.NewHeadersReq(headerHash, that.orgID, p2pComm.SYNC_DATA_ORG, testHeight)

	err := that.server.Send(reqNode, msg, false)
	if err != nil {
		log.Warn("[p2p]syncHeader failed to send a new headersReq")
	} else {
		that.appendReqTime(reqNode.GetID())
	}

	log.Info("sub chain--> Header sync request ", "height", NextHeaderId)
}

// Storage provides two interfaces, one is to take data from the database, and the other is to take data from memory.
func (that *OrgBlockSyncMgr) MainGetBlockHashByHeight(nextBlockHeight uint64) common.Hash {
	nextBlockHeaderHash := that.ledger.GetBlockHeaderHashByHeight(nextBlockHeight)
	empty := common.Hash{}
	if nextBlockHeaderHash != empty {
		return nextBlockHeaderHash
	}
	nextBlockHash, _ := that.ledger.GetBlockHashByHeight(nextBlockHeight)
	return nextBlockHash
}

func (that *OrgBlockSyncMgr) syncBlock() {
	if that.tryGetSyncBlockLock() {
		return
	}
	defer that.releaseSyncBlockLock()

	availCount := SYNC_MAX_FLIGHT_BLOCK_SIZE - that.getFlightBlockCount()
	if availCount <= 0 {
		return
	}
	curBlockHeight := uint32(that.ledger.GetCurrentBlockHeight( /*that.orgID*/ ))
	curHeaderHeight := uint32(that.ledger.GetCurrentHeaderHeight( /*that.orgID*/ ))
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
	i := uint32(0)
	reqTimes := 1
	for {
		if counter > count {
			break
		}
		i++
		nextBlockHeight := curBlockHeight + i
		// nextBlockHash,_ := that.ledger.GetBlockHashByHeight(uint64(nextBlockHeight)/*,that.orgID*/)
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
			msg := msgpack.NewBlkDataReq(nextBlockHash, that.orgID, p2pComm.SYNC_DATA_ORG)
			err := that.server.Send(reqNode, msg, false)
			if err != nil {
				log.Warn("[p2p]syncBlock ", "Height", nextBlockHeight, "ReqBlkData error", err)
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
func (that *OrgBlockSyncMgr) OnHeaderReceive(fromID uint64, headers []*types.Header) {
	if len(headers) == 0 {
		return
	}
	log.Info("Header receive ", "height", headers[0].Height, "Height", headers[len(headers)-1].Height)
	height := uint32(headers[0].Height)
	curHeaderHeight := uint32(that.ledger.GetCurrentHeaderHeight( /*that.orgID*/ ))

	//Means another gorountinue is adding header
	if height <= curHeaderHeight {
		return
	}
	if !that.isHeaderOnFlight(height) {
		return
	}
	err := that.ledger.AddHeaders(headers /*,that.orgID*/)
	that.delFlightHeader(height)
	if err != nil {
		that.addErrorRespCnt(fromID)
		n := that.getNodeWeight(fromID)
		if n != nil && n.GetErrorRespCnt() >= SYNC_MAX_ERROR_RESP_TIMES {
			that.delNode(fromID)
		}
		log.Warn("[p2p]OnHeaderReceive AddHeaders ", "error", err)
		return
	}
	that.syncHeader()
}

//OnBlockReceive receive block from net
func (that *OrgBlockSyncMgr) OnBlockReceive(fromID uint64, blockSize uint32, block *types.Block) {
	height := uint32(block.Header.Height)
	blockHash := block.Hash()
	log.Trace("[p2p]OnBlockReceive ", "Height", height)
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
	curHeaderHeight := uint32(that.ledger.GetCurrentHeaderHeight( /*that.orgID*/ ))
	nextHeader := curHeaderHeight + 1
	if height > nextHeader {
		return
	}
	curBlockHeight := that.ledger.GetCurrentBlockHeight( /*that.orgID*/ )
	if height <= uint32(curBlockHeight) {
		return
	}

	that.addBlockCache(fromID, block)
	go that.saveBlock()
	that.syncBlock()
}

//OnAddNode to node list when a new node added
func (that *OrgBlockSyncMgr) OnAddNode(nodeId uint64) {
	log.Debug("[p2p]OnAddNode", "nodeId", nodeId)
	that.lock.Lock()
	defer that.lock.Unlock()
	if _, ok := that.nodeWeights[nodeId]; !ok {
		that.server.SentConnectToBootNode(nodeId)

		w := NewNodeWeight(nodeId)
		that.nodeWeights[nodeId] = w
		//fmt.Println(" ****** OrgBlockSyncMgr OnAddNode nodeId:", nodeId)
	}
}

//OnDelNode remove from node list. When the node disconnect
func (that *OrgBlockSyncMgr) OnDelNode(nodeId uint64) {
	that.delNode(nodeId)
}

//delNode remove from node list

func (that *OrgBlockSyncMgr) delNode(nodeId uint64) {
	that.lock.Lock()
	defer that.lock.Unlock()

	if _, ok := that.nodeWeights[nodeId]; ok {
		that.server.SentDisconnectToBootNode(nodeId, false)
		remote := that.server.GetNode(nodeId)
		if remote != nil {
			remote.DelRemoteOrg(that.orgID)
		}
		log.Info(" ***** OnDelNode", "nodeId", nodeId)
	}

	delete(that.nodeWeights, nodeId)
	log.Info("delNode", "nodeId", nodeId)
	if len(that.nodeWeights) == 0 {
		log.Warn("no sync nodes")
	}
}

func (that *OrgBlockSyncMgr) tryGetSyncHeaderLock() bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	if that.syncHeaderLock {
		return true
	}
	that.syncHeaderLock = true
	return false
}

func (that *OrgBlockSyncMgr) releaseSyncHeaderLock() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.syncHeaderLock = false
}

func (that *OrgBlockSyncMgr) tryGetSyncBlockLock() bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	if that.syncBlockLock {
		return true
	}
	that.syncBlockLock = true
	return false
}

func (that *OrgBlockSyncMgr) releaseSyncBlockLock() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.syncBlockLock = false
}

func (that *OrgBlockSyncMgr) addBlockCache(nodeID uint64, block *types.Block) bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	blockInfo := &BlockInfo{
		nodeID: nodeID,
		block:  block,
	}
	that.blocksCache[uint32(block.Header.Height)] = blockInfo
	return true
}

func (that *OrgBlockSyncMgr) getBlockCache(blockHeight uint32) (uint64, *types.Block) {
	that.lock.RLock()
	defer that.lock.RUnlock()
	blockInfo, ok := that.blocksCache[blockHeight]
	if !ok {
		return 0, nil
	}
	return blockInfo.nodeID, blockInfo.block
}

func (that *OrgBlockSyncMgr) delBlockCache(blockHeight uint32) {
	that.lock.Lock()
	defer that.lock.Unlock()
	delete(that.blocksCache, blockHeight)
}

// used
func (that *OrgBlockSyncMgr) tryGetSaveBlockLock() bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	if that.saveBlockLock {
		return true
	}
	that.saveBlockLock = true
	return false
}

func (that *OrgBlockSyncMgr) releaseSaveBlockLock() {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.saveBlockLock = false
}

func (that *OrgBlockSyncMgr) saveBlock() {
	// 只有这里用
	if that.tryGetSaveBlockLock() {
		return
	}
	defer that.releaseSaveBlockLock()
	curBlockHeight := uint32(that.ledger.GetCurrentBlockHeight( /*that.orgID*/ ))
	nextBlockHeight := curBlockHeight + 1
	that.lock.Lock()
	for height := range that.blocksCache {
		if height <= curBlockHeight {
			delete(that.blocksCache, height)
		}
	}
	that.lock.Unlock()
	for {
		fromID, nextBlock := that.getBlockCache(nextBlockHeight)
		if nextBlock == nil {
			return
		}
		// err := that.ledger.AddBlock(nextBlock,that.orgID)
		// swp
		blkInfo, err := validation.ValidateBlock(nextBlock, that.ledger)
		if err == nil {
			err = that.ledger.SaveAll(blkInfo)
		}

		that.delBlockCache(nextBlockHeight)
		if err != nil {
			that.addErrorRespCnt(fromID)
			n := that.getNodeWeight(fromID)
			if n != nil && n.GetErrorRespCnt() >= SYNC_MAX_ERROR_RESP_TIMES {
				that.delNode(fromID)
			}
			log.Warn("[p2p]saveBlock ", "Height", nextBlockHeight, " AddBlock error", err)
			reqNode := that.getNextNode(nextBlockHeight)
			if reqNode == nil {
				return
			}
			that.addFlightBlock(reqNode.GetID(), nextBlockHeight, nextBlock.Hash())
			msg := msgpack.NewBlkDataReq(nextBlock.Hash(), that.orgID, p2pComm.SYNC_DATA_ORG)
			err := that.server.Send(reqNode, msg, false)
			if err != nil {
				log.Warn("[p2p]require new block", " error", err)
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

func (that *OrgBlockSyncMgr) isInBlockCache(blockHeight uint32) bool {
	that.lock.RLock()
	defer that.lock.RUnlock()
	_, ok := that.blocksCache[blockHeight]
	return ok
}

func (that *OrgBlockSyncMgr) getBlockCacheSize() int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return len(that.blocksCache)
}

func (that *OrgBlockSyncMgr) addFlightHeader(nodeId uint64, height uint32) {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.flightHeaders[height] = NewSyncFlightInfo(height, nodeId)
}

func (that *OrgBlockSyncMgr) getFlightHeader(height uint32) *SyncFlightInfo {
	that.lock.RLock()
	defer that.lock.RUnlock()
	info, ok := that.flightHeaders[height]
	if !ok {
		return nil
	}
	return info
}

func (that *OrgBlockSyncMgr) delFlightHeader(height uint32) bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	_, ok := that.flightHeaders[height]
	if !ok {
		return false
	}
	delete(that.flightHeaders, height)
	return true
}

func (that *OrgBlockSyncMgr) getFlightHeaderCount() int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return len(that.flightHeaders)
}

// used
func (that *OrgBlockSyncMgr) isHeaderOnFlight(height uint32) bool {
	flightInfo := that.getFlightHeader(height)
	return flightInfo != nil
}

// used
func (that *OrgBlockSyncMgr) addFlightBlock(nodeId uint64, height uint32, blockHash common.Hash) {
	that.lock.Lock()
	defer that.lock.Unlock()
	that.flightBlocks[blockHash] = append(that.flightBlocks[blockHash], NewSyncFlightInfo(height, nodeId))
}

// used
func (that *OrgBlockSyncMgr) getFlightBlock(blockHash common.Hash) []*SyncFlightInfo {
	that.lock.RLock()
	defer that.lock.RUnlock()
	info, ok := that.flightBlocks[blockHash]
	if !ok {
		return nil
	}
	return info
}

// used
func (that *OrgBlockSyncMgr) delFlightBlock(blockHash common.Hash) bool {
	that.lock.Lock()
	defer that.lock.Unlock()
	_, ok := that.flightBlocks[blockHash]
	if !ok {
		return false
	}
	delete(that.flightBlocks, blockHash)
	return true
}

// used
func (that *OrgBlockSyncMgr) getFlightBlockCount() int {
	that.lock.RLock()
	defer that.lock.RUnlock()
	cnt := 0
	for hash := range that.flightBlocks {
		cnt += len(that.flightBlocks[hash])
	}
	return cnt
}

// used
func (that *OrgBlockSyncMgr) isBlockOnFlight(blockHash common.Hash) bool {
	flightInfos := that.getFlightBlock(blockHash)
	if len(flightInfos) != 0 {
		return true
	}
	return false
}

func (that *OrgBlockSyncMgr) getNextNode(nextBlockHeight uint32) *peer.Peer {
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
		nodeBlockHeight := n.GetRemoteOrgHeight(that.orgID)
		if nextBlockHeight <= uint32(nodeBlockHeight) {
			return n
		}
	}
}

// used
func (that *OrgBlockSyncMgr) getNodeWithMinFailedTimes(flightInfo *SyncFlightInfo, curBlockHeight uint32) *peer.Peer {
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
func (that *OrgBlockSyncMgr) Close() {
	close(that.exitCh)
}

//getNodeWeight get nodeweight by id // used
func (that *OrgBlockSyncMgr) getNodeWeight(nodeId uint64) *NodeWeight {
	that.lock.RLock()
	defer that.lock.RUnlock()
	return that.nodeWeights[nodeId]
}

//getAllNodeWeights get all nodeweight and return a slice
func (that *OrgBlockSyncMgr) getAllNodeWeights() NodeWeights {
	that.lock.RLock()
	defer that.lock.RUnlock()
	weights := make(NodeWeights, 0, len(that.nodeWeights))
	for _, w := range that.nodeWeights {
		weights = append(weights, w)
	}
	return weights
}

//addTimeoutCnt incre a node's timeout count // used
func (that *OrgBlockSyncMgr) addTimeoutCnt(nodeId uint64) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AddTimeoutCnt()
	}
}

//addErrorRespCnt incre a node's error resp count // used
func (that *OrgBlockSyncMgr) addErrorRespCnt(nodeId uint64) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AddErrorRespCnt()
	}
}

//appendReqTime append a node's request time // used
func (that *OrgBlockSyncMgr) appendReqTime(nodeId uint64) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AppendNewReqtime()
	}
}

//addNewSpeed apend the new speed to tail, remove the oldest one
func (that *OrgBlockSyncMgr) addNewSpeed(nodeId uint64, speed float32) {
	n := that.getNodeWeight(nodeId)
	if n != nil {
		n.AppendNewSpeed(speed)
	}
}

//pingOutsyncNodes send ping msg to lower height nodes for syncing
func (that *OrgBlockSyncMgr) pingOutsyncNodes(curHeight uint32) {
	peers := make([]*peer.Peer, 0)
	that.lock.RLock()
	maxHeight := curHeight
	for id := range that.nodeWeights {
		peer := that.server.GetNode(id)
		if peer == nil {
			continue
		}
		peerHeight := uint32(peer.GetRemoteOrgHeight(that.orgID))
		if peerHeight >= maxHeight {
			maxHeight = peerHeight
		}
		if peerHeight < curHeight {
			peers = append(peers, peer)
		}
	}
	that.lock.RUnlock()
	if curHeight > maxHeight-SYNC_MAX_HEIGHT_OFFSET && len(peers) > 0 {
		that.server.PingTo(peers, false, that.orgID)
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
