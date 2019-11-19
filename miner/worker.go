package miner

import (
	"sync"
	"sync/atomic"
	"time"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/consense/poa"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
	common2 "github.com/sixexorg/magnetic-ring/txpool/orgchain"
)

const (
	resultQueueSize  = 10
	miningLogAtDepth = 5

	// txChanSize is the size of channel listening to TxPreEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096
	// chainHeadChanSize is the size of channel listening to ChainHeadEvent.
	chainHeadChanSize = 10
	// chainSideChanSize is the size of channel listening to ChainSideEvent.
	chainSideChanSize = 10
)

// Agent can register themself with the worker
type Agent interface {
	Work() chan<- *Work
	SetReturnCh(chan<- *Result)
	Stop()
	Start()
	//GetHashRate() int64
}

// Work is the workers current environment and holds
// all of the current state information
type Work struct {
	config    *config.CliqueConfig
	orgname   common.Hash
	tcount    int          // tx count in cycle
	Block     *types.Block // the new block
	header    *types.Header
	txs       []*types.Transaction
	receipts  []*types.Receipt
	storedata *storelaw.OrgBlockInfo

	createdAt time.Time
}

type Result struct {
	Work  *Work
	Block *types.Block
}

type worker struct {
	config    *config.CliqueConfig
	agent     Agent
	recv      chan *Result
	coinbase  common.Address
	extra     []byte
	ledger    *storages.LedgerStoreImp
	currentMu sync.Mutex
	current   *Work
	poaIns    *poa.Clique
	txpool    *common2.SubPool
	mining    int32
	atWork    int32
	orgname   common.Address

	mu sync.Mutex
	wg sync.WaitGroup
}

func newWorker(config *config.CliqueConfig, coinbase common.Address, ledger *storages.LedgerStoreImp, orgname common.Address, poains *poa.Clique, txpool *common2.SubPool) *worker {
	worker := &worker{
		config:   config,
		recv:     make(chan *Result, resultQueueSize),
		ledger:   ledger,
		coinbase: coinbase,
		orgname:  orgname,
		poaIns:   poains,
		txpool:   txpool,
	}

	//go worker.update()
	go worker.wait()
	//worker.commitNewWork()

	return worker
}

func (self *worker) setEtherbase(addr common.Address) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.coinbase = addr
}

func (self *worker) setExtra(extra []byte) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.extra = extra
}

func (self *worker) start() {
	self.mu.Lock()
	defer self.mu.Unlock()

	atomic.StoreInt32(&self.mining, 1)

	self.agent.Start()
}

func (self *worker) stop() {
	self.wg.Wait()

	self.mu.Lock()
	defer self.mu.Unlock()
	if atomic.LoadInt32(&self.mining) == 1 {
		self.agent.Stop()
	}
	atomic.StoreInt32(&self.mining, 0)
	atomic.StoreInt32(&self.atWork, 0)
}

func (self *worker) register(agent Agent) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agent = agent
	agent.SetReturnCh(self.recv)
}

func (self *worker) unregister() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.agent.Stop()
}

func (self *worker) wait() {
	for {
		for result := range self.recv {
			atomic.AddInt32(&self.atWork, -1)
			if result == nil {
				time.Sleep(time.Second * 5)
				self.commitNewWork()
				continue
			}
			fmt.Println("~~~~~~~~~~øøøøøøøøøsave block", time.Unix(0, int64(result.Work.storedata.Block.Header.Timestamp)).Format("15:04:05.000"),
				result.Work.storedata.Block.Header.Height, result.Work.storedata.Block.Header.Coinbase.ToString())
			result.Work.storedata.Block = result.Block
			err := self.ledger.SaveAll(result.Work.storedata)
			fmt.Println("øøøøøøøøøsave block", result.Work.storedata.Block.Header.Height, result.Work.storedata.Block.Header.Coinbase.ToString(), err)
			self.txpool.RefreshValidator(self.ledger, self.orgname)

			if result.Work.storedata.Block.Header.Height == self.ledger.GetCurrentBlockHeight() {
				time.Sleep(time.Millisecond * 50)
			}
			self.commitNewWork()
		}
	}
}

// push sends a new work task to currently live miner agents.
func (self *worker) push(work *Work) {
	if atomic.LoadInt32(&self.mining) != 1 {
		return
	}
	atomic.AddInt32(&self.atWork, 1)
	if ch := self.agent.Work(); ch != nil {
		ch <- work
	} else {
		log.Info("worker push no work", "work", work)
	}
}

// makeCurrent creates a new environment for the current cycle.
func (self *worker) makeCurrent(parent *types.Block, header *types.Header) error {
	work := &Work{
		config:    self.config,
		header:    header,
		createdAt: time.Now(),
	}

	work.tcount = 0
	self.current = work
	return nil
}

func (self *worker) commitNewWork() {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.currentMu.Lock()
	defer self.currentMu.Unlock()

	tstart := time.Now()
	h, _ := self.ledger.GetCurrentBlock()
	parent, _ := self.ledger.GetBlockByHeight(h)
	//tstamp := uint64(tstart.Unix())
	//if parent.Header.Timestamp >= tstamp {
	//	tstamp = parent.Header.Timestamp + 1
	//}
	// this will ensure we're not going off too far in the future
	//if now := uint64(time.Now().Unix()); tstamp > now+1 {
	//	wait := time.Duration(tstamp-now) * 100 * time.Millisecond//wait := time.Duration(tstamp-now) * time.Second
	//	log.Info("Mining too far in the future", "wait", wait)
	//	time.Sleep(wait)
	//}
	//self.txpool.RefreshValidator(self.ledger, self.orgname)
	blkInfo := self.txpool.Execute()
	blkInfo.Block.Header.Coinbase = self.coinbase
	blkInfo.Block.Header.Timestamp = parent.Header.Timestamp
	blkInfo.Block.Header.Timestamp = uint64(time.Unix(0, int64(blkInfo.Block.Header.Timestamp)).Add(time.Millisecond * 500).UnixNano())
	if time.Now().After(time.Unix(0, int64(blkInfo.Block.Header.Timestamp))) {
		blkInfo.Block.Header.Timestamp = uint64(time.Now().UnixNano())
	}
	if err := self.poaIns.Prepare(blkInfo.Block.Header); err != nil {
		log.Error("Failed to prepare header for mining", "err", err)
		return
	}

	err := self.makeCurrent(parent, blkInfo.Block.Header)
	if err != nil {
		log.Error("Failed to create mining context", "err", err)
		return
	}

	work := self.current
	work.storedata = blkInfo
	if work.Block, err = self.poaIns.Finalize(blkInfo); err != nil {
		log.Error("Failed to finalize block for sealing", "err", err)
		return
	}
	//self.txpool.RefreshValidator(self.ledger, self.orgname)
	// We only care about logging if we're actually mining.
	if atomic.LoadInt32(&self.mining) == 1 {
		log.Info("Commit new mining work", "number", work.Block.Header.Height, "txs", work.tcount, "uncles", 0, "elapsed", time.Since(tstart))
	}
	self.push(work)
}
