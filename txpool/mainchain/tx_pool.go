package mainchain

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/config"

	"github.com/sixexorg/magnetic-ring/log"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/radar/mainchain"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/mainchain/validation"
)

var (
	mainTxPool *TxPool
)

type TxPool struct {
	pdmgr *PendingMgr
	//queue     *TxQueue
	waitTxNum map[common.Hash]uint8
	waitPool  types.Transactions
	txChan    chan *types.Transaction

	maxPending   uint32
	maxInPending uint32
	maxInQueue   uint32

	stateValidator *validation.StateValidate

	ticker *time.Ticker

	txpoolPid *actor.PID

	mustPackTxs []*types.Transaction
	mainRadar   *mainchain.LeagueConsumers
}

func NewTxPool() *TxPool {
	pool := new(TxPool)
	pool.pdmgr = NewPendingMgr()
	//pool.queue = NewTxQueue()
	pool.maxPending = config.GlobalConfig.TxPoolCfg.MaxPending
	pool.maxInPending = config.GlobalConfig.TxPoolCfg.MaxInPending
	pool.maxInQueue = config.GlobalConfig.TxPoolCfg.MaxInQueue
	pool.txChan = make(chan *types.Transaction, pool.maxInQueue)

	pool.RefreshValidator()
	pool.waitPool = make(types.Transactions, 0, 10000)
	pool.waitTxNum = make(map[common.Hash]uint8, 10000)
	pool.mustPackTxs = make([]*types.Transaction, 0)
	return pool
}
func (pool *TxPool) SetMainRadar(mainRadar *mainchain.LeagueConsumers) {
	pool.mainRadar = mainRadar
}

func (pool *TxPool) AppendMustPackTx(txs ...*types.Transaction) {

	for _, tx := range txs {
		pool.mustPackTxs = append(pool.mustPackTxs, tx)
	}

}

type PendingMgr struct {
	sync.RWMutex
	pendingTxs map[common.Hash]*types.Transaction
}

func InitPool() (*TxPool, error) {
	var err error
	mainTxPool, err = InitTxPool()

	if err != nil {
		return nil, err
	}

	mainTxPool.Start()
	return mainTxPool, nil
}

func GetPool() (*TxPool, error) {
	if mainTxPool == nil {
		return nil, errors.ERR_TXPOOL_UNINIT
	}
	return mainTxPool, nil
}

func (pool *TxPool) Start() {
	pool.RefreshValidator()
	go func() {
		for {
			select {
			case tx := <-pool.txChan:

				//if !pool.queue.IsEmpty() {
				//tx = pool.queue.Dequeue()

				err := pool.AddTx(tx)
				if err != nil {
					fmt.Printf("addtx and validate error=%v\n", err)
					log.Info("addtx error", "error", err, "targetBlockHeight", pool.stateValidator.TargetHeight, "errors.ERR_TXPOOL_OUTOFMAX", pool.maxInQueue, "len(pool.txChan)", len(pool.txChan))
				}
				//}
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

func InitTxPool() (*TxPool, error) {
	pool := NewTxPool()

	poolActor := NewTxActor(pool)

	pid, err := startActor(poolActor, "txpoolAcotor")
	if err != nil {
		return nil, err
	}

	bactor.RegistActorPid(bactor.TXPOOLACTOR, pid)
	pool.txpoolPid = pid

	return pool, nil

}

func (pool *TxPool) refreshWaitPool() {
	if pool.waitPool.Len() > 0 {
		for k, _ := range pool.waitPool {
			pool.TxEnqueue(pool.waitPool[k])
		}
	}
	pool.waitPool = make(types.Transactions, 0, 10000)
	pool.waitTxNum = make(map[common.Hash]uint8, 10000)
}
func (pool *TxPool) RefreshValidator() {
	if pool != nil && pool.stateValidator != nil {
		//log.Info("func txpool RefreshValidator 01", "oldTargetHeight", pool.stateValidator.TargetHeight, "txlen", pool.stateValidator.GetTxLen())
		fmt.Println("func txpool RefreshValidator 01 ", "oldTargetHeight=", pool.stateValidator.TargetHeight, " txlen=", pool.stateValidator.GetTxLen())
	}
	ledgerStore := storages.GetLedgerStore()
	if ledgerStore == nil {
		return
	}
	oldSV := pool.stateValidator
	pool.stateValidator = validation.NewStateValidate(ledgerStore)
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

func (pool *TxPool) AddTx(tx *types.Transaction) error {
	result, err := pool.stateValidator.VerifyTx(tx)

	if err != nil {
		return err
	}

	txHash := tx.Hash()

	switch result {
	case -1:
		return err
	case 0: //
		if uint32(len(pool.txChan)) < pool.maxInQueue {
			pool.waitTxNum[txHash]++
			if uint32(pool.waitTxNum[txHash]) == config.GlobalConfig.TxPoolCfg.MaxTxInPool { // 下一个块在考虑  尝试四次都失败后执行
				pool.waitPool = append(pool.waitPool, tx)
			} else if uint32(pool.waitTxNum[txHash]) < config.GlobalConfig.TxPoolCfg.MaxTxInPool {
				pool.txChan <- tx
			}
		} else {
			return errors.ERR_TXPOOL_OUTOFMAX
		}

		return nil
	case 1: //
		fmt.Printf("validate tx success,txhash=%s\n", txHash.String())
		p2ppid, err := bactor.GetActorPid(bactor.P2PACTOR)
		if err != nil {
			log.Error("tx_pool.go get p2ppid error", "error", err)
		} else {
			p2ppid.Tell(tx)
		}
		pool.pdmgr.addTx(tx, pool.maxPending)
		return nil
	}
	return nil
}

func (pool *TxPool) GetTxpoolPID() *actor.PID {
	return pool.txpoolPid
}

func (pool *TxPool) TxEnqueue(tx *types.Transaction) error {
	log.Info("magnetic try to enqueue", "tx", tx, "queue.size", len(pool.txChan))
	//if uint32(pool.queue.Size()) >= pool.maxInQueue {
	//	return errors.ERR_TXPOOL_OUTOFMAX
	//}
	if uint32(len(pool.txChan)) >= pool.maxInQueue {
		return errors.ERR_TXPOOL_OUTOFMAX
	}
	log.Info("magnetic enqueue success", "tx", tx, "queue.size", len(pool.txChan))
	//pool.queue.Enqueue(tx)
	pool.txChan <- tx
	return nil
}

/**
	generate block
*/
func (pool *TxPool) GenerateBlock(height uint64, packtx bool) *types.Block {
	//pool.ticker.Stop()

	sts := states.AccountStates{}
	var txs *types.Transactions
	if packtx {
		txs = pool.pdmgr.getTxs(pool.maxPending)
		if txs != nil && txs.Len() > 0 {
			sort.Sort(txs)
		}
	}
	var txsroot common.Hash
	var txns types.Transactions

	if txs != nil {
		txsroot = txs.GetHashRoot()
		txns = *txs
	}

	block := &types.Block{
		Header: &types.Header{
			Height:        height + 1,
			Version:       types.TxVersion,
			PrevBlockHash: storages.GetLedgerStore().GetCurrentBlockHash(),
			LeagueRoot:    common.Hash{},
			ReceiptsRoot:  common.Hash{},
			TxRoot:        txsroot,
			StateRoot:     sts.GetHashRoot(),
			Timestamp:     uint64(time.Now().Unix()),
		},
		Transactions: txns,
	}

	return block
}

func (pool *TxPool) ValidateSyncTxs(txhashes []*common.Hash) error {

	if pool.pdmgr.pendingTxs == nil || len(pool.pdmgr.pendingTxs) < 1 {
		return errors.ERR_TXPOOL_TXNOTFOUND
	}

	//for _, v := range txhashes {
	//vtx := pool.queue.Remove(*v)

	//if vtx == nil {
	//	return errors.ERR_TXPOOL_TXNOTFOUND
	//}

	//result, err := pool.stateValidator.VerifyTx(vtx)
	//
	//switch result {
	//case -1:
	//	return err
	//case 0:
	//	if uint32(pool.queue.Size()) < pool.maxInQueue {
	//		pool.queue.Enqueue(vtx)
	//	} else {
	//		return errors.ERR_TXPOOL_OUTOFMAX
	//	}
	//	return nil
	//case 1:
	//	pool.pdmgr.addTx(vtx, pool.maxPending)
	//	return nil
	//}
	//}
	return nil
}

func (pool *TxPool) Execute() *storages.BlockInfo {
	log.Info("func txpool Execute", "targetHeight", pool.stateValidator.TargetHeight)

	/*for _, tx := range pool.mustPackTxs {
		ret, err := pool.stateValidator.VerifyTx(tx)
		if err != nil {
			log.Error("range pool.mustPackTxs verifyTx error", "ret", ret, "error", err)
		}
	}*/
	objectiveTxs, err := pool.mainRadar.GenerateMainTxs()
	if err != nil {
		log.Error("GenerateMainTxs  failed", "error", err)
	}
	for _, tx := range objectiveTxs {
		if validation.AutoNonceContains(tx.TxType) {
			nonce := validation.AccountNonceInstance.GetAccountNonce(tx.TxData.From)
			tx.TxData.Nonce = nonce + 1
			validation.AccountNonceInstance.SetNonce(tx.TxData.From, nonce+1)
			fmt.Println("🚰 Execute txhash", tx.TxData.LeagueId.ToString())
		}
		ret, err := pool.stateValidator.VerifyTx(tx)
		if err != nil {
			log.Error("range pool.mustPackTxs verifyTx error", "ret", ret, "error", err)
		}
		if ret != 1 {
			fmt.Println("🚰 🚰 🚰 objTx verify failed!!! result:", ret, tx.TxType)
		}
		if tx.TxType == types.ConsensusLeague {
			fmt.Printf("🚰 🚰 🚰  ♋️ leagueId:%s startheight:%d endheight:%d energy:%s blockroot:%s\n",
				tx.TxData.LeagueId.ToString(),
				tx.TxData.StartHeight,
				tx.TxData.EndHeight,
				tx.TxData.Energy.String(),
				tx.TxData.BlockRoot.String())
		}
	}
	return pool.stateValidator.ExecuteOplogs()

}

func (pool *TxPool) RemovePendingTxs(hashes []common.Hash) {
	pool.pdmgr.removePendingTxs(hashes)
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

/*
	get txs from pending transaction list
*/
func (pm *PendingMgr) removePendingTxs(txhashes []common.Hash) {

	pm.RLock()
	defer pm.RUnlock()

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
func (pm *PendingMgr)   addTx(tx *types.Transaction, maxPending uint32) error {

	pm.Lock()
	defer pm.Unlock()

	if uint32(len(pm.pendingTxs)) > maxPending {
		return errors.ERR_TXPOOL_OUTOFMAX
	}

	pm.pendingTxs[tx.Hash()] = tx
	return nil
}
