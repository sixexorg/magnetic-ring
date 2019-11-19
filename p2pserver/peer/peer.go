package peer

import (
	"errors"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/p2pserver/common"
	"github.com/sixexorg/magnetic-ring/p2pserver/discover"
	"github.com/sixexorg/magnetic-ring/p2pserver/sync/conn"
)

// PeerCom provides the basic information of a peer
type PeerCom struct {
	id           uint64
	version      uint32
	services     uint64
	relay        bool
	httpInfoPort uint16
	syncPort     uint16
	consPort     uint16
	height       uint64
	orgheight    map[comm.Address]uint64 // org info
	syncorg      sync.RWMutex
}

// SetID sets a peer's id
func (this *PeerCom) SetID(id uint64) {
	this.id = id
}

// GetID returns a peer's id
func (this *PeerCom) GetID() uint64 {
	return this.id
}

// SetVersion sets a peer's version
func (this *PeerCom) SetVersion(version uint32) {
	this.version = version
}

// GetVersion returns a peer's version
func (this *PeerCom) GetVersion() uint32 {
	return this.version
}

// SetServices sets a peer's services
func (this *PeerCom) SetServices(services uint64) {
	this.services = services
}

// GetServices returns a peer's services
func (this *PeerCom) GetServices() uint64 {
	return this.services
}

// SerRelay sets a peer's relay
func (this *PeerCom) SetRelay(relay bool) {
	this.relay = relay
}

// GetRelay returns a peer's relay
func (this *PeerCom) GetRelay() bool {
	return this.relay
}

// SetSyncPort sets a peer's sync port
func (this *PeerCom) SetSyncPort(port uint16) {
	this.syncPort = port
}

// GetSyncPort returns a peer's sync port
func (this *PeerCom) GetSyncPort() uint16 {
	return this.syncPort
}

// SetConsPort sets a peer's consensus port
func (this *PeerCom) SetConsPort(port uint16) {
	this.consPort = port
}

// GetConsPort returns a peer's consensus port
func (this *PeerCom) GetConsPort() uint16 {
	return this.consPort
}

// SetHttpInfoPort sets a peer's http info port
func (this *PeerCom) SetHttpInfoPort(port uint16) {
	this.httpInfoPort = port
}

// GetHttpInfoPort returns a peer's http info port
func (this *PeerCom) GetHttpInfoPort() uint16 {
	return this.httpInfoPort
}

// SetHeight sets a peer's height
func (this *PeerCom) SetHeight(height uint64) {
	this.height = height
}

// GetHeight returns a peer's height
func (this *PeerCom) GetHeight() uint64 {
	return this.height
}

// p2pserver call the PeerAddOrg fuc ownnode
func (this *PeerCom) PeerAddOrg(id comm.Address) {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}

	this.orgheight[id] = 0
}

// p2pserver call the PeerDelOrg fuc ownnode
func (this *PeerCom) PeerDelOrg(id comm.Address) {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}

	delete(this.orgheight, id)
	if len(this.orgheight) == 0 {
		this.orgheight = make(map[comm.Address]uint64)
	}
}

// p2pserver call the PeerGetOrg fuc ownnode
func (this *PeerCom) PeerGetOrg() []comm.Address {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}
	result := make([]comm.Address, 0)
	for id, _ := range this.orgheight {
		result = append(result, id)
	}

	return result
}

func (this *PeerCom) PeerGetRealOrg() []comm.Address {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}
	result := make([]comm.Address, 0)
	for id, _ := range this.orgheight {

		if !id.Equals(common.StellarNodeID) {
			result = append(result, id)
		}
	}

	return result
}

// SetOrgHeight sets a org peer's height : ownnode
func (this *PeerCom) SetOwnOrgHeight(height uint64, id comm.Address) {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}

	this.orgheight[id] = height
}

// GetOwnOrgHeight returns a own org peer's height : ownnode
func (this *PeerCom) GetOwnOrgHeight(id comm.Address) uint64 {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}

	return this.orgheight[id]
}

// Remote peer handler

// SetRemoteOrgHeight sets a org peer's height : remotenode
func (this *PeerCom) SetRemoteOrgHeight(height uint64, id comm.Address) bool {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}
	bnew := true
	if _, ok := this.orgheight[id]; ok {
		bnew = false
	}
	this.orgheight[id] = height
	return bnew
}

// GetRemoteOrgHeight returns a remote org peer's height : remotenode
func (this *PeerCom) GetRemoteOrgHeight(id comm.Address) uint64 {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}

	return this.orgheight[id]
}

func (this *PeerCom) GetRemoteOrgs() []comm.Address {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()

	result := make([]comm.Address, 0)
	for orgid, _ := range this.orgheight {
		result = append(result, orgid)
	}
	return result
}

func (this *PeerCom) DelRemoteOrg(id comm.Address) bool {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()

	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}

	bhave := false
	if _, ok := this.orgheight[id]; ok {
		bhave = true
	}
	delete(this.orgheight, id)
	if len(this.orgheight) == 0 {
		this.orgheight = make(map[comm.Address]uint64)
	}
	return bhave
}

func (this *PeerCom) BHaveOrgs() bool {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()

	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}

	return len(this.orgheight) > 0
}

func (this *PeerCom) BHaveOrgID(id comm.Address) bool {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()

	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}
	bhave := false
	if _, ok := this.orgheight[id]; ok {
		bhave = true
	}
	return bhave
}

func (this *PeerCom) BHaveOrgsExceptId(id comm.Address) bool {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}
	bhave := false
	if len(this.orgheight) > 1 {
		bhave = true
		return bhave
	} else if len(this.orgheight) == 1 {
		if _, ok := this.orgheight[id]; !ok {
			bhave = true
		}
	}

	return bhave
}

func (this *PeerCom) BHaveOrgsId(id comm.Address) bool {
	this.syncorg.Lock()
	defer this.syncorg.Unlock()
	if this.orgheight == nil {
		this.orgheight = make(map[comm.Address]uint64)
	}
	bhave := false
	if _, ok := this.orgheight[id]; ok {
		bhave = true
	}

	return bhave
}

//Peer represent the node in p2p
type Peer struct {
	base      PeerCom
	cap       [32]byte
	SyncLink  *conn.Link
	ConsLink  *conn.Link
	syncState uint32
	consState uint32
	txnCnt    uint64
	rxTxnCnt  uint64
	connLock  sync.RWMutex
	node      *discover.Node
}

//NewPeer return new peer without publickey initial
func NewPeer() *Peer {
	p := &Peer{
		syncState: common.INIT,
		consState: common.INIT,
	}
	p.SyncLink = conn.NewLink()
	p.ConsLink = conn.NewLink()
	p.base.orgheight = make(map[comm.Address]uint64)
	runtime.SetFinalizer(p, rmPeer)
	return p
}

//rmPeer print a debug log when peer be finalized by system
func rmPeer(p *Peer) {
	log.Debug("[p2p]Remove unused", "peerid", p.GetID())
}

//DumpInfo print all information of peer
func (this *Peer) DumpInfo() {
	log.Debug("[p2p]Node info:")
	log.Debug("[p2p]", "syncState", this.syncState)
	log.Debug("[p2p]", "consState", this.consState)
	log.Debug("[p2p]", "id", this.GetID())
	log.Debug("[p2p]", "addr", this.SyncLink.GetAddr())
	log.Debug("[p2p]", "cap", this.cap)
	log.Debug("[p2p]", "version", this.GetVersion())
	log.Debug("[p2p]", "services", this.GetServices())
	log.Debug("[p2p]", "syncPort", this.GetSyncPort())
	log.Debug("[p2p]", "consPort", this.GetConsPort())
	log.Debug("[p2p]", "relay", this.GetRelay())
	log.Debug("[p2p]", "height", this.GetHeight())
}

//GetVersion return peer`s version
func (this *Peer) GetVersion() uint32 {
	return this.base.GetVersion()
}

//GetHeight return peer`s block height
func (this *Peer) GetHeight() uint64 {
	return this.base.GetHeight()
}

//SetHeight set height to peer
func (this *Peer) SetHeight(height uint64) {
	this.base.SetHeight(height)
}

//GetOrgHeight return org peer`s block height
func (this *Peer) GetRemoteOrgHeight(id comm.Address) uint64 {
	return this.base.GetRemoteOrgHeight(id)
}

//SetOrgHeight set org height to peer
func (this *Peer) SetRemoteOrgHeight(height uint64, id comm.Address) bool {
	return this.base.SetRemoteOrgHeight(height, id)
}

// GetRemoteOrgs get Remote node orgs
func (this *Peer) GetRemoteOrgs() []comm.Address {
	return this.base.GetRemoteOrgs()
}

// DelRemoteOrg del remote org to peer
func (this *Peer) DelRemoteOrg(id comm.Address) bool {
	return this.base.DelRemoteOrg(id)
}

//BHaveOrgs the Remote node have org
func (this *Peer) BHaveOrgs() bool {
	return this.base.BHaveOrgs()
}

//BHaveOrgs the Remote node have org
func (this *Peer) BHaveOrgsExceptId(id comm.Address) bool {
	return this.base.BHaveOrgsExceptId(id)
}

//BHaveOrgID the Remote node have orgid
func (this *Peer) BHaveOrgID(id comm.Address) bool {
	return this.base.BHaveOrgID(id)
}

//GetConsConn return consensus link
func (this *Peer) GetConsConn() *conn.Link {
	return this.ConsLink
}

//SetConsConn set consensue link to peer
func (this *Peer) SetConsConn(consLink *conn.Link) {
	this.ConsLink = consLink
}

//GetSyncState return sync state
func (this *Peer) GetSyncState() uint32 {
	return this.syncState
}

//SetSyncState set sync state to peer
func (this *Peer) SetSyncState(state uint32) {
	atomic.StoreUint32(&(this.syncState), state)
}

//GetConsState return peer`s consensus state
func (this *Peer) GetConsState() uint32 {
	return this.consState
}

//SetConsState set consensus state to peer
func (this *Peer) SetConsState(state uint32) {
	atomic.StoreUint32(&(this.consState), state)
}

//GetSyncPort return peer`s sync port
func (this *Peer) GetSyncPort() uint16 {
	return this.SyncLink.GetPort()
}

//GetConsPort return peer`s consensus port
func (this *Peer) GetConsPort() uint16 {
	return this.ConsLink.GetPort()
}

//SetConsPort set peer`s consensus port
func (this *Peer) SetConsPort(port uint16) {
	this.ConsLink.SetPort(port)
}

//SendToSync call sync link to send buffer
func (this *Peer) SendToSync(msg common.Message) error {
	if this.SyncLink != nil && this.SyncLink.Valid() {
		return this.SyncLink.Tx(msg)
	}
	return errors.New("[p2p]sync link invalid")
}

//SendToCons call consensus link to send buffer
func (this *Peer) SendToCons(msg common.Message) error {
	if this.ConsLink != nil && this.ConsLink.Valid() {
		return this.ConsLink.Tx(msg)
	}
	return errors.New("[p2p]cons link invalid")
}

//CloseSync halt sync connection
func (this *Peer) CloseSync() {
	//fmt.Println(" ********* Peer CloseSync ... ")
	this.SetSyncState(common.INACTIVITY)
	conn := this.SyncLink.GetConn()
	rlpconn := this.SyncLink.GetRLPConn()
	this.connLock.Lock()
	if rlpconn != nil {
		rlpconn.Close()
	}

	if conn != nil {
		conn.Close()
	}
	this.connLock.Unlock()
}

//CloseCons halt consensus connection
func (this *Peer) CloseCons() {
	this.SetConsState(common.INACTIVITY)
	conn := this.ConsLink.GetConn()
	this.connLock.Lock()
	if conn != nil {
		conn.Close()

	}
	this.connLock.Unlock()
}

//GetID return peer`s id
func (this *Peer) GetID() uint64 {
	return this.base.GetID()
}

//GetRelay return peer`s relay state
func (this *Peer) GetRelay() bool {
	return this.base.GetRelay()
}

//GetServices return peer`s service state
func (this *Peer) GetServices() uint64 {
	return this.base.GetServices()
}

//GetTimeStamp return peer`s latest contact time in ticks
func (this *Peer) GetTimeStamp() int64 {
	return this.SyncLink.GetRXTime().UnixNano()
}

//GetContactTime return peer`s latest contact time in Time struct
func (this *Peer) GetContactTime() time.Time {
	return this.SyncLink.GetRXTime()
}

//GetAddr return peer`s sync link address
func (this *Peer) GetAddr() string {
	return this.SyncLink.GetAddr()
}

//GetAddr16 return peer`s sync link address in []byte
func (this *Peer) GetAddr16() ([16]byte, error) {
	var result [16]byte
	addrIp, err := common.ParseIPAddr(this.GetAddr())
	if err != nil {
		return result, err
	}
	ip := net.ParseIP(addrIp).To16()
	if ip == nil {
		log.Warn("[p2p]parse ip address error", "Addr", this.GetAddr())
		return result, errors.New("[p2p]parse ip address error")
	}

	copy(result[:], ip[:16])
	return result, nil
}

//AttachSyncChan set msg chan to sync link
func (this *Peer) AttachSyncChan(msgchan chan *common.MsgPayload) {
	this.SyncLink.SetChan(msgchan)
}

//AttachConsChan set msg chan to consensus link
func (this *Peer) AttachConsChan(msgchan chan *common.MsgPayload) {
	this.ConsLink.SetChan(msgchan)
}

//Send transfer buffer by sync or cons link
func (this *Peer) Send(msg common.Message, isConsensus bool) error {

	if isConsensus && this.ConsLink.Valid() {
		return this.SendToCons(msg)
	}
	return this.SendToSync(msg)
}

//SetHttpInfoState set peer`s httpinfo state
func (this *Peer) SetHttpInfoState(httpInfo bool) {
	if httpInfo {
		this.cap[common.HTTP_INFO_FLAG] = 0x01
	} else {
		this.cap[common.HTTP_INFO_FLAG] = 0x00
	}
}

//GetHttpInfoState return peer`s httpinfo state
func (this *Peer) GetHttpInfoState() bool {
	return this.cap[common.HTTP_INFO_FLAG] == 1
}

//GetHttpInfoPort return peer`s httpinfo port
func (this *Peer) GetHttpInfoPort() uint16 {
	return this.base.GetHttpInfoPort()
}

//SetHttpInfoPort set peer`s httpinfo port
func (this *Peer) SetHttpInfoPort(port uint16) {
	this.base.SetHttpInfoPort(port)
}

//GetHttpInfoPort return peer`s httpinfo port
func (this *Peer) GetNode() *discover.Node {
	return this.node
}

//SetHttpInfoPort set peer`s httpinfo port
func (this *Peer) SetNode(node *discover.Node) {
	this.node = node
}

//UpdateInfo update peer`s information
func (this *Peer) UpdateInfo(t time.Time, version uint32, services uint64,
	syncPort uint16, consPort uint16, nonce uint64, relay uint8, height uint64) {

	this.SyncLink.UpdateRXTime(t)
	this.base.SetID(nonce)
	this.base.SetVersion(version)
	this.base.SetServices(services)
	this.base.SetSyncPort(syncPort)
	this.base.SetConsPort(consPort)
	this.SyncLink.SetPort(syncPort)
	this.ConsLink.SetPort(consPort)
	if relay == 0 {
		this.base.SetRelay(false)
	} else {
		this.base.SetRelay(true)
	}
	this.SetHeight(uint64(height))
}
