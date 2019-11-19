package setup

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
	"github.com/sixexorg/magnetic-ring/bactor/actorstart"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/mock"
	p2pcommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
	"fmt"
)

//before Start(),need init global config
//ordinary:
//	1: bootnode
//  2: log
//  3: node account
//  4: main store
//  5: main block genesis
//  6: p2p and p2p-actor
//  7: cycle-actor  in order to transit orgblock
//  8: assign p2p-actor in the cycle-actor
//  9: txpool and txpool-actor
//star:
// 	1: the store across the chain
//	2: main radar
//  3: assign mainRadar in the txpool
//	4: mainRadar-actor
//  5: consensus
//exposed api:
//	1:runRouter
//league:
//  1: containers just nitialize once
//  2: createContainers
func Start(ctx *cli.Context) {
	//1
	p2pcommon.InitBootNode(config.GlobalConfig.BootNode.IP)
	//2
	initLog(config.GlobalConfig.SysCfg.LogPath)
	//3
	acct := initNodeAccount(ctx) //The node public private key needs to be bound to the account public private key. If it is a request for a star node.
	//4
	ledger := initLedger(config.GlobalConfig.SysCfg.StoreDir)
	//5
	mainchainGenesis(&config.GlobalConfig.Genesis, ledger)
	//6
	fmt.Println("before &&&&&&------->>>>>>>>>>>RestoreVars")
	p2pActor := initP2P()
	fmt.Println("after &&&&&&------->>>>>>>>>>>RestoreVars")
	//7  to pass league block to cross-chain validation,
	cycle := actorstart.InitCycleActor(p2pActor) // Broadcasting a sub-chain block to a star node. Adding this ordinary node does not require starting the star node radar.
	//8  to check mainchain tx's nonce,balance and more
	mainPool := initTxPool()
	if true {// If it is a star node
		//1 Cross-chain radar storage
		acrossLedger := initAcrossStore(config.GlobalConfig.SysCfg.StoreDir)
		//2 the star need check leagues's block Main chain radar stellar node consumption A node block data
		mainRadar := initMainRadar(acct, ledger, acrossLedger)
		//3 decorate txpool with mainRadar,then it can execute the data from mainRadar. txpool need to register the radar of the public chain, the radar will generate unsigned transactions.
		mainPool.SetMainRadar(mainRadar)
		//4 the mainRadar has data source after that
		mainActor := actorstart.InitMainRadarActor()
		//5 turn on this switch,the cycelRadar can push orgBlock to cross-chain validation
		cycle.SetMainRadarActor(mainActor)
		ledger.RestoreVars()

		//4 local can package blocks
		initConsensus(p2pActor, acct)
	}
	//ledger.RestoreVars()
	//for http
	go runRouter(config.GlobalConfig.SysCfg.HttpPort)

	//for org-chain
	if true {
		initContainers(p2pActor)
		go func() {
			if config.GlobalConfig.SysCfg.GenBlock {
				CreateContainer(mock.AccountNormal_1, 1, config.GlobalConfig.SysCfg.StoreDir, &config.GlobalConfig.CliqueCfg)
			}
		}()
	}
	waitFor()
}
func waitFor() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-c
	log.Info("server shut down")
}
