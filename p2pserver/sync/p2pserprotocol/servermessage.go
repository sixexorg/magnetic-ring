package p2pserprotocol

import (
	"github.com/sixexorg/magnetic-ring/common"
	comm "github.com/sixexorg/magnetic-ring/p2pserver/common"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"
	ledger "github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

type SyncP2PSer interface {
	GetNodeFromDiscoverID(discoverNodeID string) *peer.Peer
	GetNode(id uint64) *peer.Peer
	ReachMinConnection() bool
	Send(p *peer.Peer, msg comm.Message, isConsensus bool) error
	GetLedger() *ledger.LedgerStoreImp
	PingTo(peers []*peer.Peer, bTimer bool, orgID ...common.Address)
	SentConnectToBootNode(remoteid uint64)
	SentDisconnectToBootNode(remoteid uint64, bstellar bool)
	Xmit(message interface{}) error
}
