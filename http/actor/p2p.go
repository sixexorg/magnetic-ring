package httpactor

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/common"
	p2pcmn "github.com/sixexorg/magnetic-ring/p2pserver/common"
)

var
(
	p2pPid *actor.PID
)

func SetTP2pPid(pid *actor.PID)  {
	txpoolPid = pid
}

func TellP2pSync(leagueId common.Address,BAdd bool){

	syncobj := p2pcmn.HandleOrg{
		leagueId,BAdd,
	}

	txpoolPid.Tell(&syncobj)
}

