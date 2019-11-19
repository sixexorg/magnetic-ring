package bactor

import (
	"sync"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/errors"
)

var (
	actorMap = map[BactorType]*actor.PID{}
	am       sync.Mutex
)

const (
	TXPOOLACTOR         = BactorType("TXPOOLACTOR")
	CONSENSUSACTOR      = BactorType("CONSENACTOR")
	P2PACTOR            = BactorType("P2PACTOR")
	MAINRADARACTOR      = BactorType("NOTIFY2P2PACTOR")
	CYCLEACTOR          = BactorType("CYCLEACTOR")
	STARTBROADCASTSATOR = BactorType("STARTBROADCAST2PACTOR")
)

type BactorType string

func RegistActorPid(module BactorType, pid *actor.PID) {
	am.Lock()
	defer am.Unlock()
	actorMap[module] = pid
}
func GetActorPid(module BactorType) (*actor.PID, error) {
	am.Lock()
	defer am.Unlock()
	var pid *actor.PID
	var ok bool

	if pid, ok = actorMap[module]; !ok {
		return nil, errors.ERR_ACTOR_NOTREG
	}
	return pid, nil
}
