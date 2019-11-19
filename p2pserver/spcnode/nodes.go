package spcnode

import (
	"sync"
	"time"

	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
)

type NodeStar struct {
	sync.RWMutex
	quit chan bool
	p2pser SpectNodeP2PSer
	connANodes	map[uint64]*NodeInfo // 保存对端A节点的信息
	bStarNode bool
}

func NewNodeStar(p2pser SpectNodeP2PSer) *NodeStar {
	ser := new(NodeStar)
	ser.quit = make(chan bool)
	ser.p2pser = p2pser
	ser.connANodes = make(map[uint64]*NodeInfo)

	return ser
}

func (this *NodeStar) BStarNode() bool {
	return this.bStarNode
}

func (this *NodeStar) OnReceive(peerid uint64,bothersidereq bool) {
	//fmt.Println(" ******** NodeStar OnReceive......peerid:",peerid,",bothersidereq:",bothersidereq)
	this.RLock()
	defer this.RUnlock()

	if bothersidereq {
		if _,ok := this.connANodes[peerid]; ok {
			this.connANodes[peerid] = &NodeInfo{
				reqtime: time.Now().Unix(),
				errcount: 0,
			}
		}
	} else {
		delete(this.connANodes,peerid)
	}
}

func (this *NodeStar) Close() {
	this.quit <- true
}

func (this *NodeStar) Start() {
	go this.heartBeatService()
}

func (this *NodeStar) heartBeatService() {
	t := time.NewTicker(time.Second * PingSpcInterval)
	defer t.Stop()

	for {
		select {
		case <- t.C:
			this.pingANode()
			this.timeout()
		case <-this.quit:
			break
		}
	}
}

func (this *NodeStar) timeout() {
	this.RLock()
	defer this.RUnlock()

	errpeerid := make([]uint64,0)
	for peerid,nodeinfo := range this.connANodes {
		if time.Now().Unix() - nodeinfo.reqtime >= PingSpcTimeOut {
			nodeinfo.errcount++
		}

		if nodeinfo.errcount >= PingSpcTimeOutCount {
			errpeerid = append(errpeerid,peerid)
		}
	}
	for _,id := range errpeerid {
		delete(this.connANodes,id)
	}
}

func (this *NodeStar) pingANode() {
	this.RLock()
	defer this.RUnlock()

	peernul := make([]uint64,0)
	for peerid,info := range this.connANodes {
		bnul := this.send(peerid)
		if bnul {
			peernul = append(peernul,peerid)
		} else {
			this.connANodes[peerid] = &NodeInfo{
				reqtime: time.Now().Unix(),
				errcount: 0,
				nodeID: info.nodeID,
			}
		}
	}
	for _,id := range peernul {
		delete(this.connANodes,id)
	}
}

func (this *NodeStar) send(peerid uint64) bool {
	peer := this.p2pser.GetNode(peerid)
	if peer == nil {
		return true
	}
	pingspc := msgpack.NewPingSpcMsg(false)
	go this.p2pser.Send(peer, pingspc, false)
	return false
}

func (this *NodeStar) AddStar() {
	this.bStarNode = true
}

func (this *NodeStar) DelStar() {
	this.bStarNode = false
	this.delAllPeers()
}

func (this *NodeStar) GetPeers() []uint64 {
	this.RLock()
	defer this.RUnlock()

	peers := make([]uint64,0)
	for peerid,_ := range this.connANodes {
		peers = append(peers,peerid)
	}
	return peers
}

func (this *NodeStar) AddPeers(id uint64) {
	this.RLock()
	defer this.RUnlock()

	this.connANodes[id] = &NodeInfo{
		reqtime: time.Now().Unix(),
		errcount: 0,
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

func (this *NodeStar) delAllPeers() {
	this.connANodes = make(map[uint64]*NodeInfo)
}