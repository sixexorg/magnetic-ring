package actorstart

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/radar/cycleactor"
	"github.com/sixexorg/magnetic-ring/radar/mainchain"
)

func InitCycleActor(p2pActor bactor.Teller) *cycleactor.CycleActor {
	cycle := cycleactor.NewCycleActor(p2pActor)
	props := actor.FromProducer(func() actor.Actor {
		return cycle
	})
	cycleActor := actor.Spawn(props)
	bactor.RegistActorPid(bactor.CYCLEACTOR, cycleActor)
	return cycle
}
func InitMainRadarActor() bactor.Teller {
	props := actor.FromProducer(func() actor.Actor { return &mainchain.MainRadarActor{} })
	mainRadar := actor.Spawn(props)
	bactor.RegistActorPid(bactor.MAINRADARACTOR, mainRadar)
	return mainRadar
}
