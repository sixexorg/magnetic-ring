package mainchain

import (
	"math/big"

	"sync"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/extstorages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

var tellerNil = &actor.PID{}

type MacroConsumer interface {
	////////////public///////////////
	//ReceiverBlocks is puts the blocks to be validated into the validation pool ahead of time
	ReceiveBlocks(blocks []*orgtypes.Block)
	//Truncate is to notifies the program that this round of calculations will be closed,Stop gracefully
	Truncate()
	//Identify  is to actively launch validation，run new round
	Identify()

	////////////private///////////////
	await()
	newRound()
	routing()
	setStopSignal(leagueId common.Address, tx *maintypes.Transaction)
	getLeagueConsume(leagueId common.Address) *LeagueConsumer
}

//run automatically
type MicroConsumer interface {
	registerConsume(adapter *ConsumerAdapter, setStopSignal func(leagueId common.Address, tx *maintypes.Transaction), getLightHouse func() bool)
	receiveBlock(block *orgtypes.Block)
	verifyBlock(block *orgtypes.Block, adpter *ConsumerAdapter) (storelaw.AccountStaters, *big.Int, error)
	consume(adapter *ConsumerAdapter)
	saveAll(block *orgtypes.Block, states storelaw.AccountStaters, fee *big.Int, adpter *ConsumerAdapter) error
	createMainChainTx() (*maintypes.Transaction, error)
	checkGenesisBlock(block *orgtypes.Block)
	clearHistorical()
	checkReferencedTransaction(tx *orgtypes.Transaction, adapter *ConsumerAdapter) error
	authenticate(block *orgtypes.Block) bool
}

type ConsumerAdapter struct {
	FuncGenesis   func(leagueId common.Address) (*maintypes.Transaction, uint64, error)
	FuncTx        func(txHash common.Hash) (*maintypes.Transaction, uint64, error)
	FuncBlk       func(height uint64) (*maintypes.Header, error)
	Ledger        *extstorages.LedgerStoreImp
	FuncValidate  func(block *orgtypes.Block, ledgerStore storelaw.Ledger4Validation) (blkInfo *storelaw.OrgBlockInfo, err error)
	P2pTlr        bactor.Teller
	TxPoolTlr     bactor.Teller
	P2pTlrInit    bactor.InitTellerfunc
	TxPoolTlrInit bactor.InitTellerfunc
	NodeId        string
	TlrReady      bool
	m             sync.Mutex
}

func (this *ConsumerAdapter) initTeller() {
	this.m.Lock()
	this.m.Unlock()
	if !this.TlrReady {
		var err error
		this.P2pTlr, err = this.P2pTlrInit()
		if err != nil {
			panic(err)
		}
		this.TxPoolTlr, err = this.TxPoolTlrInit()
		if err != nil {
			panic(err)
		}
		this.TlrReady = true
	}
}

type LeagueStatePipe struct {
	StateSignal chan interface{}
	Successed   chan common.Address
}

func NewLeagueStatePipe() *LeagueStatePipe {
	lsp := &LeagueStatePipe{
		StateSignal: make(chan interface{}, 50),
		Successed:   make(chan common.Address, 5),
	}
	return lsp
}
func (this *LeagueStatePipe) Kill() {
	close(this.StateSignal)
	close(this.Successed)
}
func (this *LeagueStatePipe) sendOk(leagueId common.Address) {
	this.Successed <- leagueId
}
func (this *LeagueStatePipe) sendError(leagueId common.Address, err error) {
	this.StateSignal <- &LeagueErr{LeagueId: leagueId, Err: err}
}
func (this *LeagueStatePipe) sendLeagueNeed(leagueId common.Address, height uint64) {
	this.StateSignal <- &LeagueNeed{LeagueId: leagueId, BlockHeight: height}
}
func (this *LeagueStatePipe) sendLeagueExec(leagueId common.Address, height uint64) {
	this.StateSignal <- &LeagueExec{LeagueId: leagueId, Confirmed: height}
}

type LeagueNeed struct {
	LeagueId    common.Address
	BlockHeight uint64
}
type LeagueExec struct {
	LeagueId  common.Address
	Confirmed uint64
}
type LeagueErr struct {
	LeagueId common.Address
	Err      error
}

type MainTxCradle struct {
	LeagueId     common.Address
	Check        *maintypes.Transaction
	ReferenceTxs maintypes.Transactions
}
