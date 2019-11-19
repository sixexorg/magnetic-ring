package dpoa

import (
	//"bytes"
	"fmt"
	"math"
	"sync"

	"github.com/sixexorg/magnetic-ring/bactor"

	"github.com/sixexorg/magnetic-ring/node"
	//"time"
	//"encoding/json"

	//"github.com/ontio/ontology-crypto/vrf"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/account"

	//"github.com/sixexorg/magnetic-ring/common"
	actorTypes "github.com/sixexorg/magnetic-ring/consense/actor"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	"github.com/sixexorg/magnetic-ring/log"

	//"github.com/sixexorg/magnetic-ring/consense/vbft/config"
	//"github.com/sixexorg/magnetic-ring/core/ledger"
	//"github.com/sixexorg/magnetic-ring/core/types"
	//"github.com/sixexorg/magnetic-ring/events"
	"github.com/sixexorg/magnetic-ring/events/message"
	p2pmsg "github.com/sixexorg/magnetic-ring/p2pserver/common"

	//"github.com/ontio/ontology-crypto/keypair"
	//"github.com/sixexorg/magnetic-ring/core/signature"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"

	//msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	//"github.com/sixexorg/magnetic-ring/crypto"
	//"encoding/hex"
	p2pcomm "github.com/sixexorg/magnetic-ring/p2pserver/common"
	"sync/atomic"
)

type BftAction struct {
	Type     comm.BftActionType
	BlockNum uint64
	forEmpty bool
}

type Server struct {
	//poolActor	*actorTypes.TxPoolActor
	p2p *actorTypes.P2PActor
	//ledger		*ledger.Ledger
	ledger     *storages.LedgerStoreImp
	pid        *actor.PID
	account    account.NormalAccount
	accountStr string
	store      *BlockStore
	stateMgr   *StateMgr
	dpoaMgr    *DpoaMgr
	peerpool   *PeerPool
	trans      *TransAction
	timer      *EventTimer
	syncer     *Syncer
	msgpool    *MsgPool
	bftActionC chan *BftAction
	proc       *ProcMode
	//sub         *events.ActorSubscriber
	rsvCount uint64

	quitC       chan struct{}
	stopNewHtCh chan struct{}
	quitWg      sync.WaitGroup
	notifyBlock *Feed
	notifyState *Feed
}

func NewServer(account account.NormalAccount, p2p *actor.PID) (*Server, error) {
	server := &Server{
		//poolActor:          &actorTypes.TxPoolActor{Pool: txpool},
		p2p:         &actorTypes.P2PActor{P2P: p2p},
		ledger:      storages.GetLedgerStore(),
		account:     account,
		notifyBlock: &Feed{},
		notifyState: &Feed{},
		accountStr:  fmt.Sprintf("%x", account.PublicKey().Bytes()),
		quitC:       make(chan struct{}, 0),
	}

	props := actor.FromProducer(func() actor.Actor {
		return server
	})

	pid, err := actor.SpawnNamed(props, "consensus_dpoa")
	if err != nil {
		return nil, err
	}
	server.pid = pid
	bactor.RegistActorPid(bactor.CONSENSUSACTOR, pid)
	//server.sub = events.NewActorSubscriber(pid)

	return server, nil
}

func (self *Server) LoadChainConfig(chainStore *BlockStore) error {
	//self.config = &comm.ChainConfig{}
	for _, n := range node.GurStars() {
		if self.accountStr != n {
			DftConfig.Peers = append(DftConfig.Peers, n)
		}
	}
	DftConfig.Peers = append(DftConfig.Peers, node.CurEarth())
	//if self.accountStr != "04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f" {
	//	DftConfig.Peers = append(DftConfig.Peers, "04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f")
	//}
	//if self.accountStr != "044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6" {
	//	DftConfig.Peers = append(DftConfig.Peers, "044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6")
	//}
	//if self.accountStr != "0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f" {
	//	DftConfig.Peers = append(DftConfig.Peers, "0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f")
	//}
	//if self.accountStr != "04d2db562f13d94fd31d5d500152cac0bfd1692b9fc1185f2fbea712dbd34f7e6c65ce05303ee3a4ce772e0513c75e95a3f3dcc97ea45e22cfebbe3a658de4a493" {
	//	DftConfig.Peers = append(DftConfig.Peers, "04d2db562f13d94fd31d5d500152cac0bfd1692b9fc1185f2fbea712dbd34f7e6c65ce05303ee3a4ce772e0513c75e95a3f3dcc97ea45e22cfebbe3a658de4a493")
	//}
	//if self.accountStr != "029acf7eeb3faa596ce8915a8a9b5ac00717388401ba6b0e330aa37230988ff0ef" {
	//	DftConfig.Peers = append(DftConfig.Peers, "029acf7eeb3faa596ce8915a8a9b5ac00717388401ba6b0e330aa37230988ff0ef")
	//}
	//
	//DftConfig.Peers = append(DftConfig.Peers, "04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f")
	//DftConfig.Peers = append(DftConfig.Peers, "044a147deabaa89e15aab6f586ce8c9b68bf11043ca83387b125d82489252d94e858f7b43a43000d7a949ff1bd8742fcad57434d08b2455bfc5f681dd0cf0a32f6")
	//DftConfig.Peers = append(DftConfig.Peers, "0405269acdc54c24220f67911a3ac709b129ff2454717875b789288954bb2e6afaed4d59ff0f4c4d14afff9f6f4e1ffb71a1e5457fa2ca7440a020218559ab7f3f")
	DftConfig.account = self.account
	DftConfig.accountStr = self.accountStr

	return nil
}

func (self *Server) Start() error {
	return self.start()
}

func (self *Server) Halt() error {
	return nil
}

func (self *Server) GetPID() *actor.PID {
	return self.pid
}

func (self *Server) actionLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case action := <-self.bftActionC:
			switch action.Type {
			case comm.FastForward:
			case comm.ReBroadcast:
				blkNum := self.GetCurrentBlockNo()
				if blkNum > action.BlockNum {
					continue
				}
				self.trans.heartbeat()
			}
		case <-self.quitC:
			log.Info("server %d actionLoop quit", self.accountStr)
			return
		}
	}
}

func (self *Server) initialize() error {
	var err error
	if self.store, err = NewBlockStore(self.ledger, self.accountStr, self.p2p.P2P); err != nil {
		return err
	}

	if err := self.LoadChainConfig(self.store); err != nil {
		log.Error("failed to load config: %s", err)
		return fmt.Errorf("failed to load config: %s", err)
	}
	log.Info("chain config loaded from local", "current blockNum:", self.store.db.GetCurrentBlockHeight()+1)

	self.peerpool = NewPeerPool()
	self.msgpool = newMsgPool(self.rsvCount)
	self.dpoaMgr = NewdpoaMgr(DftConfig, self.store, self.msgpool, self.p2p.P2P)
	self.stateMgr = NewStateMgr(self.store, self.notifyState, DftConfig)
	self.trans = NewtransAction(self.msgpool, self.peerpool, self.stateMgr, self.dpoaMgr, self.p2p, DftConfig)
	self.proc = NewprocMode(self.dpoaMgr, self.msgpool, self.trans, self.notifyBlock, self.notifyState, DftConfig)
	self.timer = NewEventTimer()
	self.syncer = newSyncer(self.stateMgr, DftConfig)
	self.bftActionC = make(chan *BftAction, 8)
	//self.sub.Subscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)
	self.peerpool.Init(DftConfig.Peers)
	self.syncer.run()
	self.stateMgr.run()
	self.trans.start()
	go self.timerLoop()
	go self.actionLoop()
	self.proc.Start()
	self.stateMgr.StateEventC <- &StateEvent{
		Type: ConfigLoaded,
	}
	log.Info("peer started", "PublicKey:", self.accountStr)
	// TODO: start peer-conn-handlers
	aa := p2pcomm.StellarNodeConnInfo{BAdd: true}

	p2ppid, err := bactor.GetActorPid(bactor.P2PACTOR)

	if err != nil {
		log.Error("p2pactor not init", "err", err)
		return err
	}
	p2ppid.Tell(&aa)
	//self.p2p.P2P.Tell(&aa)
	return nil
}

func (self *Server) start() error {
	if err := self.initialize(); err != nil {
		return fmt.Errorf("dpoa server start failed: %s", err)
	}

	self.timer.startPeerTicker(math.MaxUint32)
	self.proc.Process()

	return nil
}

func (self *Server) stop() error {
	//self.sub.Unsubscribe(message.TOPIC_SAVE_BLOCK_COMPLETE)

	close(self.quitC)
	self.quitWg.Wait()

	self.syncer.stop()
	self.timer.stop()
	self.msgpool.clean()
	self.store.close()
	self.peerpool.clean()

	return nil
}

func (srv *Server) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *actor.Restarting:
		log.Info("dpoa actor restarting")
	case *actor.Stopping:
		log.Info("dpoa actor stopping")
	case *actor.Stopped:
		log.Info("dpoa actor stopped")
	case *actor.Started:
		log.Info("dpoa actor started")
	case *actor.Restart:
		log.Info("dpoa actor restart")
	case *actorTypes.StartConsensus:
		log.Info("dpoa actor start consensus")
	case *actorTypes.StopConsensus:
		srv.stop()
	case *message.SaveBlockCompleteMsg:
		log.Info("dpoa actor receives block complete event. block height=%d, numtx=%d",
			msg.Block.Header.Height, len(msg.Block.Transactions))
		srv.handleBlockPersistCompleted(msg.Block)
	case *p2pmsg.ConsensusPayload:
		srv.NewConsensusPayload(msg)
	case *types.Block:
		srv.handleBlockPersistCompleted(msg)
	case *p2pmsg.PeerAnn:
		srv.handleAnn(msg)

	default:
		//log.Info("vbft actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *Server) handleAnn(ann *p2pcomm.PeerAnn) {
	var peerid string = ""
	for p, v := range self.peerpool.P2pMap {
		if v == ann.PeerId {
			peerid = p
		}
	}
	if len(peerid) == 0 {
		return
	}

	self.stateMgr.StateEventC <- &StateEvent{
		Type: UpdatePeerState,
		peerState: &PeerState{
			peerIdx:           peerid,
			connected:         true,
			committedBlockNum: ann.Height,
		},
	}
}

func (self *Server) handleBlockPersistCompleted(block *types.Block) {
	log.Info("1----------------|||------>>>>>>>>>>.persist block", "height", block.Header.Height, "block.Hash", block.Hash(), "timestamp", block.Header.Timestamp, "payload", len(block.Header.ConsensusPayload))
	fmt.Println("1----------------|||------>>>>>>>>>>.persist block", "height", block.Header.Height, "block.Hash", block.Hash(), "timestamp", block.Header.Timestamp, "payload", len(block.Header.ConsensusPayload))
	curHeight := atomic.LoadUint64(&self.store.curHeight)
	if !self.store.isEarth() && curHeight < block.Header.Height{
		fmt.Println("2----------------|||------>>>>>>>>>>", "curheight", atomic.LoadUint64(&self.store.curHeight), "height", block.Header.Height)
		atomic.AddUint64(&self.store.curHeight,1)
		if self.stateMgr.currentState == SyncReady {
			self.stateMgr.currentState = Syncing
			self.proc.state = Syncing
			self.stateMgr.setSyncedReady()
		}
	}
	self.store.onBlockSealed(block.Header.Height)
	self.timer.onBlockSealed(uint32(block.Header.Height))
	self.msgpool.onBlockSealed(block.Header.Height)
	self.notifyBlock.Send(*block)
	self.trans.heartbeat()

	if self.store.isEarth() {
		p := &p2pcomm.EarthNotifyBlk{
			BlkHeight: block.Header.Height,
			BlkHash:   block.Header.Hash(),
		}
		self.p2p.Broadcast(p)
	}
	pp := &p2pmsg.NotifyBlk{
		BlkHeight: block.Header.Height,
		BlkHash:   block.Header.Hash(),
	}
	self.p2p.Broadcast(pp)
}

func (srv *Server) NewConsensusPayload(payload *p2pmsg.ConsensusPayload) {
	peerID := fmt.Sprintf("%x", payload.Owner.Bytes())
	if srv.peerpool.isNewPeer(peerID) {
		srv.peerpool.peerConnected(peerID)
	}
	p2pid, present := srv.peerpool.getP2pId(string(peerID))
	if !present || p2pid != payload.PeerId {
		srv.peerpool.addP2pId(peerID, payload.PeerId)
	}

	msg := &comm.P2pMsgPayload{
		FromPeer: peerID,
		Payload:  payload,
	}
	srv.trans.recvMsg(peerID, msg)
}

func (self *Server) getState() ServerState {
	return self.stateMgr.getState()
}

func (self *Server) timerLoop() {
	self.quitWg.Add(1)
	defer self.quitWg.Done()

	for {
		select {
		case evt := <-self.timer.C:
			if err := self.processTimerEvent(evt); err != nil {
				log.Error("failed to process timer evt: %d, err: %s", evt.evtType, err)
			}

		case <-self.quitC:
			log.Info("server %d timerLoop quit", self.accountStr)
			return
		}
	}
}

func (self *Server) processTimerEvent(evt *TimerEvent) error {
	switch evt.evtType {
	case EventProposalBackoff:
	case EventProposeBlockTimeout:
	case EventRandomBackoff:
	case EventPropose2ndBlockTimeout:
	case EventEndorseBlockTimeout:
	case EventEndorseEmptyBlockTimeout:
	case EventCommitBlockTimeout:
	case EventPeerHeartbeat:
		self.trans.heartbeat()
	case EventTxPool:
	case EventTxBlockTimeout:
	}
	return nil
}

func (srv *Server) GetCommittedBlockNo() uint64 {
	return srv.store.getLatestBlockNumber()
}

//func (self *Server) restartSyncing() {
//	self.stateMgr.checkStartSyncing(self.GetCommittedBlockNo(), true)
//}

func (self *Server) GetCurrentBlockNo() uint64 {
	return self.store.db.GetCurrentBlockHeight() + 1
}

//func (self *Server) checkSyncing() {
//	self.stateMgr.checkStartSyncing(self.GetCommittedBlockNo(), false)
//}
