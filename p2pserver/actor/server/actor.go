package server

import (
	"reflect"

	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	"github.com/sixexorg/magnetic-ring/meacount"

	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/p2pserver"
	"github.com/sixexorg/magnetic-ring/p2pserver/common"
)

type P2PActor struct {
	props  *actor.Props
	server *p2pserver.P2PServer
}

// NewP2PActor creates an actor to handle the messages from
// txnpool and consensus
func NewP2PActor(p2pServer *p2pserver.P2PServer) *P2PActor {
	return &P2PActor{
		server: p2pServer,
	}
}

//start a actor called net_server
func (this *P2PActor) Start() (*actor.PID, error) {
	this.props = actor.FromProducer(func() actor.Actor { return this })
	p2pPid, err := actor.SpawnNamed(this.props, "net_server")
	return p2pPid, err
}

//message handler
func (this *P2PActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Restarting:
		log.Warn("[p2p]actor restarting")
	case *actor.Stopping:
		log.Warn("[p2p]actor stopping")
	case *actor.Stopped:
		log.Warn("[p2p]actor stopped")
	case *actor.Started:
		log.Debug("[p2p]actor started")
	case *actor.Restart:
		log.Warn("[p2p]actor restart")
	case *StopServerReq:
		this.handleStopServerReq(ctx, msg)
	case *GetPortReq:
		this.handleGetPortReq(ctx, msg)
	case *GetVersionReq:
		this.handleGetVersionReq(ctx, msg)
	case *GetConnectionCntReq:
		this.handleGetConnectionCntReq(ctx, msg)
	case *GetSyncPortReq:
		this.handleGetSyncPortReq(ctx, msg)
	case *GetConsPortReq:
		this.handleGetConsPortReq(ctx, msg)
	case *GetIdReq:
		this.handleGetIDReq(ctx, msg)
	case *GetConnectionStateReq:
		this.handleGetConnectionStateReq(ctx, msg)
	case *GetTimeReq:
		this.handleGetTimeReq(ctx, msg)
	case *GetNeighborAddrsReq:
		this.handleGetNeighborAddrsReq(ctx, msg)
	case *GetRelayStateReq:
		this.handleGetRelayStateReq(ctx, msg)
	case *GetNodeTypeReq:
		this.handleGetNodeTypeReq(ctx, msg)
	case *TransmitConsensusMsgReq:
		this.handleTransmitConsensusMsgReq(ctx, msg)
	case *common.AppendPeerID:
		this.server.OnAddNode(msg.ID)
	case *common.RemovePeerID:
		this.server.OnDelNode(msg.ID)
	case *common.AppendHeaders: // recv headrs sync main and org
		if msg.SyncType == common.SYNC_DATA_MAIN {
			this.server.OnHeaderReceive(msg.FromID, msg.Headers)
		} else if msg.SyncType == common.SYNC_DATA_ORG {
			this.server.OrgHeaderReceive(msg.FromID, msg.HeaderOrgs, msg.OrgID)
		}
	case *common.AppendBlock: // recv blocks sync main and org
		if msg.SyncType == common.SYNC_DATA_MAIN {
			this.server.OnBlockReceive(msg.FromID, msg.BlockSize, msg.Block)
		} else if msg.SyncType == common.SYNC_DATA_ORG {
			this.server.OrgBlockReceive(msg.FromID, msg.BlockSize, msg.BlockOrg, msg.OrgID)
		}
	case *common.HandleOrg:
		//if msg.BAdd {
		//	this.server.AddOrg(msg.OrgID)
		//} else {
		//	this.server.DelOrg(msg.OrgID)
		//}
	case *common.HandleAddDelOrgPeerID:
		if msg.BAdd {
			for _, orgid := range msg.OrgID {
				this.server.OrgAddNode(msg.PeerID, orgid)
			}
		} else {
			for _, orgid := range msg.OrgID {
				this.server.OrgDelNode(msg.PeerID, orgid)
			}
		}
	case *common.StellarNodeConnInfo:
		log.Debug(" ****** msg.BAdd:", "badd", msg.BAdd)
		if msg.BAdd {
			this.server.StellarNodeAdd()
		} else {
			this.server.StellarNodeDel()
		}
	case *common.StellarAddDelPeerID:
		log.Debug(" ****** msg.BAdd,msg.PeerID", "badd", msg.BAdd, "peerid", msg.PeerID)
		if msg.BAdd {
			this.server.RemoteAddStellarNode(msg.PeerID)
		} else {
			this.server.RemoteDelStellarNode(msg.PeerID)
		}
	case *common.ANodeConnInfo:
		fmt.Println("⭕ 📩 🌐 set leagueId into node,the orgId is", msg.OrgId.ToString())
		if msg.BAdd {
			if !this.server.AddANode(msg.OrgId) {
				log.Error("P2PActor Receive add ANode err", "OrgId", msg.OrgId)
			}
		} else {
			this.server.DelANode(msg.OrgId)
		}
	case *common.PingPongSpcHandler:
		this.server.HandlerPingPongSpc(msg)
	case *common.StarConnedNodeAID:
		this.server.StarConnedNodeAInfo(msg.PeerID)
	case *common.OrgPendingData:

		if msg.BANodeSrc {
			fmt.Println("🌐  p2p actor ⭕  block send to 🔆  for verification")
			this.server.AToStellarPendingData(msg)
		} else {
			fmt.Println("🌐  p2p actor 🔆️  block send to 🔆  for verification,will be occure timeout begin😤")
			this.server.StellarToStellarPendingData(msg)
		}
	case *comm.ConsenseNotify:
		this.handleSyncReqConsensusState(ctx, msg)
	default:
		err := this.server.Xmit(ctx.Message())
		if nil != err {
			log.Warn("[p2p]error xmit message", "err", err.Error(), "Message", reflect.TypeOf(ctx.Message()))
		}
	}
}

//stop handler
func (this *P2PActor) handleStopServerReq(ctx actor.Context, req *StopServerReq) {
	this.server.Stop()
	if ctx.Sender() != nil {
		resp := &StopServerRsp{}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//get port handler
func (this *P2PActor) handleGetPortReq(ctx actor.Context, req *GetPortReq) {
	syncPort, consPort := this.server.GetPort()
	if ctx.Sender() != nil {
		resp := &GetPortRsp{
			SyncPort: syncPort,
			ConsPort: consPort,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//version handler
func (this *P2PActor) handleGetVersionReq(ctx actor.Context, req *GetVersionReq) {
	version := this.server.GetVersion()
	if ctx.Sender() != nil {
		resp := &GetVersionRsp{
			Version: version,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//connection count handler
func (this *P2PActor) handleGetConnectionCntReq(ctx actor.Context, req *GetConnectionCntReq) {
	cnt := this.server.GetConnectionCnt()
	if ctx.Sender() != nil {
		resp := &GetConnectionCntRsp{
			Cnt: cnt,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//sync port handler
func (this *P2PActor) handleGetSyncPortReq(ctx actor.Context, req *GetSyncPortReq) {
	var syncPort uint16
	//TODO
	if ctx.Sender() != nil {
		resp := &GetSyncPortRsp{
			SyncPort: syncPort,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//consensus port handler
func (this *P2PActor) handleGetConsPortReq(ctx actor.Context, req *GetConsPortReq) {
	var consPort uint16
	//TODO
	if ctx.Sender() != nil {
		resp := &GetConsPortRsp{
			ConsPort: consPort,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//get id handler
func (this *P2PActor) handleGetIDReq(ctx actor.Context, req *GetIdReq) {
	id := this.server.GetID()
	if ctx.Sender() != nil {
		resp := &GetIdRsp{
			Id: id,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//connection state handler
func (this *P2PActor) handleGetConnectionStateReq(ctx actor.Context, req *GetConnectionStateReq) {
	state := this.server.GetConnectionState()
	if ctx.Sender() != nil {
		resp := &GetConnectionStateRsp{
			State: state,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//timestamp handler
func (this *P2PActor) handleGetTimeReq(ctx actor.Context, req *GetTimeReq) {
	time := this.server.GetTime()
	if ctx.Sender() != nil {
		resp := &GetTimeRsp{
			Time: time,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//nbr peer`s address handler
func (this *P2PActor) handleGetNeighborAddrsReq(ctx actor.Context, req *GetNeighborAddrsReq) {
	addrs := this.server.GetNeighborAddrs()
	if ctx.Sender() != nil {
		resp := &GetNeighborAddrsRsp{
			Addrs: addrs,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//peer`s relay state handler
func (this *P2PActor) handleGetRelayStateReq(ctx actor.Context, req *GetRelayStateReq) {
	ret := this.server.GetNetWork().GetRelay()
	if ctx.Sender() != nil {
		resp := &GetRelayStateRsp{
			Relay: ret,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

//peer`s service type handler
func (this *P2PActor) handleGetNodeTypeReq(ctx actor.Context, req *GetNodeTypeReq) {
	ret := this.server.GetNetWork().GetServices()
	if ctx.Sender() != nil {
		resp := &GetNodeTypeRsp{
			NodeType: ret,
		}
		ctx.Sender().Request(resp, ctx.Self())
	}
}

func (this *P2PActor) handleTransmitConsensusMsgReq(ctx actor.Context, req *TransmitConsensusMsgReq) {
	peer := this.server.GetNetWork().GetPeer(req.Target)
	if peer != nil {
		this.server.Send(peer, req.Msg, true)
	} else {
		log.Warn("[p2p]can`t transmit consensus msg:no valid neighbor", "peer", req.Target)
	}
}

func (this *P2PActor) handleSyncReqConsensusState(ctx actor.Context, req *comm.ConsenseNotify) {
	fmt.Printf("handleSyncReqConsensusState:%+v\n", req)

	if !req.Istart {
		this.server.SetSyncStatus(false, 0)
		return
	}
	act := meacount.GetOwner()
	if act == nil {
		log.Error("handleSyncReqConsensusState meacount.GetOwner is nil")
		return
	}

	ndactImpl := act.(*account.NormalAccountImpl)
	pubkstr := fmt.Sprintf("%x", ndactImpl.PublicKey().Bytes())

	for _, pubk := range req.ProcNodes {
		if pubk == pubkstr {
			this.server.SetSyncStatus(true, req.BlkNum)
			return
		}
	}
}
