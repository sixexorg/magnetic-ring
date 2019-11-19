package sync

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/bactor"

	"github.com/sixexorg/magnetic-ring/p2pserver/temp"

	"github.com/hashicorp/golang-lru"
	evtActor "github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/log"

	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	actor "github.com/sixexorg/magnetic-ring/p2pserver/actor/req"
	msgCommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
	msgpack "github.com/sixexorg/magnetic-ring/p2pserver/message"
	"github.com/sixexorg/magnetic-ring/p2pserver/net/protocol"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstorages"
	ledger "github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

//respCache cache for some response data
var (
	respCache *lru.ARCCache

	nlhcache *nlhCache
)

type nlhCache struct {
	nlh     map[string]int64
	nlhLock sync.Mutex
}

func (nlh *nlhCache) watch() {
	ticker := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-ticker.C:
			now := time.Now().Unix()
			for k, v := range nlh.nlh {
				diff := now - v
				if diff >= 300 {
					delete(nlh.nlh, k)
				}
			}
		}
	}
}

func initNlhCache() {
	nlhcache = new(nlhCache)
	nlhcache.nlh = make(map[string]int64)
	go nlhcache.watch()
}

// AddrReqHandle handles the neighbor address request from peer
func AddrReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive addr request message", "data.Addr", data.Addr,
		"data.Id", data.Id)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in AddrReqHandle")
		return
	}

	var addrStr []msgCommon.PeerAddr
	addrStr = p2p.GetNeighborAddrs()
	//check mask peers
	mskPeers := config.GlobalConfig.P2PCfg.ReservedCfg.MaskPeers
	if config.GlobalConfig.P2PCfg.ReservedPeersOnly && len(mskPeers) > 0 {
		for i := 0; i < len(addrStr); i++ {
			var ip net.IP
			ip = addrStr[i].IpAddr[:]
			address := ip.To16().String()
			for j := 0; j < len(mskPeers); j++ {
				if address == mskPeers[j] {
					addrStr = append(addrStr[:i], addrStr[i+1:]...)
					i--
					break
				}
			}
		}

	}
	msg := msgpack.NewAddrs(addrStr)
	err := p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn("err", "err", err)
		return
	}
}

// HeaderReqHandle handles the header sync req from peer
func HeadersReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive headers request message", "Addr", data.Addr, "Id", data.Id)

	headersReq := data.Payload.(*msgCommon.HeadersReq)
	// 圈子的特殊处理
	if headersReq.SyncType == msgCommon.SYNC_DATA_ORG {
		OrgHeadersReqHandle(data, p2p, pid)
		return
	}
	startHash := headersReq.HashStart
	stopHash := headersReq.HashEnd

	// headers, err := GetHeadersFromHash(startHash, stopHash,headersReq.OrgID,headersReq.SyncType)

	headers, err := GetHeadersFromHeight(headersReq.Height)
	if err != nil {
		log.Warn("get headers in HeadersReqHandle ", "error", err.Error(), "startHash", startHash.String(), "stopHash", stopHash.String())
		return
	}
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in HeadersReqHandle, peer", "id", data.Id)
		return
	}
	msg := msgpack.NewHeaders(headers, nil, headersReq.OrgID, headersReq.SyncType)
	err = p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn("err", "err", err)
		return
	}
}

//PingHandle handle ping msg from peer
func PingHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receiveping message", "Addr", data.Addr, "Id", data.Id, "Height", data.Payload.(*msgCommon.Ping).Height)

	ping := data.Payload.(*msgCommon.Ping)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in PingHandle")
		return
	}
	remotetype := ping.InfoType
	height := uint32(0)
	owninfo := make([]*msgCommon.OrgPIPOInfo, 0)
	if remotetype == msgCommon.PING_INFO_ALL || remotetype == msgCommon.PING_INFO_MAIN {
		//remote
		remotePeer.SetHeight(ping.Height)
		//own
		height = uint32(ledger.GetLedgerStore().GetCurrentBlockHeight())
		p2p.SetHeight(uint64(height))
	}

	remotedelorgs := make([]common.Address, 0)
	if remotetype == msgCommon.PING_INFO_ALL {

		orgs := p2p.PeerGetOrg()
		ownorgmap := make(map[common.Address]bool)

		for _, ownorgid := range orgs {
			ownorgmap[ownorgid] = true
		}

		if len(ping.OrgInfo) == 0 {
			for id, _ := range ownorgmap {
				remotedelorgs = handleDelPeer(data, p2p, pid, id, remotedelorgs)
			}
		}

		for _, info := range ping.OrgInfo {

			if _, ok := ownorgmap[info.OrgID]; ok {
				//updata remote peer orgid -> height
				remotePeer.SetRemoteOrgHeight(info.Height, info.OrgID)
				//own
				orgheight := uint64(0)
				if info.OrgID != msgCommon.StellarNodeID {
					if temp.GetLedger(info.OrgID, "sync_handler.go 001") != nil {
						orgheight = temp.GetLedger(info.OrgID, "sync_handler.go 002").GetCurrentBlockHeight()
					}
				}
				// orgheight := ledger.GetLedgerStore().GetCurrentBlockHeight(/*info.OrgID*/)
				ownorginfo := msgpack.NewOrgPIPOMsg(info.OrgID, uint64(orgheight))
				owninfo = append(owninfo, ownorginfo)

				p2p.SetOwnOrgHeight(uint64(orgheight), info.OrgID)
			} else {
				remotedelorgs = handleDelPeer(data, p2p, pid, info.OrgID, remotedelorgs)
			}
		}
	} else if remotetype == msgCommon.PING_INFO_ORG && len(ping.OrgInfo) == 1 {
		remoteorginfo := ping.OrgInfo[0]
		if p2p.BHaveOrgsId(remoteorginfo.OrgID) {
			//updata remote peer orgid -> height
			remotePeer.SetRemoteOrgHeight(remoteorginfo.Height, remoteorginfo.OrgID)
			//own
			orgheight := uint64(0)
			if temp.GetLedger(remoteorginfo.OrgID, "sync_handler.go 003") != nil {
				orgheight = temp.GetLedger(remoteorginfo.OrgID, "sync_handler.go 004").GetCurrentBlockHeight()
			}
			// orgheight := ledger.GetLedgerStore().GetCurrentBlockHeight(/*remoteorginfo.OrgID*/)
			ownorginfo := msgpack.NewOrgPIPOMsg(remoteorginfo.OrgID, uint64(orgheight))
			owninfo = append(owninfo, ownorginfo)

			p2p.SetOwnOrgHeight(uint64(orgheight), remoteorginfo.OrgID)
		} else {
			remotedelorgs = handleDelPeer(data, p2p, pid, remoteorginfo.OrgID, remotedelorgs)
		}
	}

	if len(remotedelorgs) > 0 {
		appendorgs := &msgCommon.HandleAddDelOrgPeerID{
			BAdd:   false,
			PeerID: data.Id,
			OrgID:  remotedelorgs,
		}
		pid.Tell(appendorgs)
	}

	// height: main chan
	msg := msgpack.NewPongMsg(uint64(height), owninfo, remotetype)
	err := p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn("err", "err", err)
	}
}

func handleDelPeer(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID,
	orgID common.Address, remotedelorgs []common.Address) []common.Address {
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]handleDelPeer invalid in PingHandle")
		return remotedelorgs
	}
	if msgCommon.StellarNodeID == orgID {
		bhave := remotePeer.DelRemoteOrg(msgCommon.StellarNodeID)
		if bhave {
			pid.Tell(&msgCommon.StellarAddDelPeerID{
				BAdd:   false,
				PeerID: data.Id,
			})
		}
	} else {
		remotedelorgs = append(remotedelorgs, orgID)
	}
	return remotedelorgs
}

///PongHandle handle pong msg from peer
func PongHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive pong message", "Addr", data.Addr, "Id", data.Id)

	pong := data.Payload.(*msgCommon.Pong)

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in PongHandle")
		return
	}
	remotetype := pong.InfoType
	if remotetype == msgCommon.PING_INFO_ALL || remotetype == msgCommon.PING_INFO_MAIN {
		//remote
		remotePeer.SetHeight(pong.Height)
	}

	remoteorgs := make([]common.Address, 0)
	if remotetype == msgCommon.PING_INFO_ALL || remotetype == msgCommon.PING_INFO_ORG {

		ownorgs := p2p.PeerGetOrg()
		maporg := make(map[common.Address]bool)

		for _, ownorgid := range ownorgs {
			maporg[ownorgid] = true
		}

		for _, info := range pong.OrgInfo {

			if _, ok := maporg[info.OrgID]; ok {
				//remote
				bnew := remotePeer.SetRemoteOrgHeight(info.Height, info.OrgID)
				if msgCommon.StellarNodeID == info.OrgID {
					if bnew {
						pid.Tell(&msgCommon.StellarAddDelPeerID{
							BAdd:   true,
							PeerID: data.Id,
						})
					}
				} else {

					remoteorgs = append(remoteorgs, info.OrgID)
				}
			}
		}
	}
	if len(remoteorgs) > 0 {
		appendorgs := &msgCommon.HandleAddDelOrgPeerID{
			BAdd:   true,
			PeerID: data.Id,
			OrgID:  remoteorgs,
		}
		pid.Tell(appendorgs)
	}
}

// BlkHeaderHandle handles the sync headers from peer
func BlkHeaderHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive block header message", "Addr", data.Addr, "Id", data.Id)
	if pid != nil {
		var blkHeader = data.Payload.(*msgCommon.BlkHeader)
		input := &msgCommon.AppendHeaders{
			FromID:     data.Id,
			Headers:    blkHeader.BlkHdr,
			HeaderOrgs: blkHeader.BlkOrgHdr,
			OrgID:      blkHeader.OrgID,
			SyncType:   blkHeader.SyncType,
		}
		pid.Tell(input)
	}
}

// BlockHandle handles the block message from peer
func BlockHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Info("[p2p]receive block message from p2p", "Addr", data.Addr, "Id", data.Id, "data.Payload.CmdType", data.Payload.CmdType())

	if pid != nil {
		var block = data.Payload.(*msgCommon.Block)

		if block != nil {
			if block.SyncType == msgCommon.SYNC_DATA_A_TO_STELLAR || block.SyncType == msgCommon.SYNC_DATA_STELLAR_TO_STELLAR || block.SyncType == msgCommon.SYNC_DATA_ORG {

				if block.SyncType == msgCommon.SYNC_DATA_ORG {
					OrgDataReqHandle(data, p2p, pid)
					return
				} else if block.SyncType == msgCommon.SYNC_DATA_A_TO_STELLAR {
					ANodeSendToStellarPeingData(data, p2p, pid)
					return
				} else if block.SyncType == msgCommon.SYNC_DATA_STELLAR_TO_STELLAR {
					StellaNodeSendToStellarPeingData(data, p2p, pid)
					return
				}

			}

		}

		input := &msgCommon.AppendBlock{
			FromID:    data.Id,
			BlockSize: data.PayloadSize,
			Block:     block.Blk,
			BlockOrg:  block.BlkOrg,
			OrgID:     block.OrgID,
			SyncType:  block.SyncType,
		}
		pid.Tell(input)
	}
}

// ConsensusHandle handles the consensus message from peer
func ConsensusHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("[p2p]receive consensus message", "Addr", data.Addr, "Id", data.Id)

	if actor.ConsensusPid != nil {
		var consensus = data.Payload.(*msgCommon.Consensus)
		if err := consensus.Cons.Verify(); err != nil {
			log.Warn("err", "err", err)
			return
		}
		consensus.Cons.PeerId = data.Id
		actor.ConsensusPid.Tell(&consensus.Cons)
	}
}

// NotFoundHandle handles the not found message from peer
func NotFoundHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	var notFound = data.Payload.(*msgCommon.NotFound)
	log.Debug("[p2p]receive notFound message, hash is ", "Hash", notFound.Hash)
}

// TransactionHandle handles the transaction message from peer
func TransactionHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive transaction message", "Addr", data.Addr, "Id", data.Id)

	var trn = data.Payload.(*msgCommon.Trn)
	// synctype := trn.SyncType
	// orgid := trn.OrgID
	// actor.AddTransaction(trn.Txn)
	actor.AddTransaction(trn)
	log.Trace("[p2p]receive Transaction message hash", "Hash", trn.Txn.Hash())

}

// VersionHandle handles version handshake protocol from peer
func VersionHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive version message", "Addr", data.Addr, "Id", data.Id)

	version := data.Payload.(*msgCommon.Version)

	remotePeer := p2p.GetPeerFromAddr(data.Addr)
	if remotePeer == nil {
		log.Debug("[p2p]peer is not exist", "Addr", data.Addr)
		//peer not exist,just remove list and return
		p2p.RemoveFromConnectingList(data.Addr)
		return
	}
	addrIp, err := msgCommon.ParseIPAddr(data.Addr)
	if err != nil {
		log.Warn("err", "err", err)
		return
	}
	nodeAddr := addrIp + ":" +
		strconv.Itoa(int(version.P.SyncPort))
	if config.GlobalConfig.P2PCfg.ReservedPeersOnly && len(config.GlobalConfig.P2PCfg.ReservedCfg.ReservedPeers) > 0 {
		found := false
		for _, addr := range config.GlobalConfig.P2PCfg.ReservedCfg.ReservedPeers {
			if strings.HasPrefix(data.Addr, addr) {
				log.Debug("[p2p]peer in reserved list", "Addr", data.Addr)
				found = true
				break
			}
		}
		if !found {
			remotePeer.CloseSync()
			remotePeer.CloseCons()
			log.Debug("[p2p]peer not in reserved list,close", "Addr", data.Addr)
			return
		}

	}

	if version.P.Nonce == p2p.GetID() {
		p2p.RemoveFromInConnRecord(remotePeer.GetAddr())
		p2p.RemoveFromOutConnRecord(remotePeer.GetAddr())
		log.Warn("[p2p]the node handshake with itself", "Addr", remotePeer.GetAddr())
		p2p.SetOwnAddress(nodeAddr)
		remotePeer.CloseSync()
		return
	}

	s := remotePeer.GetSyncState()
	if s != msgCommon.INIT && s != msgCommon.HAND {
		log.Warn("[p2p]unknown status to received version", "s", s, "Addr", remotePeer.GetAddr())
		remotePeer.CloseSync()
		return
	}

	// Obsolete node
	p := p2p.GetPeer(version.P.Nonce)
	if p != nil {
		ipOld, err := msgCommon.ParseIPAddr(p.GetAddr())
		if err != nil {
			log.Warn("[p2p]exist peer ip format is wrong", "Nonce", version.P.Nonce, "Addr", p.GetAddr())
			return
		}
		ipNew, err := msgCommon.ParseIPAddr(data.Addr)
		if err != nil {
			remotePeer.CloseSync()
			log.Warn("[p2p]connecting peer ip format is wrong, close", "Nonce", version.P.Nonce, "Addr", data.Addr)
			return
		}
		if ipNew == ipOld {
			//same id and same ip
			n, ret := p2p.DelNbrNode(version.P.Nonce)
			if ret == true {
				log.Info("[p2p]peer reconnect", "Nonce", version.P.Nonce, "Addr", data.Addr)
				// Close the connection and release the node source
				n.CloseSync()
				n.CloseCons()
				if pid != nil {
					input := &msgCommon.RemovePeerID{
						ID: version.P.Nonce,
					}
					pid.Tell(input)
				}
			}
		} else {
			log.Warn("[p2p]same peer id from different addr close latest one", "ipOld", ipOld, "ipNew", ipNew)
			remotePeer.CloseSync()
			return

		}
	}

	if version.P.Cap[msgCommon.HTTP_INFO_FLAG] == 0x01 {
		remotePeer.SetHttpInfoState(true)
	} else {
		remotePeer.SetHttpInfoState(false)
	}
	remotePeer.SetHttpInfoPort(version.P.HttpInfoPort)
	remotePeer.SetNode(&(version.N))

	remotePeer.UpdateInfo(time.Now(), version.P.Version,
		version.P.Services, version.P.SyncPort,
		version.P.ConsPort, version.P.Nonce,
		version.P.Relay, version.P.StartHeight)
	remotePeer.SyncLink.SetID(version.P.Nonce)
	p2p.AddNbrNode(remotePeer)

	if pid != nil {
		input := &msgCommon.AppendPeerID{
			ID: version.P.Nonce,
		}
		pid.Tell(input)
	}

	var msg msgCommon.Message
	if s == msgCommon.INIT {
		remotePeer.SetSyncState(msgCommon.HAND_SHAKE)
		msg = msgpack.NewVersion(p2p, false, uint32(ledger.GetLedgerStore().GetCurrentBlockHeight()))
	} else if s == msgCommon.HAND {
		remotePeer.SetSyncState(msgCommon.HAND_SHAKED)
		msg = msgpack.NewVerAck(false)
	}
	err = p2p.Send(remotePeer, msg, false)
	if err != nil {
		log.Warn("err", "err", err)
		return
	}
}

// VerAckHandle handles the version ack from peer
func VerAckHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive verAck message from ", "Addr", data.Addr, "Id", data.Id)

	// verAck := data.Payload.(*msgCommon.VerACK)
	remotePeer := p2p.GetPeer(data.Id)

	if remotePeer == nil {
		log.Warn("[p2p]nbr node is not exist", "Id", data.Id, "Addr", data.Addr)
		return
	}
	s := remotePeer.GetSyncState()
	if s != msgCommon.HAND_SHAKE && s != msgCommon.HAND_SHAKED {
		log.Warn("[p2p]unknown status to received verAck,state", "s", s, "Addr", data.Addr)
		return
	}

	remotePeer.SetSyncState(msgCommon.ESTABLISH)
	p2p.RemoveFromConnectingList(data.Addr)
	remotePeer.DumpInfo()

	if s == msgCommon.HAND_SHAKE {
		msg := msgpack.NewVerAck(false)
		p2p.Send(remotePeer, msg, false)
	}
	msg := msgpack.NewAddrReq()
	go p2p.Send(remotePeer, msg, false)
}

// AddrHandle handles the neighbor address response message from peer
func AddrHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]handle addr message", "Addr", data.Addr, "Id", data.Id)

	var msg = data.Payload.(*msgCommon.Addr)
	for _, v := range msg.NodeAddrs {
		var ip net.IP
		ip = v.IpAddr[:]
		address := ip.To16().String() + ":" + strconv.Itoa(int(v.Port))

		if v.ID == p2p.GetID() {
			continue
		}

		if p2p.NodeEstablished(v.ID) {
			continue
		}

		if ret := p2p.GetPeerFromAddr(address); ret != nil {
			continue
		}

		if v.Port == 0 {
			continue
		}
		if p2p.IsAddrFromConnecting(address) {
			continue
		}
		log.Debug("[p2p]connect ip ", "address", address)
		go p2p.Connect(address, false, &(v.Node), false)
	}
}

// DataReqHandle handles the data req(block/Transaction) from peer
func DataReqHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive data req message", "Addr", data.Addr, "Id", data.Id)
	var dataReq = data.Payload.(*msgCommon.DataReq)

	fmt.Println("🌐  📩 dataReqHandle ", dataReq.SyncType)
	if dataReq.SyncType == msgCommon.SYNC_DATA_ORG {
		OrgDataReqHandle(data, p2p, pid)
		return
	} else if dataReq.SyncType == msgCommon.SYNC_DATA_A_TO_STELLAR {
		ANodeSendToStellarPeingData(data, p2p, pid)
		return
	} else if dataReq.SyncType == msgCommon.SYNC_DATA_STELLAR_TO_STELLAR {
		StellaNodeSendToStellarPeingData(data, p2p, pid)
		return
	}

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in DataReqHandle")
		return
	}
	reqType := common.InventoryType(dataReq.DataType)
	hash := dataReq.Hash
	orgID := dataReq.OrgID
	synctype := dataReq.SyncType
	switch reqType {
	case common.BLOCK:
		reqID := fmt.Sprintf("%x%s", reqType, hash.String())
		data := getRespCacheValue(reqID)
		var block *types.Block
		var err error
		if data != nil {
			switch data.(type) {
			case *types.Block:
				block = data.(*types.Block)
			}
		}
		if block == nil {
			if synctype == msgCommon.SYNC_DATA_MAIN {
				block, err = ledger.GetLedgerStore().GetBlockByHash(hash)
			}
			if err != nil || block == nil || block.Header == nil {
				log.Debug("[p2p]can't get block send not found message by", "hash", hash)
				msg := msgpack.NewNotFound(hash)
				err := p2p.Send(remotePeer, msg, false)
				if err != nil {
					log.Warn("err", "err", err)
					return
				}
				return
			}
			saveRespCache(reqID, block)
		}
		log.Debug("[p2p]block ", "height", block.Header.Height, "hash", hash)
		if block.Sigs == nil {
			block.Sigs = &types.SigData{}
		}
		msg := msgpack.NewBlock(block, nil, orgID, synctype)

		err = p2p.Send(remotePeer, msg, false)
		if err != nil {
			log.Warn("err", "err", err)
			return
		}
	}
}

// InvHandle handles the inventory message(block,
// transaction and consensus) from peer.
func InvHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive inv message", "Addr", data.Addr, "Id", data.Id)
	var inv = data.Payload.(*msgCommon.Inv)

	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in InvHandle")
		return
	}
	if len(inv.P.Blk) == 0 {
		log.Debug("[p2p]empty inv payload in InvHandle")
		return
	}
	var id common.Hash
	str := inv.P.Blk[0].String()
	hei := inv.P.Heis[0]
	log.Debug("[p2p]the inv", "type", inv.P.InvType, "Blk.len", len(inv.P.Blk), "str", str)

	invType := common.InventoryType(inv.P.InvType)
	switch invType {
	case common.BLOCK:
		log.Debug("[p2p]receive block message")
		for _, id = range inv.P.Blk {
			log.Debug("[p2p]receive inv-block message, hash is", "id", id)
			// TODO check the ID queue
			isContainBlock, err := ledger.GetLedgerStore().ContainBlock(id)
			if err != nil {
				log.Warn("err", "err", err)
				return
			}

			if !isContainBlock && msgCommon.LastInvHash != id {
				prnnn := &msgCommon.PeerAnn{

					PeerId: data.Id,

					Height: hei,
				}

				consensPid, er := bactor.GetActorPid(bactor.CONSENSUSACTOR)
				if er != nil {
					log.Error("sync_handler.InvHandle get pid error", "error", er)
					return
				}

				consensPid.Tell(prnnn)

				msgCommon.LastInvHash = id
				// send the block request
				log.Info("[p2p]inv request block hash", "id", id)

				msg := msgpack.NewBlkDataReq(id, common.Address{}, msgCommon.SYNC_DATA_MAIN)
				err = p2p.Send(remotePeer, msg, false)
				if err != nil {
					log.Warn("err", "err", err)
					return
				}
			}
		}
	default:
		log.Warn("[p2p]receive unknown inventory message")
	}
}

// DisconnectHandle handles the disconnect events
func DisconnectHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Debug("[p2p]receive disconnect message", "Addr", data.Addr, "Id", data.Id)
	p2p.RemoveFromInConnRecord(data.Addr)
	p2p.RemoveFromOutConnRecord(data.Addr)
	remotePeer := p2p.GetPeer(data.Id)
	defer func() {
		if actor.ConsensusPid != nil {
			actor.ConsensusPid.Tell(&msgCommon.PeerDisConnect{
				ID: data.Id,
			})
		}
	}()

	if remotePeer == nil {
		log.Debug("[p2p]disconnect peer is nil")
		return
	}

	for _, orgid := range remotePeer.GetRemoteOrgs() {
		if orgid != msgCommon.StellarNodeID {
			p2p.SyncHandleSentDisconnectToBootNode(data.Id, false)
		} else {
			fmt.Println(" ******** send hengxing")
			p2p.SyncHandleSentDisconnectToBootNode(data.Id, true)
		}
	}

	p2p.RemoveFromConnectingList(data.Addr)

	if remotePeer.SyncLink.GetAddr() == data.Addr {
		p2p.RemovePeerSyncAddress(data.Addr)
		p2p.RemovePeerConsAddress(data.Addr)
		remotePeer.CloseSync()
		remotePeer.CloseCons()
	}
	if remotePeer.ConsLink.GetAddr() == data.Addr {
		p2p.RemovePeerConsAddress(data.Addr)
		remotePeer.CloseCons()
	}
}

func GetHeadersFromHeight(srcHeight uint64) ([]*types.Header, error) {
	var count uint64 = 0
	headers := []*types.Header{}
	// var startHeight uint32
	var stopHeight uint64

	curHeight := ledger.GetLedgerStore().GetCurrentHeaderHeight()
	if srcHeight == 0 {
		if curHeight > msgCommon.MAX_BLK_HDR_CNT {
			count = msgCommon.MAX_BLK_HDR_CNT
		} else {
			count = curHeight
		}
	} else {
		// bkStop, err := ledger.DefLedger.GetHeaderByHash(stopHash)
		// if err != nil || bkStop == nil {
		// 	return nil, err
		// }
		stopHeight = srcHeight

		count = curHeight - stopHeight
		if count > msgCommon.MAX_BLK_HDR_CNT {
			count = msgCommon.MAX_BLK_HDR_CNT
		}
	}

	var i uint64
	for i = 1; i <= count; i++ {
		hash, _ := ledger.GetLedgerStore().GetBlockHashByHeight(uint64(stopHeight + i))
		hd, err := ledger.GetLedgerStore().GetHeaderByHash(hash)
		if err != nil {
			log.Debug("[p2p]net_server GetBlockWithHeight failed with", "err", err.Error(), "hash", hash, "height", stopHeight+i)
			return nil, err
		}
		headers = append(headers, hd)
	}

	return headers, nil
}

//get blk hdrs from starthash to stophash
func GetHeadersFromHash(startHash common.Hash, stopHash common.Hash,
	orgid common.Address, orgtype string) ([]*types.Header, error) {
	var count uint32 = 0
	headers := []*types.Header{}
	var startHeight uint32
	var stopHeight uint32
	curHeight := uint32(0)
	if orgtype == msgCommon.SYNC_DATA_MAIN {
		curHeight = uint32(ledger.GetLedgerStore().GetCurrentHeaderHeight())
	}

	empty := common.Hash{}
	if startHash == empty {
		if stopHash == empty {
			if curHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = curHeight
			}
		} else {
			var bkStop *types.Header
			var err error
			if orgtype == msgCommon.SYNC_DATA_MAIN {
				bkStop, err = ledger.GetLedgerStore().GetHeaderByHash(stopHash)
			}

			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = uint32(bkStop.Height)
			count = curHeight - stopHeight
			if count > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			}
		}
	} else {
		var bkStart *types.Header
		var err error
		if orgtype == msgCommon.SYNC_DATA_MAIN {
			bkStart, err = ledger.GetLedgerStore().GetHeaderByHash(startHash)
		}

		if err != nil || bkStart == nil {
			return nil, err
		}
		startHeight = uint32(bkStart.Height)
		empty := common.Hash{}
		if stopHash != empty {
			var bkStop *types.Header
			var err error
			if orgtype == msgCommon.SYNC_DATA_MAIN {
				bkStop, err = ledger.GetLedgerStore().GetHeaderByHash(stopHash)
			}

			if err != nil || bkStop == nil {
				return nil, err
			}
			stopHeight = uint32(bkStop.Height)

			// avoid unsigned integer underflow
			if startHeight < stopHeight {
				return nil, errors.New("[p2p]do not have header to send")
			}
			count = startHeight - stopHeight

			if count >= msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
				stopHeight = startHeight - msgCommon.MAX_BLK_HDR_CNT
			}
		} else {

			if startHeight > msgCommon.MAX_BLK_HDR_CNT {
				count = msgCommon.MAX_BLK_HDR_CNT
			} else {
				count = startHeight
			}
		}
	}
	var i uint32
	for i = 1; i <= count; i++ {
		var (
			hash common.Hash
			hd   *types.Header
			err  error
		)
		if orgtype == msgCommon.SYNC_DATA_MAIN {
			hash, _ = ledger.GetLedgerStore().GetBlockHashByHeight(uint64(stopHeight + i))
			hd, err = ledger.GetLedgerStore().GetHeaderByHash(hash)
		}

		if err != nil {
			log.Debug("[p2p]net_server GetBlockWithHeight failed with", "err", err.Error(), "hash", hash, "height", stopHeight+i)
			return nil, err
		}
		headers = append(headers, hd)
	}

	return headers, nil
}

//getRespCacheValue get response data from cache
func getRespCacheValue(key string) interface{} {
	if respCache == nil {
		return nil
	}
	data, ok := respCache.Get(key)
	if ok {
		return data
	}
	return nil
}

//saveRespCache save response msg to cache
func saveRespCache(key string, value interface{}) bool {
	if respCache == nil {
		var err error
		respCache, err = lru.NewARC(msgCommon.MAX_RESP_CACHE_SIZE)
		if err != nil {
			return false
		}
	}
	respCache.Add(key, value)
	return true
}

// new add
//PingHandle handle ping msg from peer
func PingSpcHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive PingSpcHandle message", "Addr", data.Addr, "Id", data.Id)

	pingspc := data.Payload.(*msgCommon.PingSpc)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in PingSpcHandle")
		return
	}
	if pingspc.BNodeASrc {
		// 判断该节点是否是恒星节点
		bstellar := false
		for _, orgid := range p2p.PeerGetOrg() {
			if orgid == msgCommon.StellarNodeID {
				bstellar = true
				break
			}
		}
		// 如果是恒星节点,就去维护nodestar的状态
		if bstellar {
			pid.Tell(&msgCommon.StarConnedNodeAID{
				PeerID: data.Id,
			}) // 向p2p actor发送对方是A节点的信息
		}

		msg := msgpack.NewPongSpcMsg(pingspc.BNodeASrc, bstellar)
		err := p2p.Send(remotePeer, msg, false)
		if err != nil {
			log.Warn("err", "err", err)
		}
	} else {
		// 判断该节点是否是A节点
		bANode := p2p.SyncHandleBANode()
		msg := msgpack.NewPongSpcMsg(pingspc.BNodeASrc, bANode)
		err := p2p.Send(remotePeer, msg, false)
		if err != nil {
			log.Warn("err", "err", err)
		}
	}
}

//PongSpcHandle handle ping msg from peer
func PongSpcHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Trace("[p2p]receive PongSpcHandle message", "Addr", data.Addr, "Id", data.Id)

	pongspc := data.Payload.(*msgCommon.PongSpc)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Debug("[p2p]remotePeer invalid in PongSpcHandle")
		return
	}

	pid.Tell(&msgCommon.PingPongSpcHandler{
		BNodeASrc:     pongspc.BNodeASrc,
		BOtherSideReq: pongspc.BOtherSideReq,
		PeerID:        data.Id,
	}) // 向p2p actor发送消息
}

func EarthNotifyHashHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Info("[p2p]receive EarthNotifyHashHandle message", "Addr", data.Addr, "Id", data.Id)

	ntf := data.Payload.(*msgCommon.EarthNotifyMsg)

	log.Info("earth ntf message", "ntf", ntf)

}

func NodeLeagueHeigtNtfHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Info("[p2p]receive NodeLeagueHeigtNtfHandle message", "Addr", data.Addr, "Id", data.Id)

	if nlhcache == nil {
		initNlhCache()
	}

	nlhcache.nlhLock.Lock()
	defer nlhcache.nlhLock.Unlock()

	ntf := data.Payload.(*msgCommon.NodeLHMsg)
	uniq := ntf.Unique()
	if _, ok := nlhcache.nlh[uniq]; ok {
		return
	}
	log.Info("Node League Heigt notify Handle message", "ntf", ntf)

	radarPid, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
	if err != nil {
		log.Error("NodeLeagueHeigtNtfHandle get radar actor error", "error", err)
		return
	}

	lh := &common.NodeLH{
		ntf.NodeId,
		ntf.LeagueId,
		ntf.Height,
	}

	nlhcache.nlh[uniq] = time.Now().Unix()
	radarPid.Tell(lh)

}

// get extData request handle
func ExtDataRequestHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Info("[p2p]receive ExtDataRequestHandle message", "Addr", data.Addr, "Id", data.Id)
	var dataReq = data.Payload.(*msgCommon.ExtDataRequest)
	remotePeer := p2p.GetPeer(data.Id)
	if remotePeer == nil {
		log.Warn("[p2p]nbr node is not exist", "Id", data.Id, "Addr", data.Addr)
		return
	}
	bstellar := false
	for _, orgid := range p2p.PeerGetOrg() {
		if orgid == msgCommon.StellarNodeID {
			bstellar = true
			break
		}
	}
	// 如果是不是恒星节点不去请求
	if !bstellar {
		fmt.Println("[p2p][ExtDataRequestHandle] not stellar,exit")
		return
	}

	//get local data
	if extstorages.GetLedgerStoreInstance() == nil {
		log.Error("[p2p] extstorages.GetLedgerStoreInstance() nil")
		return
	}
	fmt.Println("🌐 🚫 get radar extData GetExtDataByHeight start")
	extData, err := extstorages.GetLedgerStoreInstance().GetExtDataByHeight(dataReq.LeagueId, dataReq.Height)
	if err != nil {
		fmt.Printf("🌐 🚫 get radar extData error:%s , height:%d , leagueId:%s \n", err, dataReq.Height, dataReq.LeagueId.ToString())
		log.Warn("[p2p]get radar extData error", "error:", err)
		return
	}
	fmt.Printf("get extData:%+v\n", *extData)
	if len(extData.AccountStates) != 0 {
		var accountStates []*extstates.EasyLeagueAccount
		for key, value := range extData.AccountStates {
			fmt.Printf("AccountStates:%v:%v\n", key, value)
			if value != nil {
				accountStates = append(accountStates, value)
			}
		}
		extData.AccountStates = accountStates
	}
	fmt.Printf("get extData:%+v\n", *extData)
	if extData.LeagueBlock == nil {
		extData.LeagueBlock = new(extstates.LeagueBlockSimple)
	}

	msg := msgpack.NewExtDataResponseMsg(extData)
	//hash,_:= common.StringToHash("")
	//msg := msgpack.NewNotFound(hash)
	err = p2p.Send(remotePeer, msg, false)
	if err != nil {
		fmt.Printf("[p2p] Sendto[%v:%v] error:%v", remotePeer.SyncLink.GetAddr(), remotePeer.SyncLink.GetPort(), err)
	}
	return
}

// get extData response handle
func ExtDataResponseHandle(data *msgCommon.MsgPayload, p2p p2p.P2P, pid *evtActor.PID, args ...interface{}) {
	log.Info("[p2p]receive ExtDataResponseHandle message", "Addr", data.Addr, "Id", data.Id)
	//defer common.CatchPanic()
	var dataReq = data.Payload.(*msgCommon.ExtDataResponse)

	if dataReq == nil {
		log.Error("[p2p]ExtDataHandle dataReq.Data is nil")
		return
	}
	if extstorages.GetLedgerStoreInstance() == nil {
		log.Error("[p2p] extstorages.GetLedgerStoreInstance() nil")
		return
	}
	blkInfo, mainTxUsed := dataReq.GetOrgBlockInfo()
	err := extstorages.GetLedgerStoreInstance().SaveAll(blkInfo, mainTxUsed)
	if err != nil {
		log.Error("[p2p]extstorages.GetLedgerStoreInstance().SaveAll", "LeagueBlock", dataReq.LeagueBlock, "MainTxUsed", dataReq.MainTxUsed)
	} else {
		radarPid, err := bactor.GetActorPid(bactor.MAINRADARACTOR)
		if err != nil {
			log.Error("NodeLeagueHeigtNtfHandle get radar actor error", "error", err)
			return
		}

		lf := &bactor.LeagueRadarCache{
			LeagueId:  blkInfo.Block.Header.LeagueId,
			BlockHash: blkInfo.Block.Hash(),
			Height:    blkInfo.Block.Header.Height,
		}
		radarPid.Tell(lf)
	}
}
