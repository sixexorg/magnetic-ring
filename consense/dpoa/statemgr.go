package dpoa

import (
	"time"

	"github.com/sixexorg/magnetic-ring/log"
	//"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	//"github.com/sixexorg/magnetic-ring/core/types"
	"fmt"
)

type StateEventType int

type PeerState struct {
	peerIdx           string
	committedBlockNum uint64
	connected         bool
}

type stateChange struct {
	currentState     ServerState
}

const (
	ConfigLoaded     StateEventType = iota
	UpdatePeerState                 // notify statemgmt on peer heartbeat
)

type ServerState int

const (
	Init ServerState = iota
	LocalConfigured
	Configured
	Syncing
	WaitNetworkReady
	SyncReady
)

type StateEvent struct {
	Type      StateEventType
	peerState *PeerState
	blockNum  uint32
}

func isReady(state ServerState) bool {
	return state >= SyncReady
}

type StateMgr struct {
	syncReadyTimeout time.Duration
	currentState     ServerState
	StateEventC      chan *StateEvent
	peers            map[string]*PeerState
	liveTicker             *time.Timer
	lastTickChainHeight    uint32
	lastBlockSyncReqHeight uint64
	store            *BlockStore
	peerpool         *PeerPool
	notifyObj        *Feed
	cfg              *Config
	quitC            chan struct{}
}

func NewStateMgr(store *BlockStore, notifyObj *Feed, cfg *Config) *StateMgr {
	return &StateMgr{
		store:             store,
		syncReadyTimeout: time.Second * 10,
		currentState:     Init,
		StateEventC:      make(chan *StateEvent, 16),
		peers:            make(map[string]*PeerState),
		notifyObj:        notifyObj,
		cfg:              cfg,
		quitC:            make(chan struct{}),
	}
}

func (self *StateMgr) run() {
	go func() {
	for {
		select {
		case evt := <-self.StateEventC:
			switch evt.Type {
			case ConfigLoaded:
				if self.currentState == Init {
					self.currentState = LocalConfigured
				}
			case UpdatePeerState:
				if evt.peerState.connected {
					if err := self.onPeerUpdate(evt.peerState); err != nil {
						log.Error("statemgr process peer (%d) err: %s", evt.peerState.peerIdx, err)
					}
				} else {
					if err := self.onPeerDisconnected(evt.peerState.peerIdx); err != nil {
						log.Error("statmgr process peer (%d) disconn err: %s", evt.peerState.peerIdx, err)
					}
				}
			}

		case <-self.quitC:
			log.Info("server %d, state mgr quit", self.cfg.accountStr)
			return
		}
	}
	}()
}

func (self *StateMgr) onPeerDisconnected(peerIdx string) error {

	if _, present := self.peers[peerIdx]; !present {
		return nil
	}
	delete(self.peers, peerIdx)

	// start another connection if necessary
	if self.currentState == SyncReady {
		if self.peerpool.getActivePeerCount() < self.getMinActivePeerCount() {
			self.currentState = WaitNetworkReady
			st := stateChange{
				currentState: self.currentState,
			}
			self.notifyObj.Send(st)
		}
	}

	return nil
}

func (self *StateMgr) getConsensusedCommittedBlockNum() (uint64, bool) {
	C := len(self.store.GetCurStars()) / 2
	consensused := false
	var maxCommitted uint64
	myCommitted := self.store.getLatestBlockNumber()
	peers := make(map[uint64][]string)
	for _, p := range self.peers {
		n := p.committedBlockNum
		if n >= myCommitted {
			if _, present := peers[n]; !present {
				peers[n] = make([]string, 0)
			}
			for k := range peers {
				if n >= k {
					peers[k] = append(peers[k], p.peerIdx)
				}
			}
			if len(peers[n]) >= C {
				maxCommitted = n
				consensused = true
			}
		}
	}

	return maxCommitted, consensused
}

func (self *StateMgr) isSyncedReady() bool {
	// check action peer connections
	if len(self.peers) < self.getMinActivePeerCount() {
		return false
	}
	// check chain consensus
	committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
	if !ok {
		fmt.Println("StateMgr getConsensusedCommittedBlockNum", committedBlkNum, ok)
		return false
	}
	fmt.Println("StateMgr getLatestBlockNumber",self.store.getLatestBlockNumber(), committedBlkNum)
	if self.store.getLatestBlockNumber() >= committedBlkNum {
		return true
	}
	return false
}

func (self *StateMgr) setSyncedReady() error {
	prevState := self.currentState
	self.currentState = SyncReady
	if prevState <= SyncReady {
		log.Debug("server %v start sync ready", "self.cfg.accountStr", self.cfg.accountStr)
		//fmt.Println("1@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@setSyncedReady", prevState, self.store.getLatestBlockNumber()+1)
		//<-time.After(time.Second*3)
		//fmt.Println("2@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@setSyncedReady", prevState, self.store.getLatestBlockNumber()+1)

		st := stateChange{
			currentState:self.currentState,
		}
		self.notifyObj.Send(st)
	}

	return nil
}

func (self *StateMgr) onPeerUpdate(peerState *PeerState) error {
	peerIdx := peerState.peerIdx
	newPeer := false
	if _, present := self.peers[peerIdx]; !present {
		newPeer = true
	}

	log.Info("server peer update ", "accountStr", self.cfg.accountStr, "currentblk", self.store.getLatestBlockNumber()+1, "state", self.currentState,
		"frompeer", peerState.peerIdx, "committed", peerState.committedBlockNum, "newPeer", newPeer)

	self.peers[peerIdx] = peerState

	if !newPeer {
		if !self.store.isEarth() && peerState.peerIdx == self.store.earthNode(){
			//if isReady(self.currentState) && peerState.committedBlockNum >= self.store.getLatestBlockNumber()+1 {
			//	log.Warn("server seems lost sync with earth:", "accountStr", self.cfg.accountStr, "committedBlockNum",
			//		peerState.committedBlockNum, "peerIdx", peerState.peerIdx, "getLatestBlockNumber",self.store.getLatestBlockNumber()+1)
			//	if err := self.checkStartSyncing(self.store.getLatestBlockNumber(), true); err != nil {
			//		log.Error("server start syncing check failed", "accountStr", self.cfg.accountStr)
			//		return nil
			//	}
			//}
			return nil
		}else {
			if isReady(self.currentState) && peerState.committedBlockNum > self.store.getLatestBlockNumber()+1+10 {
				log.Warn("server seems lost sync", "accountStr", self.cfg.accountStr, "committedBlockNum",
					peerState.committedBlockNum, "peerIdx", peerState.peerIdx, "getLatestBlockNumber", self.store.getLatestBlockNumber()+1)
					return nil
			}
		}
	}

	switch self.currentState {
	case LocalConfigured:
		log.Info("server statemgr update","accountStr", self.cfg.accountStr, "currentState",self.currentState, "peerIdx", peerIdx, "peercnt",len(self.peers))
		if self.getSyncedPeers() > 0{
			self.currentState = Syncing
		}
	case Configured:
	case Syncing:
		if peerState.committedBlockNum > self.store.getLatestBlockNumber() {
			committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
			if ok && committedBlkNum > self.store.getLatestBlockNumber() {
				log.Info("server syncing","accountStr", self.cfg.accountStr, "syncing",self.store.getLatestBlockNumber(), "target",committedBlkNum)
				self.checkStartSyncing(self.store.getLatestBlockNumber(), false)
			}
		}
		if self.isSyncedReady() {
			log.Debug("server synced from syncing", "accountStr", self.cfg.accountStr)
			fmt.Println("==>>>setSyncedReady")
			if err := self.setSyncedReady(); err != nil {
				log.Warn("server state set syncready", self.cfg.account.PublicKey, self.currentState, err)
			}
		}
	case WaitNetworkReady:
		if self.isSyncedReady() {
			fmt.Println("==>>>WaitNetworkReady")
			log.Debug("server synced from sync-ready", "accountStr", self.cfg.accountStr)
			if err := self.setSyncedReady(); err != nil {
				log.Warn("server state set syncready", "accountStr", self.cfg.accountStr, "currentState", self.currentState, "err", err)
			}
		}
	case SyncReady:
		committedBlkNum, ok := self.getConsensusedCommittedBlockNum()
		if ok && committedBlkNum > self.store.getLatestBlockNumber()+1 {
			log.Info("server synced try fastforward from", "accountStr", self.cfg.accountStr, "getLatestBlockNumber", self.store.getLatestBlockNumber())
			self.checkStartSyncing(self.store.getLatestBlockNumber(), false)
		}
	}

	return nil
}

func (self *StateMgr) checkStartSyncing(startBlkNum uint64, forceSync bool) error {
	if self.store.nonConsensusNode() {
		return nil
	}

	var maxCommitted uint64
	peers := make(map[uint64][]string)
	for _, p := range self.peers {
		n := p.committedBlockNum
		if n > startBlkNum {
			if _, present := peers[n]; !present {
				peers[n] = make([]string, 0)
			}
			for k := range peers {
				if n >= k {
					peers[k] = append(peers[k], p.peerIdx)
				}
			}
			if len(peers[n]) >= len(self.store.GetCurStars())/2 {
				maxCommitted = n
			}
		}
	}

	if forceSync && maxCommitted == 0{
		maxCommitted = startBlkNum
	}

	if maxCommitted > startBlkNum || forceSync {
		log.Debug("server startBlkNum forceSync maxCommitted lastBlockSyncReqHeight", "accountStr",
			self.cfg.accountStr, "startBlkNum", startBlkNum, "forceSync", forceSync, "maxCommitted", maxCommitted, "lastBlockSyncReqHeight", self.lastBlockSyncReqHeight)
		preState := self.currentState
		self.currentState = Syncing
		if preState == SyncReady{
			st := stateChange{
				currentState: self.currentState,
			}
			self.notifyObj.Send(st)
		}
		startBlkNum = self.store.getLatestBlockNumber() + 1

		if maxCommitted > self.lastBlockSyncReqHeight {
			log.Info("server start syncing", "accountStr", self.cfg.accountStr, "startBlkNum", startBlkNum, "maxCommitted", maxCommitted, "peers", peers)
			self.lastBlockSyncReqHeight = maxCommitted
		}
	}

	return nil
}

func (self *StateMgr) getSyncedPeers() uint32 {
	if len(self.peers) < self.getMinActivePeerCount() {
		return 0
	}

	return uint32(len(self.peers))
}

func (self *StateMgr) getMinActivePeerCount() int {
	if self.store.isEarth(){
		return len(self.store.GetCurStars())/4
	}

	return len(self.store.GetCurStars())/2
}

func (self *StateMgr) getState() ServerState {
	return self.currentState
}
