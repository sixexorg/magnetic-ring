package dpoa

import (
	"errors"
	"sync"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	"github.com/sixexorg/magnetic-ring/log"
)



var errDropFarFutureMsg = errors.New("msg pool dropped msg for far future")

type ConsensusRoundMsgs map[comm.MsgType][]comm.ConsensusMsg // indexed by MsgType (proposal, endorsement, ...)

type ConsensusRound struct {
	blockNum uint64
	msgs     map[comm.MsgType][]interface{}
	msgHashs map[common.Hash]interface{} // for msg-dup checking
}

func newConsensusRound(num uint64) *ConsensusRound {

	r := &ConsensusRound{
		blockNum: num,
		msgs:     make(map[comm.MsgType][]interface{}),
		msgHashs: make(map[common.Hash]interface{}),
	}

	r.msgs[comm.ConsensePrepare] = make([]interface{}, 0)
	r.msgs[comm.ConsensePromise] = make([]interface{}, 0)
	r.msgs[comm.ConsenseProposer] = make([]interface{}, 0)
	r.msgs[comm.ConsenseAccept] = make([]interface{}, 0)
	r.msgs[comm.EvilEvent] = make([]interface{}, 0)

	return r
}

func (self *ConsensusRound) addMsg(msg comm.ConsensusMsg, msgHash common.Hash) error {
	if _, present := self.msgHashs[msgHash]; present {
		return nil
	}

	msgs := self.msgs[msg.Type()]
	self.msgs[msg.Type()] = append(msgs, msg)
	self.msgHashs[msgHash] = nil
	return nil
}

func (self *ConsensusRound) hasMsg(msg comm.ConsensusMsg, msgHash common.Hash) (bool, error) {
	if _, present := self.msgHashs[msgHash]; present {
		return present, nil
	}
	return false, nil
}

type timeoutData struct {
	tmSigs      map[string][]byte // pub sigs
	established bool              //
}

type MsgPool struct {
	lock        sync.RWMutex
	cacheLen    uint64
	rounds      map[uint64]*ConsensusRound         // indexed by BlockNum
	earthSigs   map[common.Hash]map[string][]byte  // hash pub sigs
	timeoutSigs map[uint64]map[uint32]*timeoutData // blknum view pub
	msgHashs    map[common.Hash]interface{}        // for msg-dup checking
}

func newMsgPool(cacheLen uint64) *MsgPool {
	// TODO
	return &MsgPool{
		lock:        sync.RWMutex{},
		cacheLen:    cacheLen,
		rounds:      make(map[uint64]*ConsensusRound),
		msgHashs:    make(map[common.Hash]interface{}),
		earthSigs:   make(map[common.Hash]map[string][]byte),
		timeoutSigs: make(map[uint64]map[uint32]*timeoutData),
	}
}

func (pool *MsgPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.rounds = make(map[uint64]*ConsensusRound)
}

func (pool *MsgPool) MarkMsgs(msgHash common.Hash) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.msgHashs[msgHash] = nil
}

func (pool *MsgPool) CheckMsgs(msgHash common.Hash) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	if _, ok := pool.msgHashs[msgHash]; ok {
		return true
	}
	return false
}

func (pool *MsgPool) AddMsg(msg comm.ConsensusMsg, msgHash common.Hash, argv ...interface{}) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	//if blkNum > pool.server.GetCurrentBlockNo()+pool.cacheLen {
	//	return errDropFarFutureMsg
	//}

	// TODO: limit #history rounds to cacheLen
	// Note: we accept msg for future rounds
	if msg.Type() == comm.EarthFetchRspSigs {
		pMsg := msg.(*comm.EarthSigsFetchRspMsg)
		if _, ok := pool.earthSigs[pMsg.BlkHash]; !ok {
			pool.earthSigs[pMsg.BlkHash] = make(map[string][]byte)
		}
		pool.earthSigs[pMsg.BlkHash][pMsg.PubKey] = pMsg.Sigs
		pool.msgHashs[msgHash] = nil
		return nil
	}

	if msg.Type() == comm.ViewTimeout { // Msg type - []list(msglist)
		var partiCfg *comm.ParticiConfig
		if len(argv) == 1 {
			partiCfg = argv[0].(*comm.ParticiConfig)
		}
		pMsg := msg.(*comm.ViewtimeoutMsg)
		if _, ok := pool.timeoutSigs[blkNum]; !ok {
			pool.timeoutSigs[blkNum] = make(map[uint32]*timeoutData)
		}
		if _, ok := pool.timeoutSigs[blkNum][pMsg.RawData.View]; !ok {
			pool.timeoutSigs[blkNum][pMsg.RawData.View] = &timeoutData{tmSigs: make(map[string][]byte)}
		}
		//len(pool.server.dpoaMgr.partiCfg.obserNodes)
		pool.timeoutSigs[blkNum][pMsg.RawData.View].tmSigs[pMsg.PubKey] = pMsg.Signature
		if partiCfg.View == pMsg.RawData.View {
			if len(pool.timeoutSigs[blkNum][pMsg.RawData.View].tmSigs[pMsg.PubKey]) > len(partiCfg.ObserNodes)/2 {
				pool.timeoutSigs[blkNum][pMsg.RawData.View].established = true
			}
		}
		pool.msgHashs[msgHash] = nil
		return nil
	}

	if _, present := pool.rounds[blkNum]; !present {
		pool.rounds[blkNum] = newConsensusRound(blkNum)
	}

	return pool.rounds[blkNum].addMsg(msg, msgHash)
}

func (pool *MsgPool) HasMsg(msg comm.ConsensusMsg, msgHash common.Hash) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if roundMsgs, present := pool.rounds[msg.GetBlockNum()]; !present {
		return false
	} else {
		if present, err := roundMsgs.hasMsg(msg, msgHash); err != nil {
			log.Error("msgpool failed to check msg avail: %s", err)
			return false
		} else {
			return present
		}
	}

	return false
}

func (pool *MsgPool) Persist() error {
	// TODO
	return nil
}

func (pool *MsgPool) TimeoutCount(blocknum uint64) int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	//fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!&&&&&&&&&&&&&^^^^^^^^^^^^TimeoutCount blocknum:", blocknum)
	if _, ok := pool.timeoutSigs[blocknum]; !ok {
		return 0
	}

	var count int = 0
	for _, v := range pool.timeoutSigs[blocknum] { //view
		if v.established {
			/*for k1, _ := range v.tmSigs {
				fmt.Println("!!!!!!!!!!!start&&&&&&&&&&&&&^^^^^^^^^^^^TimeoutCount blocknum:", "view", k, "pub", k1)
			}*/
			count++
		}
	}

	//fmt.Println("&&&&&&&&&&&&&^^^^^^^^^^^^TimeoutCount blocknum:", blocknum, "count:", count)
	return count
}

func (pool *MsgPool) GetTimeoutMsgs(blocknum uint64) map[uint32]*timeoutData {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	if v, ok := pool.timeoutSigs[blocknum]; !ok {
		return make(map[uint32]*timeoutData)
	} else {
		return v
	}
}

func (pool *MsgPool) GetEvilMsgs(blocknum uint64) []interface{} {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msgs, ok := roundMsgs.msgs[comm.EvilEvent]
	if !ok {
		return nil
	}
	return msgs
}

func (pool *MsgPool) GetEarthMsgs(prehash common.Hash) map[string][]byte {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	msgs, ok := pool.earthSigs[prehash]
	if !ok {
		return nil
	}

	return msgs
}

func (pool *MsgPool) onBlockSealed(blockNum uint64) {
	if blockNum <= pool.cacheLen {
		return
	}
	pool.lock.Lock()
	defer pool.lock.Unlock()

	toFreeRound := make([]uint64, 0)
	for n := range pool.rounds {
		if n < blockNum-pool.cacheLen {
			toFreeRound = append(toFreeRound, n)
		}
	}
	for _, n := range toFreeRound {
		delete(pool.rounds, n)
	}
}

func (pool *MsgPool) IncInsert(blockNum uint32, msgtype comm.MsgType, msg interface{}) {
}
