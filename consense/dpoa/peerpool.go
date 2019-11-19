package dpoa

import (
	"fmt"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/log"
	//"github.com/ontio/ontology-crypto/vrf"
	//"github.com/ontio/ontology-crypto/keypair"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	//"github.com/sixexorg/magnetic-ring/consensus/vbft/config"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto"
)



type Peer struct {
	Index          string
	PubKey         crypto.PublicKey
	handShake      *comm.PeerHandshakeMsg
	LatestInfo     *comm.PeerHeartbeatMsg // latest heartbeat msg
	LastUpdateTime time.Time              // time received heartbeat from peer
	connected      bool
}

type PeerPool struct {
	lock                   sync.RWMutex
	IDMap                  map[string]uint32
	P2pMap                 map[string]uint64 //value: p2p random id
	peers                  map[string]*Peer
	peerConnectionWaitings map[string]chan struct{}
}

func NewPeerPool() *PeerPool {
	return &PeerPool{
		IDMap:  make(map[string]uint32),
		P2pMap: make(map[string]uint64),
		peers:  make(map[string]*Peer),
		peerConnectionWaitings: make(map[string]chan struct{}),
	}
}
func (pool *PeerPool) Init(peers []string) error {
	for _, p := range peers {
		//pk, _ := vconfig.Pubkey(p)
		//if !vrf.ValidatePublicKey(pk) {
		//	return fmt.Errorf("peer %d: invalid peer pubkey for VRF", p)
		//}
		if err := pool.addPeer(p); err != nil {
			return fmt.Errorf("failed to add peer %d: %s", p, err)
		}
		log.Info("added peer: %s", p)
	}
	return nil
}

func (pool *PeerPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	//pool.configs = make(map[uint32]*vconfig.PeerConfig)
	pool.IDMap = make(map[string]uint32)
	pool.P2pMap = make(map[string]uint64)
	pool.peers = make(map[string]*Peer)
}

// FIXME: should rename to isPeerConnected
func (pool *PeerPool) isNewPeer(peerIdx string) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if _, present := pool.peers[peerIdx]; present {
		return !pool.peers[peerIdx].connected
	}

	return true
}

func (pool *PeerPool) addPeer(p string) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	buf, err := common.Hex2Bytes(p)
	if err != nil {
		return err
	}

	peerPK, err := crypto.UnmarshalPubkey(buf)
	//peerPK, err := vconfig.Pubkey(p)
	if err != nil {
		return fmt.Errorf("failed to unmarshal peer pubkey: %s", err)
	}

	pool.peers[p] = &Peer{
		Index:          p,
		PubKey:         peerPK,
		LastUpdateTime: time.Unix(0, 0),
		connected:      false,
	}
	return nil
}

func (pool *PeerPool) getActivePeerCount() int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	n := 0
	for _, p := range pool.peers {
		if p.connected {
			n++
		}
	}
	return n
}

func (pool *PeerPool) waitPeerConnected(peerIdx string) error {
	if !pool.isNewPeer(peerIdx) {
		// peer already connected
		return nil
	}

	var C chan struct{}
	pool.lock.Lock()
	if _, present := pool.peerConnectionWaitings[peerIdx]; !present {
		C = make(chan struct{})
		pool.peerConnectionWaitings[peerIdx] = C
	} else {
		C = pool.peerConnectionWaitings[peerIdx]
	}
	pool.lock.Unlock()

	<-C

	return nil
}

func (pool *PeerPool) peerConnected(peerIdx string) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	// new peer, rather than modify

	buf, err := common.Hex2Bytes(peerIdx)
	if err != nil {
		return err
	}

	peerPK, err := crypto.UnmarshalPubkey(buf)
	if err != nil {
		log.Error("PeerPool peerConnected", "peerIdx", peerIdx, "err", err)
		return err
	}
	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         peerPK,
		LastUpdateTime: time.Now(),
		connected:      true,
	}
	if C, present := pool.peerConnectionWaitings[peerIdx]; present {
		delete(pool.peerConnectionWaitings, peerIdx)
		close(C)
	}
	return nil
}

func (pool *PeerPool) peerDisconnected(peerIdx string) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	var lastUpdateTime time.Time
	if p, present := pool.peers[peerIdx]; present {
		lastUpdateTime = p.LastUpdateTime
	}

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		LastUpdateTime: lastUpdateTime,
		connected:      false,
	}
	return nil
}

func (pool *PeerPool) peerHandshake(peerIdx string, msg *comm.PeerHandshakeMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		handShake:      msg,
		LatestInfo:     pool.peers[peerIdx].LatestInfo,
		LastUpdateTime: time.Now(),
		connected:      true,
	}

	return nil
}

func (pool *PeerPool) peerHeartbeat(peerIdx string, msg *comm.PeerHeartbeatMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if C, present := pool.peerConnectionWaitings[peerIdx]; present {
		// wake up peer connection waitings
		delete(pool.peerConnectionWaitings, peerIdx)
		close(C)
	}

	pool.peers[peerIdx] = &Peer{
		Index:          peerIdx,
		PubKey:         pool.peers[peerIdx].PubKey,
		handShake:      pool.peers[peerIdx].handShake,
		LatestInfo:     msg,
		LastUpdateTime: time.Now(),
		connected:      true,
	}

	return nil
}

func (pool *PeerPool) getNeighbours() []*Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peers := make([]*Peer, 0)
	for _, p := range pool.peers {
		if p.connected {
			peers = append(peers, p)
		}
	}
	return peers
}

func (pool *PeerPool) GetPeerIndex(nodeId string) (uint32, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	idx, present := pool.IDMap[nodeId]
	return idx, present
}

func (pool *PeerPool) GetPeerPubKey(peerIdx string) crypto.PublicKey {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if p, present := pool.peers[peerIdx]; present && p != nil {
		return p.PubKey
	}

	return nil
}

func (pool *PeerPool) isPeerAlive(peerIdx string) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p := pool.peers[peerIdx]
	if p == nil || !p.connected {
		return false
	}

	// p2pserver keeps peer alive

	return true
}

func (pool *PeerPool) getPeer(idx string) *Peer {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	peer := pool.peers[idx]
	if peer != nil {
		if peer.PubKey == nil {

			buf, err := common.Hex2Bytes(idx)
			if err != nil {
				return nil
			}

			peer.PubKey, _ = crypto.UnmarshalPubkey(buf)
			//peer.PubKey, _ = vconfig.Pubkey(idx)
		}
		return peer
	}

	return nil
}

func (pool *PeerPool) addP2pId(peerIdx string, p2pId uint64) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.P2pMap[peerIdx] = p2pId
}

func (pool *PeerPool) getP2pId(peerIdx string) (uint64, bool) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	p2pid, present := pool.P2pMap[peerIdx]
	return p2pid, present
}

func (pool *PeerPool) RemovePeerIndex(nodeId string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	delete(pool.IDMap, nodeId)
}
