package spcnode

import (
	comm "github.com/sixexorg/magnetic-ring/p2pserver/common"
	discover "github.com/sixexorg/magnetic-ring/p2pserver/discover"
	"github.com/sixexorg/magnetic-ring/p2pserver/peer"
)

type SpectNodeP2PSer interface {
	GetNode(id uint64) *peer.Peer
	Send(p *peer.Peer, msg comm.Message, isConsensus bool) error
	HandlerANodeConn(idm map[discover.NodeID]bool)
}
