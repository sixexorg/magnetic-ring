package mainchain

import (
	"sort"
	"sync"
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/log"
)

/*
	Transaction queue
*/
type TxQueue struct {
	sync.RWMutex
	txs      map[int64]*queueTx
	txhashMp map[common.Hash]int64
}

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type queueTx struct {
	tx     *types.Transaction
	quTime int64
}

func newQueueTx(tx *types.Transaction) *queueTx {
	qt := new(queueTx)
	qt.tx = tx
	qt.quTime = time.Now().UnixNano()
	return qt
}

func NewTxQueue() *TxQueue {
	queue := new(TxQueue)
	queue.txs = make(map[int64]*queueTx)
	queue.txhashMp = make(map[common.Hash]int64)
	return queue
}

/**
  Add a deal to the end of the queue
*/
func (queue *TxQueue) Enqueue(tx *types.Transaction) {
	queue.Lock()
	defer queue.Unlock()
	qt := newQueueTx(tx)
	queue.txs[qt.quTime] = qt
	queue.txhashMp[tx.Hash()] = qt.quTime
	log.Info("current queue", "map=", queue.txs)
}

/**
	Remove and remove the transaction from the queue head
*/
func (queue *TxQueue) Dequeue() *types.Transaction {
	if len(queue.txs) < 1 {
		return nil
	}
	queue.Lock()
	defer queue.Unlock()

	timearr := make(Int64Slice, 0)

	for k, _ := range queue.txs {
		timearr = append(timearr, k)
	}

	sort.Sort(timearr)

	got := queue.txs[timearr[0]]
	delete(queue.txs, got.quTime)
	delete(queue.txhashMp, got.tx.Hash())

	return got.tx
}

func (queue *TxQueue) Remove(hash common.Hash) *types.Transaction {
	if len(queue.txs) < 1 {
		return nil
	}
	queue.Lock()
	defer queue.Unlock()

	if tm, ok := queue.txhashMp[hash]; ok {
		got := queue.txs[tm]
		delete(queue.txs, got.quTime)
		delete(queue.txhashMp, got.tx.Hash())
		return got.tx
	}

	return nil
}

/**
	Determine if the transaction is empty
*/
func (q *TxQueue) IsEmpty() bool {
	return len(q.txs) == 0
}

/**
	Get the current number of transactions in the queue
*/
func (pool *TxQueue) Size() int {
	return len(pool.txs)
}
