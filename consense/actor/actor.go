package actor

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/log"
	netActor "github.com/sixexorg/magnetic-ring/p2pserver/actor/server"
	ptypes "github.com/sixexorg/magnetic-ring/p2pserver/common"
)

type P2PActor struct {
	P2P *actor.PID
}

func (self *P2PActor) Broadcast(msg interface{}) {

	p2ppid,err := bactor.GetActorPid(bactor.P2PACTOR)
	if err != nil {
		log.Error("get p2ppid error","error",err)
	}

	p2ppid.Tell(msg)
}

func (self *P2PActor) Transmit(target uint64, msg ptypes.Message) {

	p2ppid,err := bactor.GetActorPid(bactor.P2PACTOR)
	if err != nil {
		log.Error("get p2ppid error","error",err)
	}

	p2ppid.Tell(&netActor.TransmitConsensusMsgReq{
		Target: target,
		Msg:    msg,
	})
}

type LedgerActor struct {
	Ledger *actor.PID
}
