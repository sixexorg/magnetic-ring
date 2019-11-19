package setup

import (
	"fmt"

	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/http/actor"
	"github.com/sixexorg/magnetic-ring/http/router"
)

func runRouter(httpPort int) {
	router.StartRouter()
	engin := router.GetRouter()
	txPoolActor, err := bactor.GetActorPid(bactor.TXPOOLACTOR)
	if err != nil {
		panic(err)
	}
	httpactor.SetTxPoolPid(txPoolActor)
	addr := fmt.Sprintf(":%d", httpPort)
	engin.Run(addr)
}
