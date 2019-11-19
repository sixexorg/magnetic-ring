package message

import (
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	ct "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	msgCommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
	p2pnet "github.com/sixexorg/magnetic-ring/p2pserver/net/protocol"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstates"
)

//Version package
func NewVersion(n p2pnet.P2P, isCons bool, height uint32) msgCommon.Message {
	var version msgCommon.Version
	version.P = msgCommon.VersionPayload{
		Version:      n.GetVersion(),
		Services:     n.GetServices(),
		SyncPort:     n.GetSyncPort(),
		ConsPort:     n.GetConsPort(),
		Nonce:        n.GetID(),
		IsConsensus:  isCons,
		HttpInfoPort: n.GetHttpInfoPort(),
		StartHeight:  uint64(height),
		TimeStamp:    uint64(time.Now().UnixNano()),
	}

	if n.GetRelay() {
		version.P.Relay = 1
	} else {
		version.P.Relay = 0
	}
	if config.GlobalConfig.P2PCfg.HttpInfoPort > 0 {
		version.P.Cap[msgCommon.HTTP_INFO_FLAG] = 0x01
	} else {
		version.P.Cap[msgCommon.HTTP_INFO_FLAG] = 0x00
	}

	version.N = *(n.GetNode())
	return &version
}

//version ack package
func NewVerAck(isConsensus bool) msgCommon.Message {
	var verAck msgCommon.VerACK
	verAck.IsConsensus = isConsensus

	return &verAck
}

//Peer address request package
func NewAddrReq() msgCommon.Message {
	var msg msgCommon.AddrReq
	return &msg
}

//Peer address package // p2p
func NewAddrs(nodeAddrs []msgCommon.PeerAddr) msgCommon.Message {
	var addr msgCommon.Addr
	addr.NodeAddrs = nodeAddrs

	return &addr
}

func NewOrgPIPOMsg(org common.Address, orgheight uint64) *msgCommon.OrgPIPOInfo {
	return &msgCommon.OrgPIPOInfo{
		OrgID:  org,
		Height: orgheight,
	}
}

//ping msg package
func NewPingMsg(height uint64, orginfos []*msgCommon.OrgPIPOInfo, pinginfotype string) *msgCommon.Ping {
	var ping msgCommon.Ping
	ping.Height = uint64(height)
	ping.InfoType = pinginfotype
	ping.OrgInfo = make([]msgCommon.OrgPIPOInfo, 0)
	for _, info := range orginfos {
		ping.OrgInfo = append(ping.OrgInfo, *info)
	}
	return &ping
}

//pong msg package
func NewPongMsg(height uint64, orginfos []*msgCommon.OrgPIPOInfo, pinginfotype string) *msgCommon.Pong {
	var pong msgCommon.Pong
	pong.Height = uint64(height)
	pong.InfoType = pinginfotype
	pong.OrgInfo = make([]msgCommon.OrgPIPOInfo, 0)
	for _, info := range orginfos {
		pong.OrgInfo = append(pong.OrgInfo, *info)
	}

	return &pong
}

func NewHeadersReq(curHdrHash common.Hash, orgID common.Address, synctype string, height ...uint64) msgCommon.Message {
	var h msgCommon.HeadersReq
	h.Len = 1
	h.HashEnd = curHdrHash
	if len(height) > 0 {
		h.Height = height[0]
	}
	// h.Height = height
	h.OrgID = orgID
	h.SyncType = synctype

	return &h
}

//blk hdr package
func NewHeaders(headers []*ct.Header, orgheaders []*orgtypes.Header, orgID common.Address, synctype string) msgCommon.Message {
	var blkHdr msgCommon.BlkHeader
	blkHdr.BlkHdr = headers
	blkHdr.OrgID = orgID
	blkHdr.SyncType = synctype
	blkHdr.BlkOrgHdr = orgheaders
	return &blkHdr
	// return &blkHdr
	//return msgCommon.BlkHeaderToBlkP2PHeader(&blkHdr)
}

//InvPayload // used
func NewInvPayload(invType common.InventoryType, msg []common.Hash, heis []uint64) *msgCommon.InvPayload {
	return &msgCommon.InvPayload{
		InvType: invType,
		Blk:     msg,
		Heis:    heis,
	}
}

//Inv request package // used
func NewInv(invPayload *msgCommon.InvPayload) msgCommon.Message {
	var inv msgCommon.Inv
	inv.P.Blk = invPayload.Blk
	inv.P.Heis = invPayload.Heis
	inv.P.InvType = invPayload.InvType

	return &inv
}

//transaction request package // not used
func NewTxnDataReq(hash common.Hash) msgCommon.Message {
	var dataReq msgCommon.DataReq
	dataReq.DataType = common.TRANSACTION
	dataReq.Hash = hash

	return &dataReq
}

//block request package // used
func NewBlkDataReq(hash common.Hash, orgID common.Address, synctype string) msgCommon.Message {
	var dataReq msgCommon.DataReq
	dataReq.DataType = common.BLOCK
	dataReq.Hash = hash
	dataReq.OrgID = orgID
	dataReq.SyncType = synctype

	return &dataReq
}

//consensus request package // not used
func NewConsensusDataReq(hash common.Hash) msgCommon.Message {
	var dataReq msgCommon.DataReq
	dataReq.DataType = common.CONSENSUS
	dataReq.Hash = hash

	return &dataReq
}

///block package // used
func NewBlock(bk *ct.Block, bkorg *orgtypes.Block, orgID common.Address, synctype string) msgCommon.Message {
	var blk msgCommon.Block
	blk.Blk = bk
	blk.BlkOrg = bkorg
	blk.OrgID = orgID
	blk.SyncType = synctype

	// return &blk
	return msgCommon.BlockToTrsBlock(&blk)
}

func NewBlockRdr(bk *ct.Block, bkorg *orgtypes.Block, orgID common.Address, synctype string) msgCommon.Message {
	var blk msgCommon.Block
	blk.Blk = bk
	blk.BlkOrg = bkorg
	blk.OrgID = orgID
	blk.SyncType = synctype

	// return &blk
	return msgCommon.BlockToRdrBlock(&blk)
}

//Consensus info package // used
func NewConsensus(cp *msgCommon.ConsensusPayload) msgCommon.Message {
	var cons msgCommon.Consensus
	cons.Cons = *cp

	// return &cons
	return msgCommon.ConsensusToP2PConsPld(&cons)
}

//NotFound package // used
func NewNotFound(hash common.Hash) msgCommon.Message {
	var notFound msgCommon.NotFound
	notFound.Hash = hash

	return &notFound
}

//Transaction package // used
func NewTxn(txn *ct.Transaction, orgtxn *orgtypes.Transaction, orgID common.Address, synctype string) msgCommon.Message {
	var trn msgCommon.Trn
	trn.Txn = txn
	trn.OrgTxn = orgtxn
	trn.OrgID = orgID
	trn.SyncType = synctype

	// return &trn
	return msgCommon.TrnToP2PTrn(&trn)
}

// new add
//ping msg package
func NewPingSpcMsg(bNodeASrc bool) *msgCommon.PingSpc {
	var pingspc msgCommon.PingSpc
	pingspc.BNodeASrc = bNodeASrc
	return &pingspc
}

//pong msg package
func NewPongSpcMsg(bNodeASrc bool, bOtherSideReq bool) *msgCommon.PongSpc {
	var pongspc msgCommon.PongSpc
	pongspc.BNodeASrc = bNodeASrc
	pongspc.BOtherSideReq = bOtherSideReq

	return &pongspc
}

//Transaction package // used
func NewEarthNotifyHash(earthnotify *msgCommon.EarthNotifyBlk) msgCommon.Message {
	var enh msgCommon.EarthNotifyMsg
	enh.Height = earthnotify.BlkHeight
	enh.Hash = earthnotify.BlkHash

	// return &trn
	return &enh
}

func NewNodeLHMsg(earthnotify *common.NodeLH) msgCommon.Message {
	var nodeLHMsg msgCommon.NodeLHMsg
	nodeLHMsg.Height = earthnotify.Height
	nodeLHMsg.LeagueId = earthnotify.LeagueId
	nodeLHMsg.NodeId = earthnotify.NodeId

	return &nodeLHMsg
}

func NewExtDataRequestMsg(data *msgCommon.ExtDataRequest) msgCommon.Message {
	var extMsg msgCommon.ExtDataRequest
	extMsg.LeagueId = data.LeagueId
	extMsg.Height = data.Height

	return &extMsg
}

func NewExtDataResponseMsg(data *extstates.ExtData) msgCommon.Message {
	extData := *data
	extMsg := msgCommon.ExtDataResponse{
		&extData,
	}
	return &extMsg
}
