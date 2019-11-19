package cycleactor

import (
	"sync"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/common"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/node"
	p2pcommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
)

type CycleActor struct {
	p2p       bactor.Teller
	mainRadar bactor.Teller
	start     bool
	m         sync.Mutex
}

func NewCycleActor(p2pActor bactor.Teller) *CycleActor {
	return &CycleActor{
		p2p: p2pActor,
	}
}

var (
	leagueBlockPool = make(map[common.Hash]int64)
	t               = common.NewTimingWheel(time.Second, 10)
	m               sync.RWMutex
)

func (this *CycleActor) SetMainRadarActor(mainRadarActor bactor.Teller) {
	this.mainRadar = mainRadarActor
}
func (this *CycleActor) Receive(context actor.Context) {
	m.Lock()
	defer m.Unlock()
	switch msg := context.Message().(type) {
	case *orgtypes.Block:
		blkHash := msg.Hash()
		_, ok := leagueBlockPool[blkHash]
		if !ok {
			leagueBlockPool[blkHash] = time.Now().Add(time.Minute * 2).Unix()
			opd := &p2pcommon.OrgPendingData{
				BANodeSrc: true,                // true:ANode send staller false:staller send staller
				OrgId:     msg.Header.LeagueId, // orgid
				Block:     msg,                 // tx
			}
			if node.IsStar() {
				opd.BANodeSrc = false
				this.mainRadar.Tell(msg)
			}
			this.p2p.Tell(opd)
		}
	}
}
func init() {
	go func() {
		for {
			select {
			case <-t.After(time.Second * 5):
				time := time.Now().Unix()
				for k, v := range leagueBlockPool {
					if v < time {
						delete(leagueBlockPool, k)
					}
				}
			}
		}
	}()
}
