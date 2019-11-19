package netserver

import (
	"crypto/ecdsa"
	"errors"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/log"
	ledger "github.com/sixexorg/magnetic-ring/store/mainchain/storages"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/p2pserver/common"
	"github.com/sixexorg/magnetic-ring/p2pserver/discover"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"github.com/sixexorg/magnetic-ring/p2pserver/net/protocol"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"
	"fmt"
)

type P2PSer interface {
	SentDisconnectToBootNode(remoteid uint64, bstellar bool)
	BANode() bool
}

//NewNetServer return the net object in p2p
func NewNetServer(p2pser P2PSer) p2p.P2P {
	n := &NetServer{
		SyncChan: make(chan *common.MsgPayload, common.CHAN_CAPABILITY),
		ConsChan: make(chan *common.MsgPayload, common.CHAN_CAPABILITY),
	}
	n.p2pser = p2pser

	n.PeerAddrMap.PeerSyncAddress = make(map[string]*peer.Peer)
	n.PeerAddrMap.PeerConsAddress = make(map[string]*peer.Peer)

	n.init()
	return n
}

//NetServer represent all the actions in net layer
type NetServer struct {
	base         peer.PeerCom
	synclistener net.Listener
	conslistener net.Listener
	SyncChan     chan *common.MsgPayload
	ConsChan     chan *common.MsgPayload
	ConnectingNodes
	PeerAddrMap
	Np            *peer.NbrPeers
	connectLock   sync.Mutex
	inConnRecord  InConnectionRecord
	outConnRecord OutConnectionRecord
	OwnAddress    string //network`s own address(ip : sync port),which get from version check
	//
	nodeKey   *ecdsa.PrivateKey
	ntable    *discover.Table
	bootnodes []*discover.Node
	p2pser    P2PSer
}

//InConnectionRecord include all addr connected
type InConnectionRecord struct {
	sync.RWMutex
	InConnectingAddrs []string
}

//OutConnectionRecord include all addr accepted
type OutConnectionRecord struct {
	sync.RWMutex
	OutConnectingAddrs []string
}

//ConnectingNodes include all addr in connecting state
type ConnectingNodes struct {
	sync.RWMutex
	ConnectingAddrs []string
}

//PeerAddrMap include all addr-peer list
type PeerAddrMap struct {
	sync.RWMutex
	PeerSyncAddress map[string]*peer.Peer
	PeerConsAddress map[string]*peer.Peer
}

//init initializes attribute of network server
func (this *NetServer) init() error {
	this.base.SetVersion(common.PROTOCOL_VERSION)

	// if config.DefConfig.Consensus.EnableConsensus {
	// 	this.base.SetServices(uint64(common.VERIFY_NODE))
	// } else {
	// this.base.SetServices(uint64(common.SERVICE_NODE))
	// }
	this.base.SetServices(uint64(common.SERVICE_NODE))

	if config.GlobalConfig.P2PCfg.NodePort == 0 {
		log.Error("[p2p]link port invalid")
		return errors.New("[p2p]invalid link port")
	}

	this.base.SetSyncPort(uint16(config.GlobalConfig.P2PCfg.NodePort))

	if config.GlobalConfig.P2PCfg.DualPortSupport {
		if config.GlobalConfig.P2PCfg.NodeConsensusPort == 0 {
			log.Error("[p2p]consensus port invalid")
			return errors.New("[p2p]invalid consensus port")
		}

		this.base.SetConsPort(uint16(config.GlobalConfig.P2PCfg.NodeConsensusPort))
	} else {
		this.base.SetConsPort(0)
	}

	this.base.SetRelay(true)

	rand.Seed(time.Now().UnixNano())
	id := rand.Uint64()

	this.base.SetID(id)

	log.Info("[p2p]init peer ", "ID", this.base.GetID())
	this.Np = &peer.NbrPeers{}
	this.Np.Init()

	return nil
}

//InitListen start listening on the config port
func (this *NetServer) Start() {
	this.startListening()
}

//GetVersion return self peer`s version
func (this *NetServer) GetVersion() uint32 {
	return this.base.GetVersion()
}

//GetId return peer`s id
func (this *NetServer) GetID() uint64 {
	return this.base.GetID()
}

// SetHeight sets the local's height
func (this *NetServer) SetHeight(height uint64) {
	this.base.SetHeight(height)
}

// GetHeight return peer's heigh
func (this *NetServer) GetHeight() uint64 {
	return this.base.GetHeight()
}

// SetHeight sets the local's height
func (this *NetServer) SetOwnOrgHeight(height uint64, id comm.Address) {
	this.base.SetOwnOrgHeight(height, id)
}

// GetHeight return peer's heigh
func (this *NetServer) GetOwnOrgHeight(id comm.Address) uint64 {
	return this.base.GetOwnOrgHeight(id)
}

// PeerAddOrg handle main peer add org ownnode
func (this *NetServer) PeerAddOrg(id comm.Address) {
	this.base.PeerAddOrg(id)
}

// PeerDelOrg handle main peer del org ownnode
func (this *NetServer) PeerDelOrg(id comm.Address) {
	this.base.PeerDelOrg(id)
}

// PeerGetOrg main peer's orgs ownnode
func (this *NetServer) PeerGetOrg() []comm.Address {
	return this.base.PeerGetOrg()
}

func (this *NetServer) PeerGetRealOrg() []comm.Address {
	return this.base.PeerGetRealOrg()
}

// PeerGetOrg main peer's orgs ownnode
func (this *NetServer) RemoteDelOrg(id comm.Address) []uint64 {
	return this.GetNp().DelNbrOrg(id)
}

// PeerGetOrg main peer's orgs ownnode
func (this *NetServer) BHaveOrgs() bool {
	return this.base.BHaveOrgs()
}

func (this *NetServer) BHaveOrgsExceptId(id comm.Address) bool {
	return this.base.BHaveOrgsExceptId(id)
}

func (this *NetServer) BHaveOrgsId(id comm.Address) bool {
	return this.base.BHaveOrgsId(id)
}

//GetTime return the last contact time of self peer
func (this *NetServer) GetTime() int64 {
	t := time.Now()
	return t.UnixNano()
}

//GetServices return the service state of self peer
func (this *NetServer) GetServices() uint64 {
	return this.base.GetServices()
}

//GetSyncPort return the sync port
func (this *NetServer) GetSyncPort() uint16 {
	return this.base.GetSyncPort()
}

//GetConsPort return the cons port
func (this *NetServer) GetConsPort() uint16 {
	return this.base.GetConsPort()
}

//GetHttpInfoPort return the port support info via http
func (this *NetServer) GetHttpInfoPort() uint16 {
	return this.base.GetHttpInfoPort()
}

//GetRelay return whether net module can relay msg
func (this *NetServer) GetRelay() bool {
	return this.base.GetRelay()
}

// GetPeer returns a peer with the peer id
func (this *NetServer) GetPeer(id uint64) *peer.Peer {
	return this.Np.GetPeer(id)
}

//return nbr peers collection
func (this *NetServer) GetNp() *peer.NbrPeers {
	return this.Np
}

//GetNeighborAddrs return all the nbr peer`s addr
func (this *NetServer) GetNeighborAddrs() []common.PeerAddr {
	return this.Np.GetNeighborAddrs()
}

//GetConnectionCnt return the total number of valid connections
func (this *NetServer) GetConnectionCnt() uint32 {
	return this.Np.GetNbrNodeCnt()
}

//AddNbrNode add peer to nbr peer list
func (this *NetServer) AddNbrNode(remotePeer *peer.Peer) {
	this.Np.AddNbrNode(remotePeer)
}

//DelNbrNode delete nbr peer by id
func (this *NetServer) DelNbrNode(id uint64) (*peer.Peer, bool) {
	return this.Np.DelNbrNode(id)
}

//GetNeighbors return all nbr peer
func (this *NetServer) GetNeighbors() []*peer.Peer {
	return this.Np.GetNeighbors()
}

//NodeEstablished return whether a peer is establish with self according to id
func (this *NetServer) NodeEstablished(id uint64) bool {
	return this.Np.NodeEstablished(id)
}

//Xmit called by actor, broadcast msg
func (this *NetServer) Xmit(msg common.Message, isCons bool) {
	fmt.Println("NetServerXmit----------->>>>>>>>>>>", msg.CmdType(), time.Now())
	log.Info("NetServerXmit----------->>>>>>>>>>>", "cmdtype", msg.CmdType(), "time:",time.Now())
	this.Np.Broadcast(msg, isCons)
}

//GetMsgChan return sync or consensus channel when msgrouter need msg input
func (this *NetServer) GetMsgChan(isConsensus bool) chan *common.MsgPayload {
	if isConsensus {
		return this.ConsChan
	} else {
		return this.SyncChan
	}
}

//Tx send data buf to peer
func (this *NetServer) Send(p *peer.Peer, msg common.Message, isConsensus bool) error {
	if p != nil {
		if config.GlobalConfig.P2PCfg.DualPortSupport == false {
			return p.Send(msg, false)
		}
		return p.Send(msg, isConsensus)
	}
	log.Warn("[p2p]send to a invalid peer")
	return errors.New("[p2p]send to a invalid peer")
}

//IsPeerEstablished return the establise state of given peer`s id
func (this *NetServer) IsPeerEstablished(p *peer.Peer) bool {
	if p != nil {
		return this.Np.NodeEstablished(p.GetID())
	}
	return false
}

//Connect used to connect net address under sync or cons mode
func (this *NetServer) Connect(addr string, isConsensus bool, node *discover.Node, bANode bool, orgids ...comm.Address) error {
	if this.IsAddrInOutConnRecord(addr) {
		log.Debug("[p2p]", "Address", addr, "Consensus", isConsensus)
		return nil
	}
	if this.IsOwnAddress(addr) {
		return nil
	}
	if !this.AddrValid(addr) {
		return nil
	}

	if len(orgids) <= 0 && bANode == false {
		this.connectLock.Lock()
		connCount := uint(this.GetOutConnRecordLen())
		if connCount >= config.GlobalConfig.P2PCfg.MaxConnOutBound {
			log.Warn("[p2p]Connect: out ", "connections", connCount, "reach the max limit", config.GlobalConfig.P2PCfg.MaxConnOutBound)
			this.connectLock.Unlock()
			return errors.New("[p2p]connect: out connections reach the max limit")
		}
		this.connectLock.Unlock()
	}

	if this.IsNbrPeerAddr(addr, isConsensus) {
		return nil
	}
	this.connectLock.Lock()
	if added := this.AddOutConnectingList(addr); added == false {
		log.Debug("[p2p]node exist in connecting list", "addr", addr)
	}
	this.connectLock.Unlock()

	isTls := config.GlobalConfig.P2PCfg.IsTLS
	// fmt.Println(" ***** 888888888 888888 ***** isTls:",isTls)
	var rlpconn *conn
	var conn net.Conn
	var err error
	var remotePeer *peer.Peer
	if isTls {
		conn, err = TLSDial(addr)
		if err != nil {
			this.RemoveFromConnectingList(addr)
			log.Debug("[p2p]", "connect", addr, "failed", err.Error())
			return err
		}
	} else {
		conn, err = nonTLSDial(addr)
		if err != nil {
			this.RemoveFromConnectingList(addr)
			log.Debug("[p2p]", "connect", addr, "failed", err.Error())
			return err
		}
	}

	addr = conn.RemoteAddr().String()
	log.Debug("[p2p]", "peer", conn.LocalAddr().String(), "connect with", conn.RemoteAddr().String(), "with", conn.RemoteAddr().Network())

	emptyaddr := comm.Address{}
	if len(orgids) > 0 {
		emptyaddr = orgids[0]
	}
	temrlpconn, _, _, errrlp := this.SetupConn(conn, node, &emptyaddr, bANode)
	if errrlp != nil {
		return errrlp
	}
	rlpconn = temrlpconn

	this.AddOutConnRecord(addr)
	remotePeer = peer.NewPeer()
	this.AddPeerSyncAddress(addr, remotePeer)
	remotePeer.SyncLink.SetAddr(addr)
	remotePeer.SyncLink.SetConn(conn)
	remotePeer.SyncLink.SetRLPConn(rlpconn)
	remotePeer.AttachSyncChan(this.SyncChan)
	go remotePeer.SyncLink.Rx()
	remotePeer.SetSyncState(common.HAND)

	version := msgpack.NewVersion(this, isConsensus, uint32(ledger.GetLedgerStore().GetCurrentBlockHeight()))
	err = remotePeer.Send(version, isConsensus)
	if err != nil {
		if !isConsensus {
			this.RemoveFromOutConnRecord(addr)
		}
		log.Warn("err", "err", err)
		return err
	}
	return nil
}

//Halt stop all net layer logic
func (this *NetServer) Halt() {
	peers := this.Np.GetNeighbors()
	for _, p := range peers {
		p.CloseSync()
		p.CloseCons()
	}
	if this.synclistener != nil {
		this.synclistener.Close()
	}
	if this.conslistener != nil {
		this.conslistener.Close()
	}
}

//Halt the ordinary peer
func (this *NetServer) HaltOrdinaryPeerID(id uint64) {
	peer := this.GetPeer(id)
	if peer == nil {
		return
	}
	if peer.BHaveOrgs() {
		return
	}
	peer.CloseSync()
	peer.CloseCons()

}

//Halt the org peer
func (this *NetServer) HaltOrgPeerID(peerid uint64, orgid comm.Address) {
	peer := this.GetPeer(peerid)
	if peer == nil {
		return
	}
	if peer.BHaveOrgID(orgid) {
		peer.CloseSync()
		peer.CloseCons()
	}
}

//establishing the connection to remote peers and listening for inbound peers
func (this *NetServer) startListening() error {
	var err error

	syncPort := this.base.GetSyncPort()
	//consPort := this.base.GetConsPort()
	//fmt.Println(" ********** syncPort:",syncPort,"consPort",consPort)

	if syncPort == 0 {
		log.Error("[p2p]sync port invalid")
		return errors.New("[p2p]sync port invalid")
	}

	err = this.startSyncListening(syncPort)
	if err != nil {
		log.Error("[p2p]start sync listening fail")
		return err
	}

	return nil
}

// startSyncListening starts a sync listener on the port for the inbound peer
func (this *NetServer) startSyncListening(port uint16) error {
	var err error
	this.synclistener, err = createListener(port)
	if err != nil {
		log.Error("[p2p]failed to create sync listener")
		return errors.New("[p2p]failed to create sync listener")
	}

	go this.startSyncAccept(this.synclistener)
	log.Info("[p2p]start listen on sync ", "port", port)
	return nil
}

//startSyncAccept accepts the sync connection from the inbound peer
func (this *NetServer) startSyncAccept(listener net.Listener) {
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Error("[p2p]error accepting ", "err", err.Error())
			return
		}

		log.Debug("[p2p]remote sync node connect with ", "RemoteAdd", conn.RemoteAddr(), "LocalAddr", conn.LocalAddr())
		if !this.AddrValid(conn.RemoteAddr().String()) {
			log.Warn("[p2p]remote not in reserved list, close it ", "Addr", conn.RemoteAddr())
			conn.Close()
			continue
		}

		if this.IsAddrInInConnRecord(conn.RemoteAddr().String()) {
			conn.Close()
			continue
		}

		rlpconn, bownorg, banode, errrlp := this.SetupConn(conn, nil, nil, false)
		if errrlp != nil {
			continue
		}
		if bownorg == false && banode == false {
			syncAddrCount := uint(this.GetInConnRecordLen())
			if syncAddrCount >= config.GlobalConfig.P2PCfg.MaxConnInBound {
				log.Warn("[p2p]SyncAccept: total connections", "syncAddrCount", syncAddrCount, "reach the max limit", config.GlobalConfig.P2PCfg.MaxConnInBound)
				conn.Close()
				continue
			}

			remoteIp, err := common.ParseIPAddr(conn.RemoteAddr().String())
			if err != nil {
				log.Warn("[p2p]parse ip ", "error ", err.Error())
				conn.Close()
				continue
			}
			connNum := this.GetIpCountInInConnRecord(remoteIp)
			if connNum >= config.GlobalConfig.P2PCfg.MaxConnInBoundForSingleIP {
				log.Warn("[p2p]SyncAccept: connections", "connNum", connNum,
					"with ip", remoteIp, "has reach the max limit", config.GlobalConfig.P2PCfg.MaxConnInBoundForSingleIP)
				conn.Close()
				continue
			}
		}

		remotePeer := peer.NewPeer()
		addr := conn.RemoteAddr().String()
		this.AddInConnRecord(addr)

		this.AddPeerSyncAddress(addr, remotePeer)

		remotePeer.SyncLink.SetAddr(addr)
		remotePeer.SyncLink.SetConn(conn)
		remotePeer.SyncLink.SetRLPConn(rlpconn)

		remotePeer.AttachSyncChan(this.SyncChan)
		go remotePeer.SyncLink.Rx()
	}
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) AddOutConnectingList(addr string) (added bool) {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	for _, a := range this.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return false
		}
	}
	log.Trace("[p2p]add to out connecting list", "addr", addr)
	this.ConnectingAddrs = append(this.ConnectingAddrs, addr)
	return true
}

//Remove the peer from connecting list if the connection is established
func (this *NetServer) RemoveFromConnectingList(addr string) {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	addrs := this.ConnectingAddrs[:0]
	for _, a := range this.ConnectingAddrs {
		if a != addr {
			addrs = append(addrs, a)
		}
	}
	log.Trace("[p2p]remove from out connecting list", "addr", addr)
	this.ConnectingAddrs = addrs
}

//record the peer which is going to be dialed and sent version message but not in establish state
func (this *NetServer) GetOutConnectingListLen() (count uint) {
	this.ConnectingNodes.RLock()
	defer this.ConnectingNodes.RUnlock()
	return uint(len(this.ConnectingAddrs))
}

//check  peer from connecting list
func (this *NetServer) IsAddrFromConnecting(addr string) bool {
	this.ConnectingNodes.Lock()
	defer this.ConnectingNodes.Unlock()
	for _, a := range this.ConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return true
		}
	}
	return false
}

//find exist peer from addr map
func (this *NetServer) GetPeerFromAddr(addr string) *peer.Peer {
	var p *peer.Peer
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()

	p, ok := this.PeerSyncAddress[addr]
	if ok {
		return p
	}
	p, ok = this.PeerConsAddress[addr]
	if ok {
		return p
	}
	return nil
}

//IsNbrPeerAddr return result whether the address is under connecting
func (this *NetServer) IsNbrPeerAddr(addr string, isConsensus bool) bool {
	var addrNew string
	this.Np.RLock()
	defer this.Np.RUnlock()
	for _, p := range this.Np.List {
		if p.GetSyncState() == common.HAND || p.GetSyncState() == common.HAND_SHAKE ||
			p.GetSyncState() == common.ESTABLISH {
			if isConsensus {
				addrNew = p.ConsLink.GetAddr()
			} else {
				addrNew = p.SyncLink.GetAddr()
			}
			if strings.Compare(addrNew, addr) == 0 {
				return true
			}
		}
	}
	return false
}

//AddPeerSyncAddress add sync addr to peer-addr map
func (this *NetServer) AddPeerSyncAddress(addr string, p *peer.Peer) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	log.Debug("[p2p]AddPeerSyncAddress", "addr", addr)
	this.PeerSyncAddress[addr] = p
}

//AddPeerConsAddress add cons addr to peer-addr map
func (this *NetServer) AddPeerConsAddress(addr string, p *peer.Peer) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	log.Debug("[p2p]AddPeerConsAddress", "addr", addr)
	this.PeerConsAddress[addr] = p
}

//RemovePeerSyncAddress remove sync addr from peer-addr map
func (this *NetServer) RemovePeerSyncAddress(addr string) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	if _, ok := this.PeerSyncAddress[addr]; ok {
		delete(this.PeerSyncAddress, addr)
		log.Debug("[p2p]delete Sync Address", "addr", addr)
	}
}

//RemovePeerConsAddress remove cons addr from peer-addr map
func (this *NetServer) RemovePeerConsAddress(addr string) {
	this.PeerAddrMap.Lock()
	defer this.PeerAddrMap.Unlock()
	if _, ok := this.PeerConsAddress[addr]; ok {
		delete(this.PeerConsAddress, addr)
		log.Debug("[p2p]delete Cons Address", "addr", addr)
	}
}

//GetPeerSyncAddressCount return length of cons addr from peer-addr map
func (this *NetServer) GetPeerSyncAddressCount() (count uint) {
	this.PeerAddrMap.RLock()
	defer this.PeerAddrMap.RUnlock()
	return uint(len(this.PeerSyncAddress))
}

//AddInConnRecord add in connection to inConnRecord
func (this *NetServer) AddInConnRecord(addr string) {
	this.inConnRecord.Lock()
	defer this.inConnRecord.Unlock()
	for _, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return
		}
	}
	this.inConnRecord.InConnectingAddrs = append(this.inConnRecord.InConnectingAddrs, addr)
	log.Debug("[p2p]add in record", "addr", addr)
}

//IsAddrInInConnRecord return result whether addr is in inConnRecordList
func (this *NetServer) IsAddrInInConnRecord(addr string) bool {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	for _, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return true
		}
	}
	return false
}

//IsIPInInConnRecord return result whether the IP is in inConnRecordList
func (this *NetServer) IsIPInInConnRecord(ip string) bool {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	var ipRecord string
	for _, addr := range this.inConnRecord.InConnectingAddrs {
		ipRecord, _ = common.ParseIPAddr(addr)
		if 0 == strings.Compare(ipRecord, ip) {
			return true
		}
	}
	return false
}

//RemoveInConnRecord remove in connection from inConnRecordList
func (this *NetServer) RemoveFromInConnRecord(addr string) {
	this.inConnRecord.Lock()
	defer this.inConnRecord.Unlock()
	addrs := []string{}
	for _, a := range this.inConnRecord.InConnectingAddrs {
		if strings.Compare(a, addr) != 0 {
			addrs = append(addrs, a)
		}
	}
	log.Debug("[p2p]remove in record", "addr", addr)
	this.inConnRecord.InConnectingAddrs = addrs
}

//GetInConnRecordLen return length of inConnRecordList
func (this *NetServer) GetInConnRecordLen() int {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	return len(this.inConnRecord.InConnectingAddrs)
}

//GetIpCountInInConnRecord return count of in connections with single ip
func (this *NetServer) GetIpCountInInConnRecord(ip string) uint {
	this.inConnRecord.RLock()
	defer this.inConnRecord.RUnlock()
	var count uint
	var ipRecord string
	for _, addr := range this.inConnRecord.InConnectingAddrs {
		ipRecord, _ = common.ParseIPAddr(addr)
		if 0 == strings.Compare(ipRecord, ip) {
			count++
		}
	}
	return count
}

//AddOutConnRecord add out connection to outConnRecord
func (this *NetServer) AddOutConnRecord(addr string) {
	this.outConnRecord.Lock()
	defer this.outConnRecord.Unlock()
	for _, a := range this.outConnRecord.OutConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return
		}
	}
	this.outConnRecord.OutConnectingAddrs = append(this.outConnRecord.OutConnectingAddrs, addr)
	log.Debug("[p2p]add out record", "addr", addr)
}

//IsAddrInOutConnRecord return result whether addr is in outConnRecord
func (this *NetServer) IsAddrInOutConnRecord(addr string) bool {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	for _, a := range this.outConnRecord.OutConnectingAddrs {
		if strings.Compare(a, addr) == 0 {
			return true
		}
	}
	return false
}

//RemoveOutConnRecord remove out connection from outConnRecord
func (this *NetServer) RemoveFromOutConnRecord(addr string) {
	this.outConnRecord.Lock()
	defer this.outConnRecord.Unlock()
	addrs := []string{}
	for _, a := range this.outConnRecord.OutConnectingAddrs {
		if strings.Compare(a, addr) != 0 {
			addrs = append(addrs, a)
		}
	}
	log.Debug("[p2p]remove out record", "addr", addr)
	this.outConnRecord.OutConnectingAddrs = addrs
}

//GetOutConnRecordLen return length of outConnRecord
func (this *NetServer) GetOutConnRecordLen() int {
	this.outConnRecord.RLock()
	defer this.outConnRecord.RUnlock()
	return len(this.outConnRecord.OutConnectingAddrs)
}

//AddrValid whether the addr could be connect or accept
func (this *NetServer) AddrValid(addr string) bool {
	if config.GlobalConfig.P2PCfg.ReservedPeersOnly && len(config.GlobalConfig.P2PCfg.ReservedCfg.ReservedPeers) > 0 {
		for _, ip := range config.GlobalConfig.P2PCfg.ReservedCfg.ReservedPeers {
			if strings.HasPrefix(addr, ip) {
				log.Info("[p2p]found reserved peer", "addr", addr)
				return true
			}
		}
		return false
	}
	return true
}

//check own network address
func (this *NetServer) IsOwnAddress(addr string) bool {
	if addr == this.OwnAddress {
		return true
	}
	return false
}

//Set own network address
func (this *NetServer) SetOwnAddress(addr string) {
	if addr != this.OwnAddress {
		log.Info("[p2p]set own ", "addr", addr)
		this.OwnAddress = addr
	}

}

func (this *NetServer) SetPrivateKey(nodeKey *ecdsa.PrivateKey, table *discover.Table) {
	this.nodeKey = nodeKey
	this.ntable = table
	//fmt.Println(" ***************** NetServer SetPrivateKey ")
	//this.ntable.Self().PrintNode()
}

func (this *NetServer) SetBootNodes(nodes []*discover.Node) {
	this.bootnodes = nodes
}

func (this *NetServer) GetNode() *discover.Node {
	return this.ntable.Self()
}

func (this *NetServer) SyncHandleSentDisconnectToBootNode(remoteid uint64, bstellar bool) {
	this.p2pser.SentDisconnectToBootNode(remoteid, bstellar)
}

func (this *NetServer) SyncHandleBANode() bool {
	return this.p2pser.BANode()
}

//
func (this *NetServer) SetupConn(fd net.Conn, dialDest *discover.Node,
	orgid *comm.Address, bANode bool) (*conn, bool, bool, error) {
	// Prevent leftover pending conns from entering the handshake.
	// srv.lock.Lock()
	// running := srv.running
	// srv.lock.Unlock()

	c := &conn{fd: fd, transport: newRLPX(fd)}

	// if !running {
	// 	c.close(errServerStopped)
	// 	return
	// }
	// Run the encryption handshake.
	var remodeOrgID comm.Address
	var err error
	bownorg := false
	bOppsANode := false

	if c.id, remodeOrgID, bOppsANode, err = c.doEncHandshake(this.nodeKey, dialDest, orgid, bANode); err != nil {
		//fmt.Println("Failed RLPx handshake", "addr", c.fd.RemoteAddr(), "conn", "err", err)
		c.Close(err)
		return nil, bownorg, bOppsANode, err
	}
	for _, ownorg := range this.PeerGetOrg() {
		if remodeOrgID == ownorg {
			bownorg = true
			//fmt.Println(" ********** the same org ok ok ok ok ok ok ok ok ... ")
		}
	}
	//if bOppsANode {
	//	fmt.Println(" ********** the opps is ANode ok ok ok ok ok ok ok ok ... ")
	//}

	//fmt.Println(" ******* c.id:",c.id,",dialDest:",dialDest,",remodeOrgID:",remodeOrgID,",bOppsANode:",bOppsANode)
	// clog := log.New("id", c.id, "addr", c.fd.RemoteAddr(), "conn", c.flags)
	// For dialed connections, check that the remote public key matches.
	if dialDest != nil && c.id != dialDest.ID {
		c.Close(common.DiscUnexpectedIdentity)
		//fmt.Println("Dialed identity mismatch", "want", c, dialDest.ID)
		return nil, bownorg, bOppsANode, errors.New("[p2p] c.id != dialDest.ID error")
	}
	// if err := srv.checkpoint(c, srv.posthandshake); err != nil {
	// 	fmt.Println("Rejected peer before protocol handshake", "err", err)
	// 	c.close(err)
	// 	return
	// }
	// Run the protocol handshake
	phs, err := c.doProtoHandshake(&protoHandshake{Version: baseProtocolVersion,
		ID: discover.PubkeyID(&this.nodeKey.PublicKey)})
	if err != nil {
		//fmt.Println("Failed proto handshake", "err", err)
		c.Close(err)
		return nil, bownorg, bOppsANode, err
	}
	//fmt.Println(" ****** phs.ID:",phs.ID)
	if phs.ID != c.id {
		//fmt.Println("Wrong devp2p handshake identity", "err", phs.ID)
		c.Close(common.DiscUnexpectedIdentity)
		return nil, bownorg, bOppsANode, errors.New("[p2p] phs.ID != c.id error")
	}
	return c, bownorg, bOppsANode, nil
}
