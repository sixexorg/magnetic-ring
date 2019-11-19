package p2p

import (
	"crypto/ecdsa"

	comm "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/p2pserver/common"
	"github.com/sixexorg/magnetic-ring/p2pserver/discover"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"
)

//P2P represent the net interface of p2p package
type P2P interface {
	Start()
	Halt()
	Connect(addr string, isConsensus bool, node *discover.Node, bANode bool, orgids ...comm.Address) error
	GetID() uint64
	GetVersion() uint32
	GetSyncPort() uint16
	GetConsPort() uint16
	GetHttpInfoPort() uint16
	GetRelay() bool
	GetHeight() uint64
	GetTime() int64
	GetServices() uint64
	GetNeighbors() []*peer.Peer
	GetNeighborAddrs() []common.PeerAddr
	GetConnectionCnt() uint32
	GetNp() *peer.NbrPeers
	GetPeer(uint64) *peer.Peer
	SetHeight(uint64)
	IsPeerEstablished(p *peer.Peer) bool
	Send(p *peer.Peer, msg common.Message, isConsensus bool) error
	GetMsgChan(isConsensus bool) chan *common.MsgPayload
	GetPeerFromAddr(addr string) *peer.Peer
	AddOutConnectingList(addr string) (added bool)
	GetOutConnRecordLen() int
	RemoveFromConnectingList(addr string)
	RemoveFromOutConnRecord(addr string)
	RemoveFromInConnRecord(addr string)
	AddPeerSyncAddress(addr string, p *peer.Peer)
	AddPeerConsAddress(addr string, p *peer.Peer)
	GetOutConnectingListLen() (count uint)
	RemovePeerSyncAddress(addr string)
	RemovePeerConsAddress(addr string)
	AddNbrNode(*peer.Peer)
	DelNbrNode(id uint64) (*peer.Peer, bool)
	NodeEstablished(uint64) bool
	Xmit(msg common.Message, isCons bool)
	SetOwnAddress(addr string)
	IsAddrFromConnecting(addr string) bool
	// org
	HaltOrdinaryPeerID(id uint64)
	HaltOrgPeerID(peerid uint64, orgid comm.Address)
	SetOwnOrgHeight(height uint64, id comm.Address)
	GetOwnOrgHeight(id comm.Address) uint64
	PeerAddOrg(id comm.Address)
	PeerDelOrg(id comm.Address)
	PeerGetOrg() []comm.Address
	PeerGetRealOrg() []comm.Address
	RemoteDelOrg(id comm.Address) []uint64
	BHaveOrgs() bool
	BHaveOrgsExceptId(id comm.Address) bool
	BHaveOrgsId(id comm.Address) bool
	SyncHandleSentDisconnectToBootNode(remoteid uint64, bstellar bool)
	SyncHandleBANode() bool
	//
	SetPrivateKey(nodeKey *ecdsa.PrivateKey, node *discover.Table)
	SetBootNodes(nodes []*discover.Node)
	GetNode() *discover.Node
}
