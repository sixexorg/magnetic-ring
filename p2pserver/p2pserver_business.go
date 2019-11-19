package p2pserver

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/bactor"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/p2pserver/common"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"

	// table
	"github.com/sixexorg/magnetic-ring/p2pserver/discover"
	// org
	"github.com/sixexorg/magnetic-ring/p2pserver/sync/org"
)

const (
	CHECKCONN_NUM          = 60 // Check the connection at 60 seconds
	ANODE_GETSTELLAR_TIMES = 10
)

// org info
type OrgMan struct {
	sync.RWMutex
	OrgSync map[comm.Address]*org.OrgBlockSyncMgr
}

func NewAddOrg() *OrgMan {
	ser := new(OrgMan)
	ser.OrgSync = make(map[comm.Address]*org.OrgBlockSyncMgr)

	return ser
}

func (this *OrgMan) Reset() {
	this.OrgSync = make(map[comm.Address]*org.OrgBlockSyncMgr)
}

// add org; from acror
func (this *P2PServer) AddOrg(orgID comm.Address) {
	return
	this.orgman.Lock()
	defer this.orgman.Unlock()
	log.Info("====>>>>P2PServer AddOrg", "orgID", orgID.ToString())

	if _, ok := this.orgman.OrgSync[orgID]; ok {
		return
	}
	if common.StellarNodeID != orgID {
		if len(this.orgman.OrgSync) > common.ORG_CONNECT_NUM {
			fmt.Println(" **** error org connection has reached the limit orgID:", orgID)
			return
		}
	}
	//fmt.Println("==========>>>>>>>>>>>>>>>P2PServer AddOr:", orgID.ToString())
	if this.orgConPeer(orgID) == false {
		// sleep 1s to slowdown
		time.Sleep(time.Second)
		go this.AddOrg(orgID)
		log.Warn("P2PServer AddOrg orgConPeer false", "orgID", orgID.ToString())
		return
	}
	log.Info("P2PServer AddOrg get orgids success", "orgID", orgID.ToString())

	this.network.PeerAddOrg(orgID)
	orgsync := org.NewOrgBlockSyncMgr(this, orgID)
	if common.StellarNodeID != orgID {
		go orgsync.Start()
	}
	this.orgman.OrgSync[orgID] = orgsync
	// add org and StellarNode need call PingTo
	np := this.network.GetNp()
	np.RLock()
	peers := make([]*peer.Peer, 0)
	for _, p := range np.List {
		if p.GetSyncState() == common.ESTABLISH {
			peers = append(peers, p)
		}
	}
	np.RUnlock()
	go this.PingTo(peers, false, orgID)
	//go this.orgConTimer(orgID)

	// }
	fmt.Println("🌐  📨  add league ", orgID.ToString())
}

// Timing is sent to the bootnode at regular intervals, and the bootnode is updated.
// The stellar node is also a special circle.
func (this *P2PServer) orgConTimer(orgID comm.Address) {
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			this.orgman.Lock()
			_, ok := this.orgman.OrgSync[orgID]
			this.orgman.Unlock()

			if ok {
				errNode, errs := this.ntab.SendBootNodeOrg(this.bootnodes, orgID, this.network.GetID(), true)
				log.Info("[P2PServer] orgConTimer", "orgID", orgID, ",errNode", errNode, ",errs", errs)
			} else {
				return
			}
		}
	}
}

// 1. Implement sending circle information to bootnode
// 2. Get the information in the circle
// 3. Filter the connected nodes and their own nodes
// 4. Connect the node
func (this *P2PServer) orgConPeer(orgID comm.Address) bool {
	log.Info("====>>>>P2PServer orgConPeer", "orgID", orgID.ToString())
	this.sendOrgToBN(orgID)
	//fmt.Println(" ***** 333333 ****** ")
	mapid := this.getOrgsFormBN(orgID)
	//fmt.Println("$$$$$$$$$$$$$$$$$===========>>>>>>>>>>>>", orgID.ToString(), len(mapid))
	if len(mapid) > 0 {
		this.orgConnect(orgID, mapid, false)
		return true
	}
	return false
}

func (this *P2PServer) sendOrgToBN(orgID comm.Address) {
	errNode, errs := this.ntab.SendBootNodeOrg(this.bootnodes, orgID, this.network.GetID(), true)
	log.Info("[P2PServer] sendOrgToBN", "orgID:", orgID.ToString(), ",errNode:", errNode, ",errs:", errs)
}

func (this *P2PServer) getOrgsFormBN(orgID comm.Address) map[discover.NodeID]*discover.Node {
	mapid := this.ntab.GetOrgNodesFromBN(this.bootnodes, orgID)
	log.Debug(" ********* getOrgsFormBN mapid:", "mapid", mapid)
	return mapid
}

func (this *P2PServer) orgConnect(orgID comm.Address, mapid map[discover.NodeID]*discover.Node, banode bool) {
	// connectnodes := make([]*discover.Node,0)
	np := this.network.GetNp()
	np.Lock()
	for _, p := range np.List {
		if p.GetSyncState() == common.ESTABLISH {
			if _, ok := mapid[p.GetNode().ID]; ok {
				delete(mapid, p.GetNode().ID)
			}
		}
	}
	np.Unlock()

	delete(mapid, this.ntab.Self().ID)
	//fmt.Println(" *********** mapid:", len(mapid), ",mapid:", mapid)

	for _, connectnode := range mapid {
		addr := &net.TCPAddr{IP: connectnode.IP, Port: int(connectnode.TCP)}
		this.network.Connect(addr.String(), false, connectnode, banode, orgID)
	}
	// fmt.Println(" ******* getRandList len(seedNodes):",len(seedNodes),"nodes:",seedNodes)
}

// del org; from acror
func (this *P2PServer) DelOrg(orgID comm.Address) {
	this.orgman.Lock()
	defer this.orgman.Unlock()

	// Update the status of the circle in this peer when the node deletes the circle.
	this.network.PeerDelOrg(orgID)

	if orgsync, ok := this.orgman.OrgSync[orgID]; ok {
		orgsync.Close()
		delete(this.orgman.OrgSync, orgID)
		log.Info(" [P2PServer] DelOrg", "orgID", orgID)
		go this.orgDelTimer(orgID)
		// When the node deletes a circle, the circle information of the remote node is also deleted.
		this.delOrgFromNP(orgID)

	} else {
		fmt.Println(" P2PServer DelOrg 记录下这个异常的东西 ")
	}
}

// Whenever you delete the circle information of the remote node,
// you need to send a disconnect message like bootnode.
func (this *P2PServer) delOrgFromNP(orgID comm.Address) {
	peerids := this.network.RemoteDelOrg(orgID)
	log.Info(" ****** P2PServer delOrgFromNP", "peerids", peerids)
	if len(peerids) > 0 {
		// Note: Because the star node does not maintain connection information,
		// So you only need to send the connection information that the far end becomes a non-stellar node.
		// I become a non-stellar node and send it directly out of the star node.
		if orgID != common.StellarNodeID {
			for _, peerid := range peerids {
				this.SentDisconnectToBootNode(peerid, false)
			}
		}
	}
}

// When the node exits the circle, it must report to the bootnode.
func (this *P2PServer) orgDelTimer(orgID comm.Address) {
	errNode, errs := this.ntab.SendBootNodeOrg(this.bootnodes, orgID, this.network.GetID(), false)
	log.Info("[P2PServer] orgDelTimer ", "orgID", orgID, ",errNode", errNode, ",errs", errs)
}

// The real join operation is done in sync
func (this *P2PServer) OrgAddNode(id uint64, orgID comm.Address) {
	this.orgman.Lock()
	defer this.orgman.Unlock()

	if orgsync, ok := this.orgman.OrgSync[orgID]; ok {
		orgsync.OnAddNode(id)
	} else {
		fmt.Println(" P2PServer OrgAddNode ")
	}
}

// The real delete operation is done in sync
func (this *P2PServer) OrgDelNode(id uint64, orgID comm.Address) {
	this.orgman.Lock()
	defer this.orgman.Unlock()

	if orgsync, ok := this.orgman.OrgSync[orgID]; ok {
		// fmt.Println(" ****** P2PServer OrgDelNode id:",id,",orgID:",orgID)
		orgsync.OnDelNode(id)
	}
	/* else { //
		fmt.Println(" P2PServer OrgDelNode  ")
	}*/
}

func (this *P2PServer) OrgHeaderReceive(fromID uint64, headers []*orgtypes.Header, orgID comm.Address) {
	this.orgman.Lock()
	defer this.orgman.Unlock()

	if orgsync, ok := this.orgman.OrgSync[orgID]; ok {
		orgsync.OnHeaderReceive(fromID, headers)
	} else {
		fmt.Println(" P2PServer OrgHeaderReceive  ")
	}
}

func (this *P2PServer) OrgBlockReceive(fromID uint64, blockSize uint32, block *orgtypes.Block, orgID comm.Address) {
	this.orgman.Lock()
	defer this.orgman.Unlock()

	if orgsync, ok := this.orgman.OrgSync[orgID]; ok {
		orgsync.OnBlockReceive(fromID, blockSize, block)
	} else {
		fmt.Println(" P2PServer OrgBlockReceive  ")
	}
}

// table connect callback
func (this *P2PServer) TabCallPeerConnectInfo() (uint64, []uint64) {
	result := make([]uint64, 0)
	// This node should not send connection information for orgs without a org.
	// BHaveOrgsExceptId the orgs without StellarNode
	if !this.network.BHaveOrgsExceptId(common.StellarNodeID) {
		return this.GetID(), result
	}

	np := this.network.GetNp()
	np.RLock()
	for _, tn := range np.List {
		// connected and have orgs without stellar info
		if tn.GetSyncState() == common.ESTABLISH && tn.BHaveOrgsExceptId(common.StellarNodeID) {
			result = append(result, tn.GetID())
		}
	}
	np.RUnlock()

	fmt.Println(" *8888888* P2PServer TabCallConnect ", "this.GetID()", this.GetID(), ",result", result)
	return this.GetID(), result
}

// table org callback whithout stellar info
func (this *P2PServer) TabCallPeerOrgInfo() (uint64, []comm.Address) {
	//fmt.Println(" *8888888* P2PServer TabCallOrg", "this.GetID()", this.GetID(), "PeerGetOrg", this.network.PeerGetOrg())
	orgs := make([]comm.Address, 0)
	for _, orgid := range this.network.PeerGetOrg() {
		if orgid != common.StellarNodeID { // del stellar
			orgs = append(orgs, orgid)
		}
	}
	return this.GetID(), orgs
}

// connect callback
func (this *P2PServer) TabCallSendConnectOrg(destnode *discover.Node, orgid comm.Address) {
	addr := &net.TCPAddr{IP: destnode.IP, Port: int(destnode.TCP)}
	go this.network.Connect(addr.String(), false, destnode, false)
}

// stellarnode add,merge with org
func (this *P2PServer) StellarNodeAdd() {
	log.Info(" ***** [P2PServer] StellarNodeAdd start ...")
	//
	this.nodeStar.AddStar()
	log.Info("====>>>>P2PServer StellarNodeAdd", "StellarNodeID", common.StellarNodeID.ToString())
	this.AddOrg(common.StellarNodeID)
}

// stellarnode del,merge with org
func (this *P2PServer) StellarNodeDel() {
	log.Info(" ***** [P2PServer] StellarNodeDel start ...")
	//
	this.nodeStar.DelStar()
	this.DelOrg(common.StellarNodeID)
}

// RemoteAddStellarNode remote add Stellar info
func (this *P2PServer) RemoteAddStellarNode(id uint64) {
	log.Info(" ***** [P2PServer] RemoteAddStellarNode start ...")
	// This should send connection information like bootnode
	// Now bn does not maintain the connection information of the stellar node, so there is no need to send node information to bn
	// this.SentConnectToBootNode(id)
	// Update NP information is the same as org logic in sync_handler.go
}

// RemoteDelStellarNode remote del Stellar info
func (this *P2PServer) RemoteDelStellarNode(id uint64) {
	log.Info(" ***** [P2PServer] RemoteDelStellarNode start ...")
	// Update NP information; org's del logic is done in org_block_sync.go;
	// This should send disconnection information like bootnode
	this.SentDisconnectToBootNode(id, true)

}

// org connect to bootnode
func (this *P2PServer) SentConnectToBootNode(remoteid uint64) {
	go func() {
		errNode, errs := this.ntab.SendBootNodeConnect(this.bootnodes, this.network.GetID(), remoteid, true, false)
		log.Info("[P2PServer] SentConnectToBootNode", "errNode", this.bootnodes,
			"ownid", this.network.GetID(), "remoteid", remoteid, "errNode", errNode, "errs", errs)
	}()
}

// org disconnect to bootnode
func (this *P2PServer) SentDisconnectToBootNode(remoteid uint64, bstellar bool) {
	go func() {
		errNode, errs := this.ntab.SendBootNodeConnect(this.bootnodes, this.network.GetID(), remoteid, false, bstellar)
		log.Info("[P2PServer] SentDisconnectToBootNode", "errNode", this.bootnodes,
			"ownid", this.network.GetID(), "remoteid", remoteid, "errNode", errNode, "errs", errs)
	}()
}


func (this *P2PServer) AddANode(orgid comm.Address) bool {
	// add anode
	this.nodeA.AddOrgOfA(orgid)

	fmt.Println("🌐  add node ⭕", orgid.ToString())
	// add org
	this.AddOrg(orgid)
	// get StellarNode
	/*	mapid := this.getOrgsFormBN(common.StellarNodeID)
		for j := 0; len(mapid) <= 0 && j < ANODE_GETSTELLAR_TIMES; j++ {
			log.Warn("P2PServer AddANode getOrgsFormBN StellarNodeID is null", "times", j)
			time.Sleep(time.Second)
			fmt.Println("🌐  add node ⭕  getOrgsFormBN", j)
			mapid = this.getOrgsFormBN(common.StellarNodeID)

		}*/

	//temp by rennbon 2019-01-26
	np := this.network.GetNp()
	mapid := make(map[discover.NodeID]*discover.Node)
	for _, v := range np.List {
		mapid[v.GetNode().ID] = v.GetNode()
	}
	//temp end

	if len(mapid) <= 0 {
		return false
	}
	fmt.Println("🌐  add node ⭕  mapid len:", len(mapid))
	stellarnodes := make(map[discover.NodeID]*discover.Node)
	stellarnodeid := make(map[discover.NodeID]bool)
	i := 0
	for nodeid, node := range mapid {
		stellarnodes[nodeid] = node
		stellarnodeid[nodeid] = true
		i++
		/*if i >= common.ANODE_CONNECT_STELLARNODE_NUM {
			break
		}*/
	}
	/*for k, v := range stellarnodeid {
		fmt.Println("🌐  add node ⭕  ", k.String(), v)
	}*/
	// This is connected to the star node, so let the other party know that it is the A node.
	// connet to StellarNode
	this.orgConnect(orgid, stellarnodes, true)
	// The A node knows which stellar nodes are connected, but the stellar nodes don't know which A nodes are connected.
	// Here to maintain a link, the message dialogue between the A node and the star node
	go this.anodeGetPeerID(stellarnodeid)

	//fmt.Println("🌐  add node ⭕  ️", orgid.ToString(), " ✅")
	return true

}

func (this *P2PServer) anodeGetPeerID(stellarnodeid map[discover.NodeID]bool) {
	peernodeidm := make(map[uint64]discover.NodeID)
	//fmt.Println("🌐  📝   anode add peer map len from：", len(stellarnodeid))
	for i := 0; i < 5; i++ {
		remainnodeid, addpeerid := this.anodeCheck(stellarnodeid)
		//fmt.Printf("🌐  📝  addpeerid len:%d ;remainnodeid len:%d \n", len(addpeerid), len(remainnodeid))
		if len(addpeerid) > 0 {
			for peerid, nodeid := range addpeerid {
				peernodeidm[peerid] = nodeid
			}
		}
		if len(remainnodeid) <= 0 {
			break
		}
		stellarnodeid = remainnodeid
		time.Sleep(time.Second)
	}
	//fmt.Println("🌐  📝   anode add peer map len real：", len(peernodeidm))
	this.nodeA.AddPeers(peernodeidm)
}

func (this *P2PServer) anodeCheck(stellarnodeid map[discover.NodeID]bool) (map[discover.NodeID]bool, map[uint64]discover.NodeID) {
	delnodeid := make([]discover.NodeID, 0)
	addpeerid := make(map[uint64]discover.NodeID)

	np := this.network.GetNp()
	np.RLock()
	for _, p := range np.List {
		//fmt.Printf("☢️  key:%d ,value:%s\n", k, p.GetNode().ID.String())
		if _, ok := stellarnodeid[p.GetNode().ID]; ok {
			delnodeid = append(delnodeid, p.GetNode().ID)
			addpeerid[p.GetID()] = p.GetNode().ID
		}
	}
	np.RUnlock()
	for _, nodeid := range delnodeid {
		delete(stellarnodeid, nodeid)
	}
	return stellarnodeid, addpeerid
}

func (this *P2PServer) HandlerANodeConn(idm map[discover.NodeID]bool) {
	return
	count := common.ANODE_CONNECT_STELLARNODE_NUM - len(idm)
	if count <= 0 {
		return
	}

	// get StellarNode
	fmt.Println(" ***** 111111 ****** ")
	mapid := this.getOrgsFormBN(common.StellarNodeID)
	stellarnodes := make(map[discover.NodeID]*discover.Node)
	stellarnodeid := make(map[discover.NodeID]bool)
	i := 0
	for nodeid, node := range mapid {
		if _, ok := idm[nodeid]; !ok {
			stellarnodes[nodeid] = node
			stellarnodeid[nodeid] = true
			i++
			if i >= count {
				break
			}
		}
	}

	// This is connected to the star node, so let the other party know that it is the A node.
	// connet to StellarNode
	// TODO: The first parameter is first set to empty here.
	this.orgConnect(comm.Address{}, stellarnodes, true)
	go this.anodeGetPeerID(stellarnodeid)
}

func (this *P2PServer) DelANode(orgid comm.Address) {
	this.nodeA.DelOrgOfA(orgid)
	this.DelOrg(orgid)
}

func (this *P2PServer) BANode() bool {
	return this.nodeA.BANode()
}

func (this *P2PServer) StarConnedNodeAInfo(id uint64) {
	this.nodeStar.AddPeers(id)
}

func (this *P2PServer) HandlerPingPongSpc(msg *common.PingPongSpcHandler) {
	if msg.BNodeASrc {
		this.nodeA.OnReceive(msg.PeerID, msg.BOtherSideReq)
	} else { // TODO: 这个要在恒星节点里面去处理
		this.nodeStar.OnReceive(msg.PeerID, msg.BOtherSideReq)
	}
}

// Timing to check connection information and strategy for handling disconnections
func (this *P2PServer) checkConn() {
	t := time.NewTicker(time.Second * CHECKCONN_NUM)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			this.analysConn()
		case <-this.quitCheckConn:
			return
		}
	}
}

type disAnalyOrgInfo struct {
	connect   map[uint64]bool
	predelnum int //
}

func (this *P2PServer) analysConn() {
	peersCount, orgPeerM, stellarNodeM, ordinaryNodeM := this.getRemoteConnOrg()
	// Compare if the connection is exceeded
	totalconn := config.GlobalConfig.P2PCfg.MaxConnInBound + config.GlobalConfig.P2PCfg.MaxConnOutBound
	if uint(peersCount) <= totalconn { // The number of connections has not exceeded the limit
		return
	}
	// Record the log, indicating that the connection exceeds the upper limit
	log.Info("P2PServer analysConn the connection of the node has exceeded the upper limit", "peersCount", peersCount)
	// Obtain the information of the A node maintained by the stellar node
	// and the information of the stellar node maintained by the A node.
	spcConnM := this.getSpecConn()

	// Filter some connections
	orgPeerM = this.filterConnection(orgPeerM, stellarNodeM, spcConnM)

	// Analysis of disconnection strategy
	this.analysDisconn(orgPeerM, ordinaryNodeM, peersCount-int(totalconn))
}

func (this *P2PServer) getSpecConn() map[uint64]bool {
	specConn := make(map[uint64]bool)
	for _, peerid := range this.nodeA.GetPeers() {
		specConn[peerid] = true
	}

	for _, peerid := range this.nodeStar.GetPeers() {
		specConn[peerid] = true
	}
	return specConn
}

func (this *P2PServer) getRemoteConnOrg() (int, map[comm.Address]*disAnalyOrgInfo, map[uint64]bool, map[uint64]bool) {
	peersCount := 0 // The total number of nodes
	//Pro node circle information maintained by this node
	orgpeerm := make(map[comm.Address]*disAnalyOrgInfo)
	//Pro node star node information maintained by this node
	stellarnodem := make(map[uint64]bool)
	//Pro node general node information maintained by this node
	ordinarynodem := make(map[uint64]bool)

	np := this.network.GetNp()
	np.RLock()
	for _, p := range np.List {
		if p.GetSyncState() != common.ESTABLISH {
			continue
		}
		peersCount++
		orgs := p.GetRemoteOrgs()
		if len(orgs) <= 0 {
			ordinarynodem[p.GetID()] = true
			continue
		}

		for _, orgid := range orgs {
			if orgid == common.StellarNodeID {
				stellarnodem[p.GetID()] = true
				continue
			}

			if _, ok := orgpeerm[orgid]; !ok {
				orgpeerm[orgid] = &disAnalyOrgInfo{
					connect: make(map[uint64]bool),
				}
			}
			orgpeerm[orgid].connect[p.GetID()] = true
		}
	}
	np.RUnlock()
	return peersCount, orgpeerm, stellarnodem, ordinarynodem
}

// 1. If the number of connections in the circle is less than or equal to 10, then the disconnection of this node is not checked.
// 2. If the connected node of the circle is also a star node or a special connection, then the disconnection of the node is not checked.
func (this *P2PServer) filterConnection(orgpeerm map[comm.Address]*disAnalyOrgInfo,
	stellarnodem, spcConnM map[uint64]bool) map[comm.Address]*disAnalyOrgInfo {

	delorgs := make([]comm.Address, 0)
	for orgid, orginfo := range orgpeerm {
		// No to ten connections are not deleted
		if len(orginfo.connect) <= common.ORG_CONNECT_ORG_NUM {
			delorgs = append(delorgs, orgid)
			continue
		}
		// num
		orginfo.predelnum = len(orginfo.connect) - common.ORG_CONNECT_ORG_NUM

		delpeers := make([]uint64, 0)
		for peerid, _ := range orginfo.connect {
			if _, ok := stellarnodem[peerid]; ok {
				delpeers = append(delpeers, peerid)
			}
			if _, ok := spcConnM[peerid]; ok {
				delpeers = append(delpeers, peerid)
			}
		}
		for _, peerid := range delpeers {
			delete(orgpeerm[orgid].connect, peerid)
		}
		if len(orgpeerm[orgid].connect) <= 0 {
			delorgs = append(delorgs, orgid)
		}
		delpeers = make([]uint64, 0)
	}
	for _, orgid := range delorgs {
		delete(orgpeerm, orgid)
	}
	return orgpeerm
}

func (this *P2PServer) analysDisconn(orgpeerm map[comm.Address]*disAnalyOrgInfo,
	ordinarynodem map[uint64]bool, predelcount int) {

	// handler org
	disConnPeerM := make(map[comm.Address]*disAnalyOrgInfo)
	delOrgCount := 0
	for {
		temorgpeerm, disconninfo, orgid := this.analysOrg(orgpeerm)
		count := len(disconninfo.connect)
		if count > disconninfo.predelnum {
			count = disconninfo.predelnum
		}
		if count > 0 {
			delOrgCount = delOrgCount + count
			i := 0
			disanalyinfo := new(disAnalyOrgInfo)
			disanalyinfo.connect = make(map[uint64]bool)
			for id, _ := range disconninfo.connect {
				disanalyinfo.connect[id] = true
				i++
				if i >= count {
					break
				}
			}
			disConnPeerM[orgid] = disanalyinfo
		}

		if len(temorgpeerm) <= 0 {
			break
		}
		orgpeerm = temorgpeerm
	}
	// send channel to del the org conns
	for orgid, info := range disConnPeerM {
		for peerid, _ := range info.connect {
			this.network.HaltOrgPeerID(peerid, orgid)
		}
	}

	// handler ordinarynode
	// if delOrgCount >= predelcount {
	// 	return
	// }

	if delOrgCount+2*common.ORG_CONNECT_ORG_NUM >= predelcount {
		return
	}

	if len(ordinarynodem) <= 2*common.ORG_CONNECT_ORG_NUM {
		return
	}
	preDelOrdinary := len(ordinarynodem) - 2*common.ORG_CONNECT_ORG_NUM
	disOrdinaryConn := make(map[uint64]bool)
	i := 0
	for peerid, _ := range ordinarynodem {
		disOrdinaryConn[peerid] = true
		i++
		if i >= preDelOrdinary {
			break
		}
	}
	// send channel to del the ordinaryrg conns
	// disOrdinaryConn
	for peerid, _ := range disOrdinaryConn {
		this.network.HaltOrdinaryPeerID(peerid)
	}
}

func (this *P2PServer) analysOrg(orgpeerm map[comm.Address]*disAnalyOrgInfo) (map[comm.Address]*disAnalyOrgInfo, *disAnalyOrgInfo, comm.Address) {
	// Get random org IDs and connection information for orgs
	disSrcInfo := new(disAnalyOrgInfo)
	var orgid comm.Address
	for id, info := range orgpeerm {
		if info != nil {
			disSrcInfo = info
			orgid = id
			break
		}
	}

	// Delete random org
	delete(orgpeerm, orgid)
	if len(orgpeerm) <= 0 {
		return orgpeerm, disSrcInfo, orgid
	}

	delpeeridm := make(map[uint64]bool)
	delorgid := make(map[comm.Address]bool)
	for srcpeerid, _ := range disSrcInfo.connect {
		for orgid, disinfo := range orgpeerm {
			if _, ok := disinfo.connect[srcpeerid]; ok {
				delete(disinfo.connect, srcpeerid)
				if len(disinfo.connect) <= 0 {
					delorgid[orgid] = true
				}
				delpeeridm[srcpeerid] = true
			}
		}
	}
	for id, _ := range delpeeridm {
		delete(disSrcInfo.connect, id)
	}
	for id, _ := range delorgid {
		delete(orgpeerm, id)
	}
	return orgpeerm, disSrcInfo, orgid
}

// A node sends circle data to the star node
func (this *P2PServer) AToStellarPendingData(data *common.OrgPendingData) {
	log.Debug("[p2p]AToStellarPendingData block message")
	//fmt.Printf("🌐  AToStellarPendingData ⭕️  %s,blockHeight:%d, peer len %d\n", data.OrgId.ToString(), data.Block.Header.Height, len(this.nodeA.GetPeers()))
	msg := msgpack.NewBlock(nil, data.Block, data.OrgId, common.SYNC_DATA_A_TO_STELLAR)

	notifyp2pactor, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
	if err != nil {
		log.Error("ANodeSendToStellarPeingData GetActorPid err", "org", data.OrgId, "err", err)
		return
	}
	notifyp2pactor.Tell(data.Block)

	//fmt.Println("msg genis start point blockhash=", data.Block.Hash(), " orgid=", data.OrgId, " distlen=", len(this.nodeA.GetPeers()))

	for _, peerid := range this.nodeA.GetPeers() {
		//fmt.Println("🌐 AToStellarPendingData 1", peerid)
		peer := this.GetNode(peerid)
		if peer == nil {
			continue
		}
		//fmt.Println("🌐 AToStellarPendingData 2")
		if peer.GetSyncState() == common.ESTABLISH && peer.GetRelay(){
			err := this.Send(peer, msg, false)
			if err != nil {
				log.Error("p2pserer_business atostellar pending data error", "error", err)
			}
		}
	}
}

// Broadcast circle data between stellar nodes
func (this *P2PServer) StellarToStellarPendingData(data *common.OrgPendingData) {
	log.Debug("[p2p]StellarToStellarPendingData block message")
	msg := msgpack.NewBlock(nil, data.Block, data.OrgId, common.SYNC_DATA_STELLAR_TO_STELLAR)
	//fmt.Println("🌐 height:", data.Block.Header.Height, ", hash:", data.Block.Header.LeagueId.ToString())
	np := this.network.GetNp()
	for _, p := range np.List {
		if p == nil {
			continue
		}
		if p.GetSyncState() != common.ESTABLISH || !p.BHaveOrgID(common.StellarNodeID) {
			continue
		}
		//fmt.Println("🌐 orgBlock 🔆 to  🔆 ", p.GetID())
		err := this.Send(p, msg, false)
		if err != nil {
			fmt.Printf("send to peer error=%v\n", err)
		}
	}
}

// for test ...
func (this *P2PServer) testOrg() {
	// fmt.Println(" ***** P2PServer testOrg start ...")
	time.Sleep(5 * time.Second)
	// portstr := strconv.Itoa(int(config.GlobalConfig.P2PCfg.NodePort))

	addorg, err := comm.ToAddress("ct2b5xc5uyaJrSqZ866gVvM5QtXybprb9sy")
	if err == nil {
		this.AddOrg(addorg)
		fmt.Println(" ******* P2PServer testOrg add:", addorg)
	} else {
		fmt.Println(" ******* P2PServer testOrg err:", err)
	}

	//	test add org
	// if portstr == "22001" {
	// 	go func(){
	// 		fmt.Println(" ******* dis dis")
	// 		time.Sleep(140*time.Second)
	// 		this.DelOrg(add)
	// 	}()
	// }

	//	test add star
	// if portstr == "21001" {
	// 	this.pid.Tell(&common.StellarNodeConnInfo{
	// 		BAdd: true,
	// 	})
	// }

	// if portstr == "21001" {
	// 	go func(){
	// 		fmt.Println(" ******* dis dis")
	// 		time.Sleep(40*time.Second)
	// 		this.pid.Tell(&common.StellarNodeConnInfo{
	// 			BAdd: false,
	// 		})
	// 	}()
	// }

	//	test add ANode
	// if portstr == "22001" {
	// 	this.pid.Tell(&common.ANodeConnInfo{
	// 		OrgId: addorg,
	// 		BAdd: true,
	// 	})
	// }
}
