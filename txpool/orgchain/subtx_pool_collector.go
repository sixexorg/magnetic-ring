package orgchain

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/bactor"
	orgcommon "github.com/sixexorg/magnetic-ring/p2pserver/common"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/log"

	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
	"github.com/sixexorg/magnetic-ring/store/orgchain/validation"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

type SubPool struct {
	pdmgr *PendingMgr //Maintain a collection of transactions that can be packaged
	//queue *SubTxQueue //Maintain a collection of transactions transmitted over the network
	txChan       chan *types.Transaction
	maxPending   uint32
	maxInPending uint32
	maxInQueue   uint32
	leagueId     common.Address

	waitTxNum map[common.Hash]uint8
	waitPool  types.Transactions

	stateValidator *validation.StateValidate

	txpoolPid   *actor.PID
	mustPackTxs []*types.Transaction
}

func (pool *SubPool) Start() {
	go func() {
		for {
			select {
			case tx := <-pool.txChan:
				fmt.Println("♦️ subTxpool add")
				err := pool.AddTx(pool.leagueId, tx)
				if err != nil {
					log.Error("addtx error", "tx_pool", err)
				}
			}
		}
	}()
}

func startActor(obj interface{}, id string) (*actor.PID, error) {
	props := actor.FromProducer(func() actor.Actor {
		return obj.(actor.Actor)
	})

	pid, _ := actor.SpawnNamed(props, id)
	if pid == nil {
		return nil, fmt.Errorf("fail to start actor at props:%v id:%s", props, id)
	}
	return pid, nil
}

func NewSubPool(ledger *storages.LedgerStoreImp, leadgeId common.Address) *SubPool {
	pool := new(SubPool)
	pool.pdmgr = NewPendingMgr()
	//pool.queue = NewSubTxQueue()
	pool.maxPending = config.GlobalConfig.TxPoolCfg.MaxPending
	pool.maxInPending = config.GlobalConfig.TxPoolCfg.MaxInPending
	pool.maxInQueue = config.GlobalConfig.TxPoolCfg.MaxInQueue
	pool.txChan = make(chan *types.Transaction, pool.maxInQueue)
	pool.leagueId = leadgeId

	pool.waitPool = make(types.Transactions, 0, 10000)
	pool.waitTxNum = make(map[common.Hash]uint8, 10000)

	pool.RefreshValidator(ledger, leadgeId)
	pool.mustPackTxs = make([]*types.Transaction, 0)
	poolActor := NewTxActor(pool)

	pid, err := startActor(poolActor, "subtxpoolAcotor")
	if err != nil {
		return nil
	}

	pool.txpoolPid = pid

	pool.Start()
	return pool
}

func (pool *SubPool) AppendMustPackTx(txs ...*types.Transaction) {

	for _, tx := range txs {
		pool.mustPackTxs = append(pool.mustPackTxs, tx)
	}

}

func (pool *SubPool) refreshWaitPool() {
	if pool.waitPool.Len() > 0 {
		for wp := range pool.waitPool {
			pool.TxEnqueue(pool.waitPool[wp])
		}
	}
	pool.waitPool = make(types.Transactions, 0, 10000)
	pool.waitTxNum = make(map[common.Hash]uint8, 10000)
}

func (pool *SubPool) RefreshValidator(ledger *storages.LedgerStoreImp, leadgeId common.Address) {
	if pool.stateValidator != nil {
		fmt.Println("🚰 refreshValidator txlen", pool.stateValidator.GetTxLen())
	}
	oldSV := pool.stateValidator
	pool.stateValidator = validation.NewStateValidate(ledger, leadgeId)
	if oldSV != nil {
		txch := oldSV.TxInheritance()
		for ch := range txch {
			pool.TxEnqueue(ch)
		}
	}
	pool.refreshWaitPool()
	//log.Info("func txpool RefreshValidator 02", "newTargetHeight", pool.stateValidator.TargetHeight, "txlen", pool.stateValidator.GetTxLen())
	fmt.Println("func txpool RefreshValidator 02 ", "newTargetHeight=", pool.stateValidator.TargetHeight, " txlen=", pool.stateValidator.GetTxLen())
}

func (pool *SubPool) AddTx(leagueId common.Address, tx *types.Transaction) error {
	fmt.Println("⭕️ txpool addtx 1 type:", tx.TxType)
	result, err := pool.stateValidator.VerifyTx(tx)
	txHash := tx.Hash()
	fmt.Println("⭕️ txpool addtx 2", result, err)
	if result == 1 || result == 0 {
		orgTx := orgcommon.OrgTx{
			OrgId: pool.leagueId,
			Tx:    tx,
		}
		P2PPid, err := bactor.GetActorPid(bactor.P2PACTOR)
		if err != nil {
			log.Error("tx_pool.go get p2ppid error", "error", err)
		} else {
			P2PPid.Tell(orgTx)
		}
	}
	fmt.Println("⭕️ txpool addtx 3", result, err)
	switch result {
	case -1:
		return err
	case 0:
		if uint32(len(pool.txChan)) < pool.maxInQueue {

			pool.waitTxNum[txHash]++
			if uint32(pool.waitTxNum[txHash]) == config.GlobalConfig.TxPoolCfg.MaxTxInPool {
				pool.waitPool = append(pool.waitPool, tx)
			} else if uint32(pool.waitTxNum[txHash]) < config.GlobalConfig.TxPoolCfg.MaxTxInPool {
				pool.TxEnqueue(tx)
			}

		} else {
			return errors.ERR_TXPOOL_OUTOFMAX
		}
		return nil
	case 1:
		pool.pdmgr.addTx(tx, pool.maxPending)
		return nil
	}

	return nil
}

/**
generate block
*/
func (pool *SubPool) GenerateBlock(leagueId common.Address, height uint64, packtx bool) *types.Block {

	sts := states.AccountStates{}
	var txs *types.Transactions
	if packtx {
		txs = pool.pdmgr.getTxs(pool.maxPending)
		sort.Sort(txs)
	}

	block := &types.Block{
		Header: &types.Header{
			Height:        pool.stateValidator.TargetHeight,
			Version:       types.TxVersion,
			PrevBlockHash: common.Hash{},
			ReceiptsRoot:  common.Hash{},
			StateRoot:     sts.GetHashRoot(),
			TxRoot:        txs.GetHashRoot(),
			Timestamp:     uint64(time.Now().Unix()),
		},
		Transactions: *txs,
	}

	//reset validator
	//pool.RefreshValidator()
	return block
}

/**
func (pool *SubPool) ValidateSyncTxs(leagueId common.Address, txhashes []*common.Hash) error {

	if pool.pdmgr.pendingTxs == nil || len(pool.pdmgr.pendingTxs) < 1 {
		return errors.ERR_TXPOOL_TXNOTFOUND
	}

	for _, v := range txhashes {
		vtx := pool.queue.Remove(*v)

		if vtx == nil {
			return errors.ERR_TXPOOL_TXNOTFOUND
		}

		result, err := pool.stateValidator.VerifyTx(vtx)
		switch result {
		case -1:
			return err
		case 0:
			if uint32(pool.queue.Size()) < pool.maxInQueue {
				pool.queue.Enqueue(vtx)
			} else {
				return errors.ERR_TXPOOL_OUTOFMAX
			}
			return nil
		case 1:
			pool.pdmgr.addTx(vtx, pool.maxPending)
			return nil
		}
	}
	return nil
}
**/
func (pool *SubPool) Execute() (blkInfo *storelaw.OrgBlockInfo) {
	//fmt.Println("⭕️ subtxpool execute txlen", pool.stateValidator.GetTxLen())

	for _, tx := range pool.mustPackTxs {
		ret, err := pool.stateValidator.VerifyTx(tx)
		if err != nil {
			log.Error("range pool.mustPackTxs verifyTx error", "ret", ret, "error", err)
		}
	}

	return pool.stateValidator.ExecuteOplogs()
}

type PendingMgr struct {
	sync.RWMutex
	pendingTxs map[common.Hash]*types.Transaction
}

func NewPendingMgr() *PendingMgr {
	mgr := new(PendingMgr)
	mgr.pendingTxs = make(map[common.Hash]*types.Transaction)
	return mgr
}

/*
	get txs from pending transaction list
*/
func (pm *PendingMgr) getTxs(maxInblock uint32) *types.Transactions {

	pm.Lock()
	defer pm.Unlock()
	if len(pm.pendingTxs) < 1 {
		return nil
	}

	txs := make(types.Transactions, 0)

	for _, v := range pm.pendingTxs {
		txs = append(txs, v)
	}

	sort.Sort(txs)

	le := uint32(len(txs))
	if le > maxInblock {
		le = maxInblock
	}

	ret := txs[0:le]

	return &ret
}

func (pool *SubPool) RemovePendingTxs(hashes []common.Hash) {
	pool.pdmgr.removePendingTxs(hashes)
}

/*
	get txs from pending transaction list
*/
func (pm *PendingMgr) removePendingTxs(txhashes []common.Hash) {

	pm.Lock()
	defer pm.Unlock()

	if pm.pendingTxs == nil || len(pm.pendingTxs) < 1 {
		return
	}

	for _, v := range txhashes {
		delete(pm.pendingTxs, v)
	}

}

/*
	add tx to pending transaction list
*/
func (pm *PendingMgr) addTx(tx *types.Transaction, maxPending uint32) error {

	pm.Lock()
	defer pm.Unlock()

	if uint32(len(pm.pendingTxs)) > maxPending {
		return errors.ERR_TXPOOL_OUTOFMAX
	}

	pm.pendingTxs[tx.Hash()] = tx
	return nil
}

func (pool *SubPool) TxEnqueue(tx *types.Transaction) error {

	if uint32(len(pool.txChan)) >= pool.maxInQueue {
		return errors.ERR_TXPOOL_OUTOFMAX
	}
	fmt.Println("🌈 🌈  insert tx into subTxpool")
	pool.txChan <- tx

	return nil
}

func (pool *SubPool) GetTxpoolPID() *actor.PID {
	return pool.txpoolPid
}
