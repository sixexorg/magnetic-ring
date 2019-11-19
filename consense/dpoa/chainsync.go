package dpoa

import (
	"sync"

	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
)

type SyncCheckReq struct {
	msg      comm.ConsensusMsg
	peerIdx  uint64
	blockNum uint64
}

type BlockSyncReq struct {
	targetPeers    []string
	startBlockNum  uint64
	targetBlockNum uint64 // targetBlockNum == 0, as one cancel syncing request
}

type PeerSyncer struct {
	lock          sync.Mutex
	peerIdx       string
	nextReqBlkNum uint64
	targetBlkNum  uint64
	active        bool

	//server *Server
	stat   *StateMgr
	msgC   chan comm.ConsensusMsg
}

type SyncMsg struct {
	fromPeer string
	msg      comm.ConsensusMsg
}

type BlockMsgFromPeer struct {
	fromPeer string
	block    *comm.Block
}

type BlockFromPeers map[string]*comm.Block // index by peerId

type Syncer struct {
	lock   sync.Mutex
	//server *Server
	stat   *StateMgr
	cfg    *Config

	maxRequestPerPeer int
	nextReqBlkNum     uint64
	targetBlkNum      uint64

	syncCheckReqC  chan *SyncCheckReq
	blockSyncReqC  chan *BlockSyncReq
	syncMsgC       chan *SyncMsg // receive syncmsg from server
	//blockFromPeerC chan *BlockMsgFromPeer

	peers         map[string]*PeerSyncer
	pendingBlocks map[uint64]BlockFromPeers // index by blockNum
	quitC         chan struct{}
}

func newSyncer(stat *StateMgr, cfg *Config) *Syncer {
	return &Syncer{
		stat:              stat,
		cfg:               cfg,
		maxRequestPerPeer: 4,
		nextReqBlkNum:     1,
		syncCheckReqC:     make(chan *SyncCheckReq, 4),
		blockSyncReqC:     make(chan *BlockSyncReq, 16),
		syncMsgC:          make(chan *SyncMsg, 256),
		//blockFromPeerC:    make(chan *BlockMsgFromPeer, 64),
		peers:             make(map[string]*PeerSyncer),
		pendingBlocks:     make(map[uint64]BlockFromPeers),
	}
}

func (self *Syncer) stop() {
	self.lock.Lock()
	defer self.lock.Unlock()

	close(self.syncCheckReqC)
	close(self.blockSyncReqC)
	close(self.syncMsgC)
	//close(self.blockFromPeerC)

	self.peers = make(map[string]*PeerSyncer)
	self.pendingBlocks = make(map[uint64]BlockFromPeers)
}

func (self *Syncer) run() {
	//self.server.quitWg.Add(1)
	//defer self.server.quitWg.Done()
	go func() {
		for {
			select {
			case <-self.syncCheckReqC:
			case req := <-self.blockSyncReqC:
				if req.targetBlockNum == 0 {
					// cancel fetcher for peer
					for _, id := range req.targetPeers {
						self.cancelFetcherForPeer(self.peers[id])
					}
					continue
				}

				log.Info("server %v, got sync req(%d, %d) to %v",
					self.cfg.accountStr, req.startBlockNum, req.targetBlockNum, req.targetPeers)
				if req.startBlockNum <= self.stat.store.getLatestBlockNumber() {
					req.startBlockNum = self.stat.store.getLatestBlockNumber() + 1
					log.Info("server %v, sync req start change to %d",
						self.cfg.accountStr, req.startBlockNum)
					if req.startBlockNum > req.targetBlockNum {
						continue
					}
				}
				if err := self.onNewBlockSyncReq(req); err != nil {
					log.Error("server %d failed to handle new block sync req: %s", self.cfg.accountStr, err)
				}

			case syncMsg := <-self.syncMsgC:
				if p, present := self.peers[syncMsg.fromPeer]; present {
					if p.active {
						p.msgC <- syncMsg.msg
					} else {
						// report err
						p.msgC <- nil
					}
				} else {
					// report error
				}

			case <-self.quitC:
				log.Info("server %v, syncer quit", self.cfg.accountStr)
				return
			}
		}
	}()
}

func (self *Syncer) isActive() bool {
	return self.nextReqBlkNum <= self.targetBlkNum
}

func (self *Syncer) startPeerSyncer(syncer *PeerSyncer, targetBlkNum uint64) error {

	syncer.lock.Lock()
	defer syncer.lock.Unlock()

	if targetBlkNum > syncer.targetBlkNum {
		syncer.targetBlkNum = targetBlkNum
	}
	if syncer.targetBlkNum >= syncer.nextReqBlkNum && !syncer.active {
		syncer.active = true
		go func() {
			syncer.run()
		}()
	}

	return nil
}

func (self *Syncer) cancelFetcherForPeer(peer *PeerSyncer) error {
	if peer == nil {
		return nil
	}

	peer.lock.Lock()
	defer peer.lock.Unlock()

	// TODO

	return nil
}

func (self *Syncer) onNewBlockSyncReq(req *BlockSyncReq) error {
	if req.startBlockNum < self.nextReqBlkNum {
		log.Error("server %d new blockSyncReq startblkNum %d vs %d",
			self.cfg.accountStr, req.startBlockNum, self.nextReqBlkNum)
	}
	if req.targetBlockNum <= self.targetBlkNum {
		return nil
	}
	if self.nextReqBlkNum == 1 {
		self.nextReqBlkNum = req.startBlockNum
	}
	self.targetBlkNum = req.targetBlockNum
	peers := req.targetPeers
	if len(peers) == 0 {
		for p := range self.peers {
			peers = append(peers, p)
		}
	}

	for _, peerIdx := range req.targetPeers {
		if p, present := self.peers[peerIdx]; !present || !p.active {
			nextBlkNum := self.nextReqBlkNum
			if p != nil && p.nextReqBlkNum > nextBlkNum {
				log.Info("server %v, syncer with peer %d start from %d, vs %d",
					self.cfg.accountStr, peerIdx, p.nextReqBlkNum, self.nextReqBlkNum)
				nextBlkNum = p.nextReqBlkNum
			}
			self.peers[peerIdx] = &PeerSyncer{
				peerIdx:       peerIdx,
				nextReqBlkNum: nextBlkNum,
				targetBlkNum:  self.targetBlkNum,
				active:        false,
				//server:        self.server,
				stat:          self.stat,
				msgC:          make(chan comm.ConsensusMsg, 4),
			}
		}
		p := self.peers[peerIdx]
		self.startPeerSyncer(p, self.targetBlkNum)
	}

	return nil
}

func (self *PeerSyncer) run() {
	return
}

func (self *PeerSyncer) stop(force bool) bool {
	self.lock.Lock()
	defer self.lock.Unlock()
	if force || self.nextReqBlkNum > self.targetBlkNum {
		self.active = false
		return true
	}

	return false
}
