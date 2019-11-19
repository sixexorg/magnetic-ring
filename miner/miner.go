package miner

import (
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/consense/poa"
	"github.com/sixexorg/magnetic-ring/events/message"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	common2 "github.com/sixexorg/magnetic-ring/txpool/orgchain"
)

type Miner struct {
	coinbase    common.Address
	mining      map[common.Address]int32
	canStart    map[common.Address]int32   // can start indicates whether we can start the mining operation
	shouldStart map[common.Address]int32   // should start indicates whether we should start after sync
	workers     map[common.Address]*worker // key orgid
	ledgerSet   map[common.Address]*storages.LedgerStoreImp
}

func New(coinbase common.Address) *Miner {
	miner := &Miner{
		ledgerSet:   make(map[common.Address]*storages.LedgerStoreImp),
		workers:     make(map[common.Address]*worker, 0),
		mining:      make(map[common.Address]int32, 0),
		canStart:    make(map[common.Address]int32, 0),
		shouldStart: make(map[common.Address]int32, 0),
		coinbase:    coinbase,
	}
	return miner
}

func (srv *Miner) Halt(c actor.Context) {
}

func (srv *Miner) Receive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *actor.Restarting:
		log.Info("Miner actor restarting")
	case *actor.Stopping:
		log.Info("Miner actor stopping")
	case *actor.Stopped:
		log.Info("Miner actor stopped")
	case *actor.Started:
		log.Info("Miner actor started")
	case *actor.Restart:
		log.Info("Miner actor restart")
	case *message.SaveOrgBlockCompleteMsg:
		//if wk, ok := srv.workers[msg.Block.Header.LeagueId]; ok {
		//	wk.commitNewWork()
		//}

	default:
		log.Info("vbft actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *Miner) Start() {
}

func (self *Miner) Stop() {
	for orgid, worker := range self.workers {
		worker.stop()
		self.mining[orgid] = 0
		self.shouldStart[orgid] = 0
	}
}

func (self *Miner) StartOrg(orgid common.Address) {
	if _, ok := self.ledgerSet[orgid]; !ok {
		return
	}
	self.workers[orgid].setEtherbase(self.coinbase)
	self.shouldStart[orgid] = 1
	if self.canStart[orgid] == 0 {
		log.Info("Network syncing, will start miner afterwards")
		return
	}
	self.mining[orgid] = 1

	log.Info("Starting mining operation", orgid)
	self.workers[orgid].start()
	self.workers[orgid].commitNewWork()
}

func (self *Miner) StopOrg(orgid common.Address) {
	self.workers[orgid].stop()
	delete(self.workers, orgid)
	self.mining[orgid] = 0
	self.shouldStart[orgid] = 0
}

func (self *Miner) RegisterOrg(orgid common.Address, ledger *storages.LedgerStoreImp, config *config.CliqueConfig, poains *poa.Clique, txpool *common2.SubPool) {
	if _, ok := self.ledgerSet[orgid]; ok {
		return
	}
	self.ledgerSet[orgid] = ledger
	self.workers[orgid] = newWorker(config, self.coinbase, ledger, orgid, poains, txpool)
	self.workers[orgid].register(NewCpuAgent(poains))

	self.mining[orgid] = 0
	self.shouldStart[orgid] = 0
	self.canStart[orgid] = 1
}

func (self *Miner) UnegisterOrg(orgid common.Address) {
	if work, ok := self.workers[orgid]; ok {
		work.unregister()
		delete(self.workers, orgid)
	}

	delete(self.ledgerSet, orgid)
	delete(self.mining, orgid)
	delete(self.canStart, orgid)
}

func (self *Miner) Mining(orgid common.Address) bool {
	return self.mining[orgid] > 0
}

func (self *Miner) SetEtherbase(addr common.Address) {
	self.coinbase = addr
	for _, worker := range self.workers {
		worker.setEtherbase(addr)
	}
}
