package actor

import "github.com/ontio/ontology-eventbus/actor"

var ConsensusPid *actor.PID

func SetConsensusPid(conPid *actor.PID) {
	ConsensusPid = conPid
}

type CurHeight struct {
	Height uint64
}
