package miner

import (
	"sync"

	"sync/atomic"

	"github.com/sixexorg/magnetic-ring/consense/poa"
	"github.com/sixexorg/magnetic-ring/log"
)

type CpuAgent struct {
	mu sync.Mutex

	workCh        chan *Work
	stop          chan struct{}
	quitCurrentOp chan struct{}
	returnCh      chan<- *Result

	poaIns        *poa.Clique
	isMining      int32
}

func NewCpuAgent(poaIns *poa.Clique) *CpuAgent {
	miner := &CpuAgent{
		poaIns: poaIns,
		stop:   make(chan struct{}, 1),
		workCh: make(chan *Work, 1),
	}
	return miner
}

func (self *CpuAgent) Work() chan<- *Work            { return self.workCh }
func (self *CpuAgent) SetReturnCh(ch chan<- *Result) { self.returnCh = ch }

func (self *CpuAgent) Stop() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 1, 0) {
		return // agent already stopped
	}
	self.stop <- struct{}{}
done:
	// Empty work channel
	for {
		select {
		case <-self.workCh:
		default:
			break done
		}
	}
}

func (self *CpuAgent) Start() {
	if !atomic.CompareAndSwapInt32(&self.isMining, 0, 1) {
		return // agent already started
	}
	go self.update()
}

func (self *CpuAgent) update() {
out:
	for {
		select {
		case work := <-self.workCh:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
			}
			self.quitCurrentOp = make(chan struct{})
			go self.mine(work, self.quitCurrentOp) // 挖矿计算
			self.mu.Unlock()
		case <-self.stop:
			self.mu.Lock()
			if self.quitCurrentOp != nil {
				close(self.quitCurrentOp)
				self.quitCurrentOp = nil
			}
			self.mu.Unlock()
			break out
		}
	}
}

func (self *CpuAgent) mine(work *Work, stop <-chan struct{}) {
	if result, err := self.poaIns.Seal(work.Block, stop); result != nil {
		log.Info("Successfully sealed new block", "number", result.Header.Height, "hash", result.Hash())
		self.returnCh <- &Result{work, result}
	} else {
		if err != nil {
			log.Warn("Block sealing failed", "err", err)
		}
		self.returnCh <- nil
	}
}
