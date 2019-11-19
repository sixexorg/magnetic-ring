package dpoa

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"

	//"github.com/sixexorg/magnetic-ring/core/signature"
	//"github.com/ontio/ontology-crypto/keypair"
	actorTypes "github.com/sixexorg/magnetic-ring/consense/actor"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	p2pmsg "github.com/sixexorg/magnetic-ring/p2pserver/common"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"time"
)



type TransAction struct {
	msgpool  *MsgPool
	peerpool *PeerPool
	stateMgr *StateMgr
	dpoaMgr  *DpoaMgr
	p2p      *actorTypes.P2PActor
	msgSendC chan *SendMsgEvent
	msgRecvC map[string]chan *comm.P2pMsgPayload
	cfg      *Config
	quitWg   sync.WaitGroup
	quitC    chan struct{}
}

func NewtransAction(msgpool *MsgPool, peerpool *PeerPool, stateMgr *StateMgr, dpoaMgr *DpoaMgr, p2p *actorTypes.P2PActor, cfg *Config) *TransAction {
	return &TransAction{msgpool: msgpool, peerpool: peerpool, stateMgr: stateMgr, dpoaMgr: dpoaMgr, p2p: p2p, cfg: cfg,
		msgRecvC: make(map[string]chan *comm.P2pMsgPayload), msgSendC: make(chan *SendMsgEvent, 16)}
}

func (t *TransAction) start() {
	go t.sendLoop()
	t.receiveLoop()
}

func (t *TransAction) sendLoop() {
	t.quitWg.Add(1)
	defer t.quitWg.Done()

	for {
		select {
		case evt := <-t.msgSendC:
			payload, err := SerializeVbftMsg(evt.Msg)
			if err != nil {
				log.Error("server %d failed to serialized msg (type: %d): %s", t.cfg.accountStr, evt.Msg.Type(), err)
				continue
			}
			//log.Info("func dpoa transaction sendLoop", "msg type", evt.Msg.Type())
			if evt.ToPeer == comm.BroadCast {
				// broadcast
				if err := t.broadcastToAll(payload); err != nil {
					log.Error("server %v xmit msg (type %d): %s",
						t.cfg.accountStr, evt.Msg.Type(), err)
				}
			} else {
				if err := t.sendToPeer(evt.ToPeer, payload); err != nil {
					log.Error("server %v xmit to peer %v failed: %s", t.cfg.accountStr, evt.ToPeer, err)
				}
			}
		case <-t.quitC:
			log.Info("server %d msgSendLoop quit", t.cfg.accountStr)
			return
		}
	}
}

func (t *TransAction) broadcastToAll(data []byte) error {
	msg := &p2pmsg.ConsensusPayload{
		Data:  data,
		Owner: t.cfg.account.PublicKey(),
	}

	buf := new(bytes.Buffer)
	if err := msg.SerializeUnsigned(buf); err != nil {
		return fmt.Errorf("failed to serialize consensus msg", "err", err)
	}
	var err error
	msg.Signature, err = t.cfg.account.Sign(common.Sha256(buf.Bytes()))
	if err != nil {
		log.Error("sign transaction transaction91 occured error","error",err)
		return err
	}
	t.p2p.Broadcast(msg)
	return nil
}

func (t *TransAction) sendToPeer(peerIdx string, data []byte) error {
	peer := t.peerpool.getPeer(peerIdx)
	if peer == nil {
		return fmt.Errorf("send peer failed: failed to get peer", "peerIdx", peerIdx)
	}
	msg := &p2pmsg.ConsensusPayload{
		Data:  data,
		Owner: t.cfg.account.PublicKey(),
	}

	buf := new(bytes.Buffer)
	if err := msg.SerializeUnsigned(buf); err != nil {
		return fmt.Errorf("failed to serialize consensus msg", "err", err)
	}
	var err error
	msg.Signature, err = t.cfg.account.Sign(common.Sha256(buf.Bytes()))

	if err != nil {
		log.Error("t.cfg.account.sign transaction116",)
		return err
	}

	cons := msgpack.NewConsensus(msg)
	p2pid, present := t.peerpool.getP2pId(peerIdx)
	if present {
		t.p2p.Transmit(p2pid, cons)
	} else {
		log.Error("sendToPeer transmit failed index:", "peerIdx", peerIdx)
	}
	return nil
}

func (t *TransAction) receiveLoop() {
	for _, p := range t.cfg.Peers {
		if _, present := t.msgRecvC[p]; !present {
			t.msgRecvC[string(p)] = make(chan *comm.P2pMsgPayload, 1024)
		}

		go func(pname string) {
			if err := t.run(pname); err != nil {
				log.Error("server %d, processor on peer failed: %s",
					"accountStr", t.cfg.accountStr, "err", err)
			}
		}(p)
	}
}
func (t *TransAction) receiveFromPeer(peerIdx string) (string, []byte, error) {
	if C, present := t.msgRecvC[peerIdx]; present {
		select {
		case payload := <-C:
			if payload != nil {
				return payload.FromPeer, payload.Payload.Data, nil
			}

		case <-t.quitC:
			return "", nil, fmt.Errorf("server %d quit", t.cfg.accountStr)
		}
	}

	return "", nil, fmt.Errorf("nil consensus payload")
}

func (t *TransAction) run(peerPubKey string) error {
	// broadcast heartbeat
	t.heartbeat()

	if err := t.peerpool.waitPeerConnected(peerPubKey); err != nil {
		return err
	}

	defer func() {
		// TODO: handle peer disconnection here
		log.Warn("server %d: disconnected with peer", "accountStr", t.cfg.accountStr)
		close(t.msgRecvC[peerPubKey])
		delete(t.msgRecvC, peerPubKey)

		t.peerpool.peerDisconnected(peerPubKey)
		t.stateMgr.StateEventC <- &StateEvent{
			Type: UpdatePeerState,
			peerState: &PeerState{
				peerIdx:   peerPubKey,
				connected: false,
			},
		}
	}()
	errC := make(chan error)
	go func() {
		for {
			fromPeer, msgData, err := t.receiveFromPeer(peerPubKey)
			if err != nil {
				errC <- err
				return
			}
			msg, err := DeserializeVbftMsg(msgData)

			if err != nil {
				log.Error("server %d failed to deserialize vbft msg (len %d): %s", "accountStr", t.cfg.accountStr, "len(msgData)", len(msgData), "err", err)
			} else {
				pk := t.peerpool.GetPeerPubKey(fromPeer)
				if pk == nil {
					log.Error("server %d failed to get peer %d pubkey", "accountStr", t.cfg.accountStr, "fromPeer", fromPeer)
					continue
				}

				if err := msg.Verify(pk); err != nil {
					log.Error("server %d failed to verify msg, type %d, err: %s",
						"accountStr", t.cfg.accountStr, "type", msg.Type(), "err", err)
					continue
				}

				t.handleRecvMsg(fromPeer, msg, hashData(msgData))
			}
		}
	}()

	return <-errC
}

func (t *TransAction) handleRecvMsg(peerIdx string, msg comm.ConsensusMsg, msgHash common.Hash) {
	if t.msgpool.HasMsg(msg, msgHash) {
		// dup msg checking
		log.Debug("dup msg with msg type %d from %d", "type", msg.Type(), "peeridx", peerIdx)
		return
	}

	switch msg.Type() {
	case comm.ConsenseDone:
		t.ConsenseDone(peerIdx, msg, msgHash)
	case comm.ConsensePrepare:
		t.ConsensePrepare(peerIdx, msg, msgHash)
	case comm.ConsensePromise:
		t.ConsensePromise(peerIdx, msg, msgHash)
	case comm.ConsenseProposer:
		t.ConsenseProposer(peerIdx, msg, msgHash)
	case comm.ConsenseAccept:
		t.ConsenseAccept(peerIdx, msg, msgHash)
	case comm.ViewTimeout:
		t.ViewTimeout(peerIdx, msg, msgHash)
	case comm.EvilEvent:

	case comm.EventFetchMessage:

	case comm.EventFetchRspMessage:

	case comm.EvilFetchMessage:

	case comm.EvilFetchRspMessage:

	case comm.EarthFetchRspSigs:
		t.EarthFetchRspSigs(peerIdx, msg, msgHash)
	case comm.PeerHeartbeatMessage:
		t.PeerHeartbeatMessage(peerIdx, msg, msgHash)
	case comm.BlockFetchMessage:

	case comm.BlockFetchRespMessage:

	case comm.BlockInfoFetchMessage:
		t.BlockInfoFetchMessage(peerIdx, msg, msgHash)
	case comm.BlockInfoFetchRespMessage:

	case comm.TimeoutFetchMessage:

	case comm.TimeoutFetchRspMessage:
	}
}
func (t *TransAction) GetCurrentBlockNo() uint64 {
	return t.dpoaMgr.store.db.GetCurrentBlockHeight() + 1
}

func (t *TransAction) GetCommittedBlockNo() uint64 {
	return t.dpoaMgr.store.getLatestBlockNumber()
}

func (t *TransAction) heartbeat() {
	fmt.Println("TransActionheartbeat=========>>>>>>>>>>",time.Now().String())
	log.Info("TransActionheartbeat=========>>>>>>>>>>", "time:", time.Now().String())
	//	build heartbeat msg
	msg, err := constructHeartbeatMsg(t.dpoaMgr.store)
	if err != nil {
		log.Error("failed to build heartbeat msg", "err", err)
		return
	}

	//	send to peer
	t.msgSendC <- &SendMsgEvent{
		ToPeer: comm.BroadCast,
		Msg:    msg,
	}
}

func (t *TransAction) processHeartbeatMsg(peerIdx string, msg *comm.PeerHeartbeatMsg) error {

	if err := t.peerpool.peerHeartbeat(peerIdx, msg); err != nil {
		return fmt.Errorf("failed to update peer %d: %s", peerIdx, err)
	}
	//log.Debug("server %v received heartbeat from peer %d, blkNum %d", peerIdx, msg.CommittedBlockNumber)
	t.stateMgr.StateEventC <- &StateEvent{
		Type: UpdatePeerState,
		peerState: &PeerState{
			peerIdx:   peerIdx,
			connected: true,
			//chainConfigView:   msg.ChainConfigView,
			committedBlockNum: msg.CommittedBlockNumber,
		},
	}

	return nil
}

func (t *TransAction) recvMsg(peer string, msg *comm.P2pMsgPayload) {
	if C, present := t.msgRecvC[peer]; present {
		C <- msg
	} else {
		log.Error("consensus msg without receiver node", "peerid", peer)
		return
	}
}

func (t *TransAction) sendMsg(data *SendMsgEvent) {
	t.msgSendC <- data
}
