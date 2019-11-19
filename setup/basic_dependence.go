/*
Service registration order
1. config
2. bootNode
3. nodeAccount

*/
package setup

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/urfave/cli"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/password"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/meacount"
	"github.com/sixexorg/magnetic-ring/p2pserver"
	p2pactor "github.com/sixexorg/magnetic-ring/p2pserver/actor/server"
	"github.com/sixexorg/magnetic-ring/store/mainchain/genesis"
	mainstore "github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	mainval "github.com/sixexorg/magnetic-ring/store/mainchain/validation"
)

//step1
func initConfig() {
	p, err := os.Getwd()
	if err != nil {
		return
	}
	if strings.Contains(p, "rennbon") {
		//err = config.InitConfig(path.Join(p, "crystal-node/go-crystal/config.yml"))
		err = config.InitConfig(path.Join(p, "./config.yml"))
	} else if strings.Contains(p, "rudy") {
		//err = config.InitConfig(path.Join(p, "crystal-node/go-crystal/config.yml"))
		err = config.InitConfig(path.Join(p, "./config.yml"))
	} else if strings.Contains(p, "dachu") {
		err = config.InitConfig(path.Join(p, "./config.yml"))
	} else {
		err = config.InitConfig(path.Join(p, "./config.yml"))
	}

	if err != nil {
		fmt.Printf("init config error-->%v\n", err)
		os.Exit(0)
		return
	}
}

//step3
func initNodeAccount(ctx *cli.Context) account.Account {
	accountPath := ctx.String("wallet")
	mgr, err := account.NewManager()
	if err != nil {
		panic(err)
	}
	pasd := ctx.String("password")
	if pasd == "" {
		pasw, err := password.GetPassword()
		if err != nil {
			os.Exit(10)
		}
		pasd = string(pasw)
	}
	address, err := mgr.ImportNormalAccount(accountPath, pasd)
	if err != nil {
		log.Error("import normal account error", "error", err)
		os.Exit(9)
	}
	acct, err := mgr.GetNormalAccount(address, pasd)
	if err != nil {
		log.Error("get normal account error", "error", err)
		os.Exit(8)
	}
	nodeimpl := acct.(account.Account)
	meacount.SetOwner(nodeimpl)
	return acct
}

//step4
func initLog(logpath string) {
	//logger = log.New("module", "main")
	//logger.SetHandler(log.MultiHandler(
	//	log.StderrHandler,
	//	log.LvlFilterHandler(
	//		log.LvlInfo,
	//		log.Must.FileHandler("errors.json", log.LogfmtFormat()))))
	//log.Info("log init completely")
	log.InitMagneticLog(logpath)
	log.PrintOrigins(true)
}

//step5
func initLedger(dbDir string) *mainstore.LedgerStoreImp {
	ledger, err := mainstore.NewLedgerStore(dbDir)
	if err != nil {
		panic(fmt.Errorf("func initLedger err %s", err))
	}
	mainval.NewAccountNonceCache(func(height uint64, account common.Address) uint64 {
		ass, err := ledger.GetAccountByHeight(height, account)
		if err != nil {
			return 0
		}
		return ass.Data.Nonce
	})
	return ledger
}

//step6
func mainchainGenesis(conf *config.GenesisConfig, ledger *mainstore.LedgerStoreImp) {
	genesis, err := genesis.InitGenesis(conf)
	if err != nil {
		panic(err)
	}
	if ledger.GetCurrentBlockHeight() == 0 {
		err := genesis.GenesisBlock(ledger)
		if err != nil {
			panic(err)
		}
	}
}

//step7
func initP2P() *actor.PID {
	p2p := p2pserver.NewServer()
	p2pActor := p2pactor.NewP2PActor(p2p)
	p2pPID, err := p2pActor.Start()
	if err != nil {
		panic(fmt.Errorf("p2pActor init error %s", err))
	}
	p2p.SetPID(p2pPID)
	bactor.RegistActorPid(bactor.P2PACTOR, p2pPID)
	err = p2p.Start()
	if err != nil {
		panic(fmt.Errorf("initP2P err %s", err))
	}
	p2p.WaitForPeersStart()
	return p2pPID
}
