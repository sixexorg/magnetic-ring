package p2pserver

import (
	"crypto/ecdsa"
	//"fmt"

	// "encoding/json"
	"errors"
	// "io/ioutil"
	"math/rand"
	"net"

	// "os"
	"reflect"
	"strconv"

	// "strings"
	"sync"
	"time"

	evtActor "github.com/ontio/ontology-eventbus/actor"
	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/log"
	ledger "github.com/sixexorg/magnetic-ring/store/mainchain/storages"

	// orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sixexorg/magnetic-ring/p2pserver/common"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"github.com/sixexorg/magnetic-ring/p2pserver/net/netserver"
	p2pnet "github.com/sixexorg/magnetic-ring/p2pserver/net/protocol"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"
	utils "github.com/sixexorg/magnetic-ring/p2pserver/sync"

	// table
	"github.com/sixexorg/magnetic-ring/p2pserver/discover"
	"github.com/sixexorg/magnetic-ring/p2pserver/spcnode"

	// org

	"fmt"

	mainsync "github.com/sixexorg/magnetic-ring/p2pserver/sync/main"
	"github.com/sixexorg/magnetic-ring/p2pserver/temp"
)

type recentData struct {
	addr    string
	nodestr string
}

//P2PServer control all network activities
type P2PServer struct {
	network   p2pnet.P2P
	msgRouter *utils.MessageRouter
	pid       *evtActor.PID
	blockSync *mainsync.BlockSyncMgr
	ledger    *ledger.LedgerStoreImp
	ReconnectAddrs
	quitSyncRecent chan bool
	quitOnline     chan bool
	quitHeartBeat  chan bool
	quitCheckConn  chan bool
	// table
	ntab      *discover.Table
	bootnodes []*discover.Node
	nodeKey   *ecdsa.PrivateKey
	// org
	orgman   *OrgMan
	nodeA    *spcnode.NodeA
	nodeStar *spcnode.NodeStar
}

type retryPeerInfo struct {
	num     int
	nodestr string
}

//ReconnectAddrs contain addr need to reconnect
type ReconnectAddrs struct {
	sync.RWMutex
	RetryAddrs map[string]*retryPeerInfo
}

//NewServer return a new p2pserver according to the pubkey
func NewServer() *P2PServer {
	common.InitP2PVariable()

	p := &P2PServer{}
	n := netserver.NewNetServer(p)

	p.network = n
	p.ledger = ledger.GetLedgerStore()

	p.msgRouter = utils.NewMsgRouter(p.network)
	p.blockSync = mainsync.NewBlockSyncMgr(p)
	p.quitSyncRecent = make(chan bool)
	p.quitOnline = make(chan bool)
	p.quitHeartBeat = make(chan bool)
	p.quitCheckConn = make(chan bool)
	p.orgman = NewAddOrg()
	p.nodeA = spcnode.NewNodeA(p)
	p.nodeStar = spcnode.NewNodeStar(p)
	return p
}

//GetConnectionCnt return the established connect count
func (this *P2PServer) GetLedger() *ledger.LedgerStoreImp {
	return this.ledger
}

//GetConnectionCnt return the established connect count
func (this *P2PServer) GetConnectionCnt() uint32 {
	return this.network.GetConnectionCnt()
}

func (this *P2PServer) genPrikey() error {
	var (
		nodeKey *ecdsa.PrivateKey
		err     error
	)
	nodeKey, err = crypto.GenerateKey()
	if err != nil {
		return err
	}
	this.nodeKey = nodeKey

	portstr := strconv.Itoa(int(config.GlobalConfig.P2PCfg.NodePort))
	ntab, err := discover.ListenUDP(this.nodeKey, "0.0.0.0:"+portstr, nil, "", nil,
		this.TabCallPeerConnectInfo, this.TabCallPeerOrgInfo, this.TabCallSendConnectOrg, false, comm.Address{})
	if err != nil {
		return err
	}

	this.ntab = ntab
	this.bootnodes = make([]*discover.Node, 0)
	// gen bootnode
	for _, nodestr := range common.BootNodes {
		node, err := discover.ParseNode(nodestr)
		if err != nil {
			return err
		}
		this.bootnodes = append(this.bootnodes, node)
	}

	if err := ntab.SetFallbackNodes(this.bootnodes); err != nil {
		return err
	}
	// deliver prvikey to network
	this.network.SetPrivateKey(this.nodeKey, this.ntab)
	this.network.SetBootNodes(this.bootnodes)

	return nil
}

//Start create all services
func (this *P2PServer) Start() error {
	if this.genPrikey() != nil {
		return errors.New("[p2p]genPrikey invalid")
	}

	if this.network == nil {
		return errors.New("[p2p]network invalid")
	}

	this.network.Start()

	if this.msgRouter == nil {
		return errors.New("[p2p]msg router invalid")
	}

	this.msgRouter.Start()

	go this.connectSeedService()
	go this.keepOnlineService()
	go this.heartBeatService()
	go this.blockSync.Start()
	// NodeA
	this.nodeA.Start()
	// NodeStar
	this.nodeStar.Start()
	// for test
	// go this.testOrg()
	return nil
}

//Stop halt all service by send signal to channels
func (this *P2PServer) Stop() {
	this.network.Halt()
	this.quitSyncRecent <- true
	this.quitOnline <- true
	this.quitHeartBeat <- true
	this.quitCheckConn <- true
	this.msgRouter.Stop()
	this.blockSync.Close()
	this.nodeA.Close()
	this.nodeStar.Close()

	this.orgman.RLock()
	for _, orgsync := range this.orgman.OrgSync {
		orgsync.Close()
	}
	this.orgman.Reset()
	this.orgman.RUnlock()
}

// GetNetWork returns the low level netserver
func (this *P2PServer) GetNetWork() p2pnet.P2P {
	return this.network
}

//GetPort return two network port
func (this *P2PServer) GetPort() (uint16, uint16) {
	return this.network.GetSyncPort(), this.network.GetConsPort()
}

//GetVersion return self version
func (this *P2PServer) GetVersion() uint32 {
	return this.network.GetVersion()
}

//GetNeighborAddrs return all nbr`s address
func (this *P2PServer) GetNeighborAddrs() []common.PeerAddr {
	return this.network.GetNeighborAddrs()
}

//Xmit called by other module to broadcast msg
func (this *P2PServer) Xmit(message interface{}) error {
	log.Debug("[p2p] Xmit Start...")
	var msg common.Message
	isConsensus := false
	switch message.(type) {
	case *types.Transaction: //
		log.Debug("[p2p]TX transaction message ••••••••••••••••")
		txn := message.(*types.Transaction)
		msg = msgpack.NewTxn(txn, nil, comm.Address{}, common.SYNC_DATA_MAIN)


	case *common.OrgTx: //
		log.Debug("[p2p]TX transaction message")
		orgtx := message.(*common.OrgTx)
		txn := orgtx.Tx
		msg = msgpack.NewTxn(nil, txn, comm.Address{}, common.SYNC_DATA_ORG)

	case *types.Block: // sync bblock //
		log.Debug("[p2p]TX block message")
		block := message.(*types.Block)
		msg = msgpack.NewBlock(block, nil, comm.Address{}, common.SYNC_DATA_MAIN)
	case *common.ConsensusPayload:
		log.Debug("[p2p]TX consensus message")
		consensusPayload := message.(*common.ConsensusPayload)
		msg = msgpack.NewConsensus(consensusPayload)
		isConsensus = true

	case *common.NotifyBlk:
		//return nil
		blkntf := message.(*common.NotifyBlk)
		//time.Sleep(time.Millisecond*500)

		log.Info("[p2p]TX block hash message unique", "hash", blkntf)

		// construct inv message
		invPayload := msgpack.NewInvPayload(comm.BLOCK, []comm.Hash{blkntf.BlkHash}, []uint64{blkntf.BlkHeight})

		msg = msgpack.NewInv(invPayload)
	case *common.EarthNotifyBlk:
		ntf := message.(*common.EarthNotifyBlk)

		msg = msgpack.NewEarthNotifyHash(ntf)
	case *comm.NodeLH:
		fmt.Println("🌐 📩  p2p receive nodeLH")
		nodeLH := message.(*comm.NodeLH)
		msg = msgpack.NewNodeLHMsg(nodeLH)
	case *common.ExtDataRequest:
		fmt.Println("🌐 📩  p2p extData requtst")
		extData := message.(*common.ExtDataRequest)
		msg  = msgpack.NewExtDataRequestMsg(extData)
	default:
		log.Info("[p2p]Unknown Xmit ", "message", message, "type", reflect.TypeOf(message))
		return errors.New("[p2p]Unknown Xmit message type")
	}
	this.network.Xmit(msg, isConsensus)
	return nil
}

//Send tranfer buffer to peer
func (this *P2PServer) Send(p *peer.Peer, msg common.Message,
	isConsensus bool) error {
	if this.network.IsPeerEstablished(p) {
		return this.network.Send(p, msg, isConsensus)
	}
	log.Warn("[p2p]send to a not ESTABLISH ", "peerid", p.GetID())
	return errors.New("[p2p]send to a not ESTABLISH peer")
}

// GetID returns local node id
func (this *P2PServer) GetID() uint64 {
	return this.network.GetID()
}

// OnAddNode adds the peer id to the block sync mgr
func (this *P2PServer) SetSyncStatus(status bool,lockHeight uint64) {
	this.blockSync.SetSyncStatus(status,lockHeight)
}

// OnAddNode adds the peer id to the block sync mgr
func (this *P2PServer) OnAddNode(id uint64) {
	this.blockSync.OnAddNode(id)
}

// OnDelNode removes the peer id from the block sync mgr
func (this *P2PServer) OnDelNode(id uint64) {
	this.blockSync.OnDelNode(id)
}

// OnHeaderReceive adds the header list from network
func (this *P2PServer) OnHeaderReceive(fromID uint64, headers []*types.Header) {
	this.blockSync.OnHeaderReceive(fromID, headers)
}

// OnBlockReceive adds the block from network
func (this *P2PServer) OnBlockReceive(fromID uint64, blockSize uint32, block *types.Block) {
	this.blockSync.OnBlockReceive(fromID, blockSize, block)
}

// Todo: remove it if no use
func (this *P2PServer) GetConnectionState() uint32 {
	return common.INIT
}

//GetTime return lastet contact time
func (this *P2PServer) GetTime() int64 {
	return this.network.GetTime()
}

// SetPID sets p2p actor
func (this *P2PServer) SetPID(pid *evtActor.PID) {
	this.pid = pid
	this.msgRouter.SetPID(pid)
}

// GetPID returns p2p actor
func (this *P2PServer) GetPID() *evtActor.PID {
	return this.pid
}

//blockSyncFinished compare all nbr peers and self height at beginning
func (this *P2PServer) blockSyncFinished() bool {
	peers := this.network.GetNeighbors()
	if len(peers) == 0 {
		return false
	}

	blockHeight := this.ledger.GetCurrentBlockHeight()

	for _, v := range peers {
		if blockHeight < v.GetHeight() {
			return false
		}
	}
	return true
}

//WaitForSyncBlkFinish compare the height of self and remote peer in loop
func (this *P2PServer) WaitForSyncBlkFinish() {
	for {
		headerHeight := this.ledger.GetCurrentHeaderHeight()
		currentBlkHeight := this.ledger.GetCurrentBlockHeight()
		log.Info("[p2p]WaitForSyncBlkFinish... current block height is ", "currentBlkHeight",
			currentBlkHeight, " ,current header height is ", headerHeight)

		if this.blockSyncFinished() {
			break
		}

		<-time.After(time.Second * (time.Duration(common.SYNC_BLK_WAIT)))
	}
}

//WaitForPeersStart check whether enough peer linked in loop
func (this *P2PServer) WaitForPeersStart() {
	periodTime := common.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for {
		log.Info("[p2p]Wait for minimum connection...")
		if this.ReachMinConnection() {
			break
		}

		<-time.After(time.Second * (time.Duration(periodTime)))
	}
}

func (this *P2PServer) getRandList() map[string]*discover.Node {
	nodes := make([]*discover.Node, 5)
	num := this.ntab.ReadRandomNodes(nodes)
	lennode := len(nodes)
	seedNodes := make(map[string]*discover.Node)

	bootmap := make(map[discover.NodeID]bool)
	for _, node := range this.bootnodes {
		bootmap[node.ID] = true
	}

	for i := 0; i < num && i < lennode; i++ {
		if nodes[i] == nil {
			continue
		}
		if ok := bootmap[nodes[i].ID]; !ok {
			addr := &net.TCPAddr{IP: nodes[i].IP, Port: int(nodes[i].TCP)}
			seedNodes[addr.String()] = nodes[i]
		}
	}
	return seedNodes
}

//connectSeeds connect the seeds in seedlist and call for nbr list
func (this *P2PServer) connectSeeds() {
	seedNodes := make(map[string]*discover.Node)
	pList := make([]*peer.Peer, 0)

	seedNodes = this.getRandList()

	for nodeAddr, _ := range seedNodes {
		var ip net.IP
		np := this.network.GetNp()
		np.Lock()
		for _, tn := range np.List {
			ipAddr, _ := tn.GetAddr16()
			ip = ipAddr[:]
			addrString := ip.To16().String() + ":" +
				strconv.Itoa(int(tn.GetSyncPort()))
			if nodeAddr == addrString && tn.GetSyncState() == common.ESTABLISH {
				pList = append(pList, tn)
			}
		}
		np.Unlock()
	}
	if len(pList) > 0 {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(pList))
		this.reqNbrList(pList[index])
	} else { //not found
		for nodeAddr, node := range seedNodes {
			go this.network.Connect(nodeAddr, false, node, false)
		}
	}
}

//reachMinConnection return whether net layer have enough link under different config
func (this *P2PServer) ReachMinConnection() bool {
	// consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	// if consensusType == "" {
	// 	consensusType = "dbft"
	// }
	// minCount := config.DBFT_MIN_NODE_NUM
	// switch consensusType {
	// case "dbft":
	// case "solo":
	// 	minCount = config.SOLO_MIN_NODE_NUM
	// case "vbft":
	// 	minCount = config.VBFT_MIN_NODE_NUM

	// }
	// for test
	minCount := 2
	return int(this.GetConnectionCnt())+1 >= minCount
}

//getNode returns the peer with the id
func (this *P2PServer) GetNode(id uint64) *peer.Peer {
	return this.network.GetPeer(id)
}

//getNode returns the peer with the id
func (this *P2PServer) GetNodeFromDiscoverID(discoverNodeID string) *peer.Peer {
	return this.network.GetNp().GetPeerFromDiscoverNodeId(discoverNodeID)
}

//retryInactivePeer try to connect peer in INACTIVITY state
func (this *P2PServer) retryInactivePeer() {
	np := this.network.GetNp()
	np.Lock()
	var ip net.IP
	neighborPeers := make(map[uint64]*peer.Peer)
	for _, p := range np.List {
		addr, _ := p.GetAddr16()
		ip = addr[:]
		nodeAddr := ip.To16().String() + ":" +
			strconv.Itoa(int(p.GetSyncPort()))
		if p.GetSyncState() == common.INACTIVITY {
			log.Debug("[p2p] try reconnect", "nodeAddr", nodeAddr)
			//add addr to retry list
			this.addToRetryList(nodeAddr, p.GetNode())
			p.CloseSync()
			p.CloseCons()
		} else {
			//add others to tmp node map
			this.removeFromRetryList(nodeAddr)
			neighborPeers[p.GetID()] = p
		}
	}

	np.List = neighborPeers
	np.Unlock()

	connCount := uint(this.network.GetOutConnRecordLen())
	if connCount >= config.GlobalConfig.P2PCfg.MaxConnOutBound {
		log.Warn("[p2p]Connect: out connections", "connCount", connCount,
			"reach the max limit", config.GlobalConfig.P2PCfg.MaxConnOutBound)
		return
	}

	//try connect
	if len(this.RetryAddrs) > 0 {
		this.ReconnectAddrs.Lock()

		list := make(map[string]*retryPeerInfo)
		addrs := make([]string, 0, len(this.RetryAddrs))
		nodestr := make([]string, 0, len(this.RetryAddrs))
		for addr, v := range this.RetryAddrs {
			v.num += 1
			addrs = append(addrs, addr)
			nodestr = append(nodestr, v.nodestr)
			if v.num < common.MAX_RETRY_COUNT {
				list[addr] = v
			}
			if v.num >= common.MAX_RETRY_COUNT {
				this.network.RemoveFromConnectingList(addr)
				remotePeer := this.network.GetPeerFromAddr(addr)
				if remotePeer != nil {
					if remotePeer.SyncLink.GetAddr() == addr {
						this.network.RemovePeerSyncAddress(addr)
						this.network.RemovePeerConsAddress(addr)
					}
					if remotePeer.ConsLink.GetAddr() == addr {
						this.network.RemovePeerConsAddress(addr)
					}
					this.network.DelNbrNode(remotePeer.GetID())
				}
			}
		}

		this.RetryAddrs = list
		this.ReconnectAddrs.Unlock()
		for index, addr := range addrs {
			rand.Seed(time.Now().UnixNano())
			log.Debug("[p2p]Try to reconnect peer, peer addr is ", "addr", addr)
			<-time.After(time.Duration(rand.Intn(common.CONN_MAX_BACK)) * time.Millisecond)
			log.Debug("[p2p]Back off time`s up, start connect node")
			if index < len(nodestr) && index >= 0 {
				temnode, err := discover.ParseNode(nodestr[index])
				if err != nil {
					continue
				}
				this.network.Connect(addr, false, temnode, false)
			}
		}

	}
}

//connectSeedService make sure seed peer be connected
func (this *P2PServer) connectSeedService() {
	t := time.NewTimer(time.Second * common.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			this.connectSeeds()
			t.Stop()
			if this.ReachMinConnection() {
				t.Reset(time.Second * time.Duration(10*common.CONN_MONITOR))
			} else {
				t.Reset(time.Second * common.CONN_MONITOR)
			}
		case <-this.quitOnline:
			t.Stop()
			break
		}
	}
}

//keepOnline try connect lost peer
func (this *P2PServer) keepOnlineService() {
	t := time.NewTimer(time.Second * common.CONN_MONITOR)
	for {
		select {
		case <-t.C:
			this.retryInactivePeer()
			t.Stop()
			t.Reset(time.Second * common.CONN_MONITOR)
		case <-this.quitOnline:
			t.Stop()
			break
		}
	}
}

//reqNbrList ask the peer for its neighbor list
func (this *P2PServer) reqNbrList(p *peer.Peer) {
	msg := msgpack.NewAddrReq()
	go this.Send(p, msg, false)
}

//heartBeat send ping to nbr peers and check the timeout
func (this *P2PServer) heartBeatService() {
	var periodTime uint
	periodTime = common.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))

	for {
		select {
		case <-t.C:
			this.ping()
			this.timeout()
		case <-this.quitHeartBeat:
			t.Stop()
			break
		}
	}
}

//ping send pkg to get pong msg from others
func (this *P2PServer) ping() {

	peers := this.network.GetNeighbors()
	fmt.Println("==========>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", time.Now().String(), "len(peers):",len(peers))
	this.PingTo(peers, true)
}

// pings send pkgs to get pong msg from others
// add org and StellarNode need call this func
func (this *P2PServer) PingTo(peers []*peer.Peer, bTimer bool, orgID ...comm.Address) {
	for _, p := range peers {
		fmt.Println("P2PServerPingTo out",  "time", time.Now().String(), "GetSyncState", p.GetSyncState())
		log.Info("P2PServerPingTo out",  "time", time.Now().String(), "GetSyncState", p.GetSyncState())
		if p.GetSyncState() == common.ESTABLISH {
			if bTimer {
				height := this.ledger.GetCurrentBlockHeight()
				//
				orgs := this.network.PeerGetOrg()
				orgpings := this.getOrgInfo(orgs)
				ping := msgpack.NewPingMsg(uint64(height), orgpings, common.PING_INFO_ALL)
				log.Info("P2PServerPingTo", "height", height, "time", time.Now().String())
				go this.Send(p, ping, false)
			} else {
				if len(orgID) <= 0 { // main
					height := this.ledger.GetCurrentBlockHeight()
					ping := msgpack.NewPingMsg(uint64(height), nil, common.PING_INFO_MAIN)
					go this.Send(p, ping, false)
				} else {
					orgs := []comm.Address{orgID[0]}
					orgpings := this.getOrgInfo(orgs)
					ping := msgpack.NewPingMsg(uint64(0), orgpings, common.PING_INFO_ORG)
					go this.Send(p, ping, false)
				}
			}
		}
	}
}

func (this *P2PServer) getOrgInfo(orgIDArr []comm.Address) []*common.OrgPIPOInfo {
	result := make([]*common.OrgPIPOInfo, 0)
	for _, id := range orgIDArr {
		if id == common.StellarNodeID { //
			result = append(result, msgpack.NewOrgPIPOMsg(id, uint64(0)))
		} else {
			// SFC,
			//height := this.ledger.GetCurrentBlockHeight(/*id*/)
			orgLedger := temp.GetLedger(id)
			if orgLedger == nil {
				continue
			}
			height := orgLedger.GetCurrentBlockHeight( /*id*/ )
			result = append(result, msgpack.NewOrgPIPOMsg(id, uint64(height)))
		}
	}
	return result
}

//timeout trace whether some peer be long time no response
func (this *P2PServer) timeout() {
	peers := this.network.GetNeighbors()
	var periodTime uint
	periodTime = common.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for _, p := range peers {
		if p.GetSyncState() == common.ESTABLISH {
			t := p.GetContactTime()
			if t.Before(time.Now().Add(-1 * time.Second *
				time.Duration(periodTime) * common.KEEPALIVE_TIMEOUT)) {
				log.Warn("[p2p]keep alive timeout!!!lost remote", "peer", p.GetID(), "Addr", p.SyncLink.GetAddr(), "Addr", t.String())
				p.CloseSync()
				p.CloseCons()
			}
		}
	}
}

//addToRetryList add retry address to ReconnectAddrs
func (this *P2PServer) addToRetryList(addr string, node *discover.Node) {
	this.ReconnectAddrs.Lock()
	defer this.ReconnectAddrs.Unlock()
	if this.RetryAddrs == nil {
		this.RetryAddrs = make(map[string]*retryPeerInfo)
	}
	if _, ok := this.RetryAddrs[addr]; ok {
		delete(this.RetryAddrs, addr)
	}
	//alway set retry to 0
	this.RetryAddrs[addr] = &retryPeerInfo{
		num:     0,
		nodestr: node.String(),
	}
}

//removeFromRetryList remove connected address from ReconnectAddrs
func (this *P2PServer) removeFromRetryList(addr string) {
	this.ReconnectAddrs.Lock()
	defer this.ReconnectAddrs.Unlock()
	if len(this.RetryAddrs) > 0 {
		if _, ok := this.RetryAddrs[addr]; ok {
			delete(this.RetryAddrs, addr)
		}
	}
}
