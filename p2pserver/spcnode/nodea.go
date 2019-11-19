package spcnode

import (
	"sync"
	"time"

	"fmt"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/p2pserver/discover"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
)

const (
	PingSpcInterval     = 6
	PingSpcTimeOut      = (PingSpcInterval * 3) / 2
	PingSpcTimeOutCount = 3
	StallerNum          = 6
)

type NodeInfo struct {
	reqtime  int64
	errcount int
	nodeID   discover.NodeID
}

type NodeA struct {
	ownedorgm        map[comm.Address]bool //
	connStellarNodes map[uint64]*NodeInfo
	sync.RWMutex
	quit   chan bool
	p2pser SpectNodeP2PSer
}

func NewNodeA(p2pser SpectNodeP2PSer) *NodeA {
	ser := new(NodeA)
	ser.ownedorgm = make(map[comm.Address]bool)
	ser.connStellarNodes = make(map[uint64]*NodeInfo)
	ser.quit = make(chan bool)
	ser.p2pser = p2pser

	return ser
}

func (this *NodeA) BANode() bool {
	this.RLock()
	defer this.RUnlock()
	return len(this.ownedorgm) > 0
}

func (this *NodeA) OnReceive(peerid uint64, bothersidereq bool) {
	this.RLock()
	defer this.RUnlock()

	if bothersidereq {
		if info, ok := this.connStellarNodes[peerid]; ok {
			this.connStellarNodes[peerid] = &NodeInfo{
				reqtime:  time.Now().Unix(),
				errcount: 0,
				nodeID:   info.nodeID,
			}
		}
	} else {
		delete(this.connStellarNodes, peerid)
	}
}

func (this *NodeA) Close() {
	this.quit <- true
}

func (this *NodeA) Start() {
	go this.heartBeatService()
}

func (this *NodeA) heartBeatService() {
	t := time.NewTicker(time.Second * PingSpcInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			this.pingstellar()
			this.timeout()
			this.checknum()
		case <-this.quit:
			break
		}
	}
}

func (this *NodeA) checknum() {
	this.RLock()
	defer this.RUnlock()

	if len(this.ownedorgm) <= 0 {
		return
	}
	return

	num := len(this.connStellarNodes)
	if num <= StallerNum {
		idm := make(map[discover.NodeID]bool)
		for _, info := range this.connStellarNodes {
			idm[info.nodeID] = true
		}
		this.p2pser.HandlerANodeConn(idm)
	}
}

func (this *NodeA) timeout() {
	this.RLock()
	defer this.RUnlock()

	errpeerid := make([]uint64, 0)
	for peerid, nodeinfo := range this.connStellarNodes {
		if time.Now().Unix()-nodeinfo.reqtime >= PingSpcTimeOut {
			nodeinfo.errcount++
		}

		if nodeinfo.errcount >= PingSpcTimeOutCount {
			errpeerid = append(errpeerid, peerid)
		}
	}
	for _, id := range errpeerid {
		delete(this.connStellarNodes, id)
	}
}

func (this *NodeA) pingstellar() {
	this.RLock()
	defer this.RUnlock()

	peernul := make([]uint64, 0)
	for peerid, info := range this.connStellarNodes {
		bnul := this.send(peerid)
		if bnul {
			peernul = append(peernul, peerid)
		} else {
			this.connStellarNodes[peerid] = &NodeInfo{
				reqtime:  time.Now().Unix(),
				errcount: 0,
				nodeID:   info.nodeID,
			}
		}
	}
	for _, id := range peernul {
		delete(this.connStellarNodes, id)
	}
}

func (this *NodeA) send(peerid uint64) bool {
	peer := this.p2pser.GetNode(peerid)
	if peer == nil {
		return true
	}
	pingspc := msgpack.NewPingSpcMsg(true)
	go this.p2pser.Send(peer, pingspc, false)
	return false
}

func (this *NodeA) AddOrgOfA(orgid comm.Address) {
	this.RLock()
	defer this.RUnlock()

	this.ownedorgm[orgid] = true
}

func (this *NodeA) DelOrgOfA(orgid comm.Address) {
	this.RLock()
	defer this.RUnlock()

	delete(this.ownedorgm, orgid)
	if len(this.ownedorgm) == 0 {
		this.delAllPeers()
	}
}

func (this *NodeA) GetPeers() []uint64 {
	this.RLock()
	defer this.RUnlock()

	peers := make([]uint64, 0)
	for peerid, _ := range this.connStellarNodes {
		peers = append(peers, peerid)
	}
	return peers
}

func (this *NodeA) AddPeers(peers map[uint64]discover.NodeID) {
	this.RLock()
	defer this.RUnlock()

	//fmt.Println(" ******* NodeA AddPeers peers:",peers)

	for peerid, nodeid := range peers {
		fmt.Println("🌐  📝   anode add peer ", peerid, nodeid.String())
		this.connStellarNodes[peerid] = &NodeInfo{
			reqtime:  time.Now().Unix(),
			errcount: 0,
			nodeID:   nodeid,
		}
	}
}

// Delete the org connected by the A node
// func (this *NodeA) DelPeers(ids []uint64) {
// 	this.RLock()
// 	defer this.RUnlock()

// 	for _,id := range ids {
// 		delete(this.connStellarNodes,id)
// 	}
// }

func (this *NodeA) delAllPeers() {
	this.connStellarNodes = make(map[uint64]*NodeInfo)
}
