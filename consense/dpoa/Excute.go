package dpoa

import (
	"fmt"
	"sync"
	"errors"

	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
)

type actObj struct {
	etype   comm.ExcuteType
	action  func()
}


type SyncExec struct {
	sync.RWMutex
	execType map[comm.ExcuteType]struct{}
	actions chan actObj
	quit    chan struct{}
}

func newExcutePool(no int) *SyncExec {
	p := &SyncExec{
		actions:  make(chan actObj),
		execType: make(map[comm.ExcuteType]struct{}),
		quit:     make(chan struct{}),
	}

	for i := 0; i < no; i++ {
		go p.loop(i, p.actions)
	}
	return p
}

func (p *SyncExec) loop(i int, actions <-chan actObj) {
	defer log.Info("SyncExec loop stop", "i", i)

	for {
		select {
		case actobj := <-actions:
			//log.Info("SyncExec loop start", "i", i)
			fmt.Println("SyncExec loop start", "i", i)
			actobj.action()
			p.Lock()
			delete(p.execType, actobj.etype)
			p.Unlock()
			//log.Info("SyncExec loop stop", "i", i)
			fmt.Println("SyncExec loop stop", "i", i)
		case <-p.quit:
			return
		}
	}
}

func (p *SyncExec) add(etype comm.ExcuteType, action func()) {
	p.Lock()
	defer p.Unlock()
	if _, ok := p.execType[etype];ok{
		return
	}
	p.execType[etype] = struct{}{}
	p.actions <- actObj{etype:etype, action:action}
}

func (p *SyncExec) close() {
	close(p.quit)
}

func (self *ProcMode)Excute(etype comm.ExcuteType, action func()) error{
	if actPool, ok := self.excuteSet[etype]; ok{
		actPool.add(etype, action)
		return nil
	}
	return errors.New(fmt.Sprintf("Server Excute ExcuteType err %v", etype))
}