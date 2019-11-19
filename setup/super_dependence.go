package setup

import (
	"fmt"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/consense/dpoa"
	netreqactor "github.com/sixexorg/magnetic-ring/p2pserver/actor/req"
	"github.com/sixexorg/magnetic-ring/radar/mainchain"
	storeactor "github.com/sixexorg/magnetic-ring/store/mainchain/actor"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstorages"
	mainstore "github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/orgchain/validation"
	txpool "github.com/sixexorg/magnetic-ring/txpool/mainchain"
)

func initAcrossStore(dbDir string) *extstorages.LedgerStoreImp {
	extLedger, err := extstorages.NewLedgerStore(dbDir)
	if err != nil {
		panic(fmt.Errorf("init extLedger error ", err))
	}
	return extLedger
}

func initMainRadar(nodeAcc account.Account, ledger *mainstore.LedgerStoreImp, extLedger *extstorages.LedgerStoreImp) *mainchain.LeagueConsumers {
	nodeacct := nodeAcc.(account.NormalAccount)
	adpter := &mainchain.ConsumerAdapter{
		FuncGenesis:  ledger.GetTxByLeagueId,
		FuncTx:       ledger.GetTxByHash,
		FuncBlk:      ledger.GetBlockHeaderByHeight,
		Ledger:       extLedger,
		FuncValidate: validation.ValidateBlock,
		NodeId:       nodeacct.PublicKey().Hex(),
	}
	mainchain.NewNodeLHCache()
	return mainchain.NewLeagueConsumers(adpter)
}

func initConsensus(p2pPid *actor.PID, nodeAcc account.Account) {
	nmlAcct := nodeAcc.(account.NormalAccount)
	consensusService, err := dpoa.NewServer(nmlAcct, p2pPid)
	if err != nil {
		panic(fmt.Errorf("init consensus error ", err))
	}
	consensusService.Start()
	netreqactor.SetConsensusPid(consensusService.GetPID())
	storeactor.SetConsensusPid(consensusService.GetPID())
}

func initTxPool() *txpool.TxPool {
	pool, err := txpool.InitPool()
	if err != nil {
		panic(fmt.Errorf("init txpool err %s", err))
	}
	netreqactor.SetTxnPoolPid(pool.GetTxpoolPID())
	return pool
}
