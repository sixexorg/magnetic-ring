package mainchain

import (
	"fmt"
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/common"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/node"
)

type MainRadarActor struct {
	mainRadar *LeagueConsumers
}

func (mra *MainRadarActor) SetMainRadar(mainRadar *LeagueConsumers) {
	mra.mainRadar = mainRadar
}
func (mra *MainRadarActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *orgtypes.Block:
		fmt.Println("🔆 📡 📩 receive leagueBlock")
		leagueConsumers.ReceiveBlock(msg)
	case *common.NodeLH:
		//get all stars's league processing progress
		nodeLHC.receiveNodeLH(msg)
	case []string:
		nodeLHC.refreshStarNode(msg)
	case uint64:
		//to refresh stars node
		stars := node.GurStars()
		nodeLHC.checkNodeRoot(stars)
	case *bactor.LeagueRadarCache:
		//SupplementLeagueConsumer is used to synchronize lost cross-chain storage
		leagueConsumers.SupplementLeagueConsumer(msg.LeagueId, msg.BlockHash, msg.Height)
	default:
		fmt.Printf("container txpool-tx actor: unknown msg %v type %v\n", msg, reflect.TypeOf(msg))
	}
}
