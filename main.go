package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/sixexorg/magnetic-ring/meacount"

	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/consense/poa"
	"github.com/sixexorg/magnetic-ring/http/actor"
	"github.com/sixexorg/magnetic-ring/miner"

	password2 "github.com/sixexorg/magnetic-ring/common/password"

	"github.com/sixexorg/magnetic-ring/orgcontainer"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/urfave/cli"
	"github.com/sixexorg/magnetic-ring/cmd"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/consense/dpoa"
	"github.com/sixexorg/magnetic-ring/http/router"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/p2pserver"
	netreqactor "github.com/sixexorg/magnetic-ring/p2pserver/actor/req"
	p2pactor "github.com/sixexorg/magnetic-ring/p2pserver/actor/server"
	storeactor "github.com/sixexorg/magnetic-ring/store/mainchain/actor"
	"github.com/sixexorg/magnetic-ring/store/mainchain/genesis"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	txpool "github.com/sixexorg/magnetic-ring/txpool/mainchain"

	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/radar/mainchain"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstorages"

	p2pcommon "github.com/sixexorg/magnetic-ring/p2pserver/common"
	"github.com/sixexorg/magnetic-ring/setup"
	"github.com/sixexorg/magnetic-ring/store/mainchain/validation"
	orgvalidation "github.com/sixexorg/magnetic-ring/store/orgchain/validation"
)

var (
	logger log.Logger
	gss    *genesis.Genesis
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Magnetic CLI"
	app.Action = setup.Start
	app.Version = config.Version
	app.Copyright = "Copyright in 2018 The Magnetic Authors"
	app.Commands = []cli.Command{
		cmd.AccountCommand,
		cmd.SignCommand,
	}
	app.Flags = []cli.Flag{cli.StringFlag{
		Name:  "wallet,w",
		Value: "./keystore/wallet",
		Usage: "Wallet `<file>`",
	},
		cli.StringFlag{
			Name: "password,p",
		},
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU() / 2)
		return nil
	}
	return app
}

func main() {
	initConfig()
	//os.Args = []string{"account","add"}
	if err := setupAPP().Run(os.Args); err != nil {
		fmt.Printf("run setup app err=%v\n", err)
		os.Exit(1)
	}

}

/*func testLeague(ctx *cli.Context) {
	//initConfig()                                //1.init config
	//initLog(config.GlobalConfig.SysCfg.LogPath) //2.init log
	//initTestNormalAccount(ctx)
	//initLedger()
	//_, p2ppid, _ := initP2P(ctx)            //5.init p2p
	//initConsensus(ctx, p2ppid, testAccount) //6.init consensus
	go runRouter() //7.init api
	if config.GlobalConfig.SysCfg.GenBlock {
		startLeague() //t.subchain
	} else {
		startFollowLeague()
	}
	waitFor() //wait for end
}*/

func initTestNormalAccount(ctx *cli.Context) account.Account {
	//log.Info("wallet info","filePath",ctx.String("wallet"))

	accountPath := ctx.String("wallet")

	mgr, err := account.NewManager()
	if err != nil {
		panic(err)
	}
	pasd := ctx.String("password")
	if pasd == "" {
		pasw, err := password2.GetPassword()
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
	log.Info("init account success", "account address", acct.Address().ToString())
	nodeimpl := acct.(account.Account)
	meacount.SetOwner(nodeimpl)
	return acct
}

//todo
func getAccountByAddress(address common.Address) (account.Account, error) {
	return mock.AccountNormal_1, nil
}

/*func startLeague() {

	dbdir := "test"
	defer os.RemoveAll(dbdir)
	ledger, err := storages.NewLedgerStore(dbdir)
	if err != nil {
		return
	}

	nodeId, _ := common.ToAddress("ct2qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	txCL := &types.Transaction{
		Version: 0x01,
		TxType:  types.CreateLeague,
		TxData: &types.TxData{
			From:    testAccount.Address(),
			Nonce:   1,
			Fee:     big.NewInt(20),
			Rate:    1000,
			MinBox:  50,
			MetaBox: big.NewInt(5000),
			NodeId:  nodeId,
			Private: false,
		},
	}

	block1 := &types.Block{
		Header: &types.Header{
			Height: 1,
		},
		Transactions: types.Transactions{txCL},
	}
	err = ledger.SaveBlockForMockTest(block1)
	if err != nil {
		return
	}

	TestContainer(testAccount)
}*/
/*func startFollowLeague() {

	dbdir := "test"
	defer os.RemoveAll(dbdir)
	ledger, err := storages.NewLedgerStore(dbdir)
	if err != nil {
		return
	}

	nodeId, _ := common.ToAddress("ct2qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	txCL := &types.Transaction{
		Version: 0x01,
		TxType:  types.CreateLeague,
		TxData: &types.TxData{
			From:    testAccount.Address(),
			Nonce:   1,
			Fee:     big.NewInt(20),
			Rate:    1000,
			MinBox:  50,
			MetaBox: big.NewInt(5000),
			NodeId:  nodeId,
			Private: false,
		},
	}

	block1 := &types.Block{
		Header: &types.Header{
			Height: 1,
		},
		Transactions: types.Transactions{txCL},
	}
	err = ledger.SaveBlockForMockTest(block1)
	if err != nil {
		return
	}
	leagueId, _ := common.ToAddress("ct2b5xc5uyaJrSqZ866gVvM5QtXybprb9sy")
	err = ledger.SaveBirthCertForMockTest(leagueId, txCL.Hash())
	if err != nil {
		return
	}

	TestFollowContainer(leagueId)
}*/
func CreateContainer(account account.Account, targetHeight uint64) error {
	TestContainer(account, targetHeight)
	return nil
}
func CreateOrdContainer(leagueId common.Address) error {
	TestFollowContainer(leagueId)
	return nil
}

func TestFollowContainer(leagueId common.Address) {

	cns := orgcontainer.GetContainers()
	id, err := cns.NewContainerByLeagueId(config.GlobalConfig.SysCfg.StoreDir, leagueId)
	if err != nil {
		log.Error("create container err", "error", err)
		return
	}
	for {
		status := cns.GetContainerStatus(id)
		log.Info("collect container status", "status", status)
		if status == orgcontainer.Normal {
			break
		}
		time.Sleep(time.Second)
	}
	container, err := cns.GetContainerById(id)
	if err != nil {
		log.Error("get container by id err", "error", err)
		return
	}
	log.Info("container create success", "leagueId", container.LeagueId().ToString())

	for _, v := range cns.ListLeagueIds() {
		log.Info("container leagueId list", "leagueId", v.ToString())
	}

	_, err = cns.GetContainer(container.LeagueId())
	if err != nil {
		log.Info("container check", "err", err)
	}
}

// 创建圈子的A节点
func TestContainer(account account.Account, targetHeight uint64) {
	node := meacount.GetOwner()
	cns := orgcontainer.GetContainers()
	id, err := cns.NewContainer(config.GlobalConfig.SysCfg.StoreDir, targetHeight, account, node)
	if err != nil {
		log.Error("create container err", "error", err)
		return
	}
	log.Info("create container success", "containerId", id)
	for {
		status := cns.GetContainerStatus(id)
		log.Info("collect container status", "status", status)
		if status == orgcontainer.Normal {
			break
		}
		time.Sleep(time.Second * 60)
	}

	container, err := cns.GetContainerById(id)

	if err != nil {
		log.Error("get container by id err", "error", err)
		return
	}
	log.Info("container success", "leagueId", container.LeagueId().ToString())

	pingpoa := poa.New(&config.GlobalConfig.CliqueCfg, account, container, container.Ledger())

	poaminer := miner.New(account.Address()) // miner global instance

	poaminer.RegisterOrg(container.LeagueId(), container.Ledger(), &config.GlobalConfig.CliqueCfg, pingpoa, container.TxPool()) //config是圈子共识的配置

	poaminer.StartOrg(container.LeagueId()) // start orgnization consensus

	//subtmr := time.NewTicker(time.Second * 10)
	//for {
	//	select {
	//	case <-subtmr.C:
	//		blkInfo := container.TxPool().Execute()
	//		//log.Info("execute success","block",blk,"accountstates",acctStates)
	//
	//		difficulty := big.NewInt(9)b
	//		blkInfo.Block.Header.SetDifficulty(difficulty)
	//
	//		err := container.Ledger().SaveAll(blkInfo)
	//		if err != nil {
	//			log.Error("save all error occured", "error", err)
	//			continue
	//		}
	//		log.Info("league save all success", "current height", blkInfo.Block.Header.Height)
	//
	//		removedHashes := make([]common.Hash, 0)
	//
	//		for _, tx := range blkInfo.Block.Transactions {
	//			removedHashes = append(removedHashes, tx.Hash())
	//		}
	//
	//		container.TxPool().RemovePendingTxs(removedHashes)
	//		container.TxPool().RefreshValidator(container.Ledger(), mock.Mock_Address_1)
	//
	//		log.Info("league block info", "block", blkInfo.Block, "current height", container.Ledger().GetCurrentHeaderHeight())
	//
	//		//container.Ledger().SetCurrentBlock(blockInfo.Block.Header.Height, blockInfo.Block.Hash())
	//
	//	}
	//}

	//container.TxPool().GenerateBlock()

	//ledger := container.Ledger()
	//currentBlock := ledger.GetCurrentBlockInfo()
	//
	//block2 := &orgtypes.Block{}
	//common.DeepCopy(&block2, mock.Block2)
	//block2.Header.PrevBlockHash = currentBlock.Hash()
	//for _, v := range block2.Transactions {
	//	log.Info("iterator txhash","txhash", v.Hash())
	//}
	//_, accoutStates, _, err := orgvalidation.ValidateBlock(block2, ledger)
	//if err != nil {
	//	log.Error("validate block err","error",err)
	//	return
	//}
	//err = ledger.SaveAll(mock.Block2, accoutStates)
	//if err != nil {
	//	log.Error("save all error when ")
	//	return
	//}
}

func startMagnetic(ctx *cli.Context) {
	//initConfig()
	p2pcommon.InitBootNode(config.GlobalConfig.BootNode.IP)
	initLog(config.GlobalConfig.SysCfg.LogPath)
	acct := initTestNormalAccount(ctx)
	nodeacct := acct.(account.NormalAccount)
	mainRadar := initLedgerAndRadar(nodeacct)
	initGenesis()
	initPool(mainRadar)
	_, p2pid, _ := initP2P(ctx)
	//log.Info("p2pid","PID",p2pid)
	initConsensus(ctx, p2pid, acct)
	initContainers()
	go runRouter()

	//if config.GlobalConfig.SysCfg.GenBlock {
	//	go run()
	//}
	go func() {
		if config.GlobalConfig.SysCfg.GenBlock {
			fmt.Println("🦋 create league start")
			CreateContainer(mock.AccountNormal_1, 1)
		} else {
			/*leagueId, _ := common.ToAddress("ct2e3b22e6gBHUMV2DFF7A6YRY1g3nr3fbv")
			CreateOrdContainer(leagueId)*/
		}
	}()
	waitFor()

}

func initConfig() {
	p, err := os.Getwd()
	if err != nil {
		return
	}

	err = config.InitConfig(path.Join(p, "./config.yml"))

	if err != nil {
		fmt.Printf("init config error-->%v\n", err)
		os.Exit(0)
		return
	}
	//fmt.Println("init config success")
}

func initLedgerAndRadar(nodeacct account.NormalAccount) *mainchain.LeagueConsumers {
	props := actor.FromProducer(func() actor.Actor { return &mainchain.MainRadarActor{} })
	containerId2P2P := actor.Spawn(props)
	bactor.RegistActorPid(bactor.MAINRADARACTOR, containerId2P2P)

	leadge, err := storages.NewLedgerStore(config.GlobalConfig.SysCfg.StoreDir)
	if err != nil {
		panic(fmt.Errorf("init ledger error ", err))
	}
	log.Info("init ledger success")
	extLedger, err := extstorages.NewLedgerStore(config.GlobalConfig.SysCfg.StoreDir)
	if err != nil {
		panic(fmt.Errorf("init extLedger error ", err))
	}

	validation.NewAccountNonceCache(func(height uint64, account common.Address) uint64 {
		ass, err := leadge.GetAccountByHeight(height, account)
		if err != nil {
			return 0
		}
		return ass.Data.Nonce
	})

	adpter := &mainchain.ConsumerAdapter{
		FuncGenesis:  leadge.GetTxByLeagueId,
		FuncTx:       leadge.GetTxByHash,
		FuncBlk:      leadge.GetBlockHeaderByHeight,
		Ledger:       extLedger,
		FuncValidate: orgvalidation.ValidateBlock,
		NodeId:       nodeacct.PublicKey().Hex(),
	}
	mainchain.NewNodeLHCache()
	mainRadar := mainchain.NewLeagueConsumers(adpter)
	return mainRadar
}

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

func waitFor() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-c
	log.Info("server shut down")
}

func initPool(mainRadar *mainchain.LeagueConsumers) error {
	pool, err := txpool.InitPool()
	pool.SetMainRadar(mainRadar)
	return err
}

func runRouter() {

	router.StartRouter()

	engin := router.GetRouter()

	//pool, err := txpool.GetPool()
	//if err != nil {
	//	log.Info("get txpool error when run router", "error", err)
	//	os.Exit(2)
	//}
	pooactor, err := bactor.GetActorPid(bactor.TXPOOLACTOR)
	if err != nil {
		log.Error("get actor pid err", "error", err)
		os.Exit(0)
	}
	httpactor.SetTxPoolPid(pooactor)

	addr := fmt.Sprintf(":%d", config.GlobalConfig.SysCfg.HttpPort)

	engin.Run(addr)
}
func initGenesis() {

	genesis, err := genesis.InitGenesis(&config.GlobalConfig.Genesis)
	if err != nil {
		os.Exit(2)
	}
	gss = genesis
	ledger := storages.GetLedgerStore()

	if ledger.GetCurrentBlockHeight() == 0 {
		log.Info("no block found,init genesis")

		/*genesis, err := genesis.InitGenesis(&config.GlobalConfig.Genesis)
		if err != nil {
			logger.Error("main", "init genesis err", err)
			return
		}
		log.Info("main", "genesis", genesis)*/
		err := gss.GenesisBlock(ledger)
		if err != nil {
			logger.Error("main", "init genesis err", err)
			return
		}

		if err != nil {
			logger.Error("main", "BlockBatchCommit001", err)
			return
		}

		mainpool, err := txpool.GetPool()
		if err != nil {
			return
		}

		mainpool.RefreshValidator()

	}
}
func initP2P(ctx *cli.Context) (*p2pserver.P2PServer, *actor.PID, error) {
	p2p := p2pserver.NewServer()

	p2pActor := p2pactor.NewP2PActor(p2p)
	p2pPID, err := p2pActor.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("p2pActor init error %s", err)
	}
	p2p.SetPID(p2pPID)
	bactor.RegistActorPid(bactor.P2PACTOR, p2pPID)
	err = p2p.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("p2p service start error %s", err)
	}

	pool, err := txpool.GetPool()
	if err != nil {
		log.Info("get txpool error when run router", "error", err)
		os.Exit(2)
	}
	//
	netreqactor.SetTxnPoolPid(pool.GetTxpoolPID())
	// txpoolSvr.RegisterActor(tc.NetActor, p2pPID)
	// hserver.SetNetServerPID(p2pPID)
	p2p.WaitForPeersStart()
	// log.Infof("P2P init success")
	return p2p, p2pPID, nil
}
func initContainers() {
	pid, err := bactor.GetActorPid(bactor.P2PACTOR)
	if err != nil {
		panic(err)
	}
	orgcontainer.InitContainers(pid)
}
func initConsensus(ctx *cli.Context, p2pPid *actor.PID, acc account.Account) error {
	nmlAcct := acc.(account.NormalAccount)
	consensusService, err := dpoa.NewServer(nmlAcct, p2pPid)
	if err != nil {
		return err
	}
	consensusService.Start()

	netreqactor.SetConsensusPid(consensusService.GetPID())
	//hserver.SetConsensusPid(consensusService.GetPID())
	storeactor.SetConsensusPid(consensusService.GetPID())

	log.Info("Consensus init success")
	//return consensusService, nil
	return nil
}

func run() {
	genBlock()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			genBlock()
		}
	}
}

func genBlock() {
	curHeit := storages.GetLedgerStore().GetCurrentBlockHeight()
	//emptyHash := common.Hash{}

	ledgerStore := storages.GetLedgerStore()
	if ledgerStore == nil {
		return
	}

	if curHeit == 0 {
		log.Info("no block found,init genesis")

		/*genesis, err := genesis.InitGenesis(&config.GlobalConfig.Genesis)
		if err != nil {
			logger.Error("main", "init genesis err", err)
			return
		}
		log.Info("main", "genesis", genesis)*/
		err := gss.GenesisBlock(ledgerStore)
		if err != nil {
			logger.Error("main", "init genesis err", err)
			return
		}

		if err != nil {
			logger.Error("main", "BlockBatchCommit001", err)
			return
		}

		mainpool, err := txpool.GetPool()
		if err != nil {
			return
		}

		mainpool.RefreshValidator()

		//ledgerStore.SetCurrentBlock(blk.Header.Height, blk.Hash())
		//
		//log.Info("block info", "blk", blk)

	} else {
		pool, err := txpool.GetPool()

		//blk := pool.GenerateBlock(curHeit, true)

		blockInfo := pool.Execute()

		err = ledgerStore.SaveAll(blockInfo)
		if err != nil {
			logger.Error("main", "save genesisbubu err", err)
			return
		}

		removedHashes := make([]common.Hash, 0)

		router.TempPutTransactions(blockInfo.Block.Transactions)

		for _, tx := range blockInfo.Block.Transactions {
			removedHashes = append(removedHashes, tx.Hash())
		}

		pool.RemovePendingTxs(removedHashes)
		pool.RefreshValidator()

		log.Info("block info", "blk", blockInfo.Block, "current height", storages.GetLedgerStore().GetCurrentBlockHeight(), "headerHeigh", ledgerStore.GetCurrentHeaderHeight())

		//ledgerStore.SetCurrentBlock(blockInfo.Block.Header.Height, blockInfo.Block.Hash())

		time.Sleep(time.Second)

		tbk, err := ledgerStore.GetBlockByHeight(curHeit)
		if err != nil {
			logger.Error("GetBlockByHeight err", "error", err)
			return
		}
		log.Info("tbk info", "heigt", curHeit, "tbk", tbk)

	}
}
