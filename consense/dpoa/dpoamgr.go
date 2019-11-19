package dpoa

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/bactor"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/log"
	txpool "github.com/sixexorg/magnetic-ring/txpool/mainchain"
)

type PaxState uint8

const (
	PaxStateMin PaxState = iota
	PaxStateInit
	PaxStateWaiting
	PaxStateProcessing
	PaxStateObserving
	PaxStateDone
	PaxStateTimeout
	PaxStateMax
)

type PaxRole uint8

const (
	PaxRoleMin PaxRole = iota
	PaxRoleProcesser
	PaxRoleObserver
	PaxRoleFail
	PaxRoleMax
)

type DpoaMgr struct {
	sync.RWMutex
	txpool   *txpool.TxPool
	partiCfg *comm.ParticiConfig
	paxState PaxState
	role     PaxRole
	recvCh   chan interface{}
	sendCh   chan interface{}
	handles  map[string]reflect.Value
	paxosIns *Paxos
	cfg      *Config
	store    *BlockStore
	msgpool  *MsgPool
	annPid   *actor.PID
	p2pPid   *actor.PID
}

func NewdpoaMgr(cfg *Config, store *BlockStore, msgpool *MsgPool, p2pActor *actor.PID) *DpoaMgr {
	txpool, _ := txpool.GetPool()
	p := &DpoaMgr{cfg: cfg, handles: make(map[string]reflect.Value), paxState: PaxStateInit, partiCfg: &comm.ParticiConfig{},
		sendCh: make(chan interface{}, 100), recvCh: make(chan interface{}, 100), store: store, msgpool: msgpool, txpool: txpool}
	p.paxosIns = NewPaxos(cfg.accountStr, cfg.account, p.sendCh)
	p.register(comm.P1a{}, p.paxosIns.HandleP1a)
	p.register(comm.P1b{}, p.paxosIns.HandleP1b)
	p.register(comm.P2a{}, p.paxosIns.HandleP2a)
	p.register(comm.P2b{}, p.paxosIns.HandleP2b)
	p.annPid, _ = bactor.GetActorPid(bactor.MAINRADARACTOR)
	p.p2pPid = p2pActor
	return p
}

func (n *DpoaMgr) register(m interface{}, f interface{}) {
	t := reflect.TypeOf(m)
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func || fn.Type().NumIn() != 1 || fn.Type().In(0) != t {
		fmt.Println("----------------------------", fn.Kind() != reflect.Func, fn.Type().NumIn() != 1, fn.Type().In(0))
		panic("register handle function error")
	}
	n.handles[t.String()] = fn
}

func (n *DpoaMgr) Run() {
	log.Info("node start running", "accountStr:", n.cfg.accountStr)
	go n.recv()
}

// recv receives messages from socket and pass to message channel
func (n *DpoaMgr) recv() {
	for {
		select {
		case m := <-n.recvCh:
			if n.paxState != PaxStateProcessing {
				continue
			}
			v := reflect.ValueOf(m)
			name := v.Type().String()
			//fmt.Println("==>node handle", v.Type().String(), msg)
			f, exists := n.handles[name]
			if !exists {
				log.Error("no registered handle function for message type %v", name)
			}
			f.Call([]reflect.Value{v})
			//case blkNum := <-n.annNewBlock:
			//	if blkNum == n.partiCfg.blkNum {
			//		close(n.workCh)
			//	}
		}
	}
}

func (n *DpoaMgr) SendCh() chan<- interface{} {
	return n.recvCh
}
func (n *DpoaMgr) SendData(data interface{}) {
	n.recvCh <- data
}

func (n *DpoaMgr) RecvCh() <-chan interface{} {
	return n.sendCh
}

type PaxosSt struct {
	state   PaxState
	patiCfg *comm.ParticiConfig
}

func (n *DpoaMgr) GetpaxState() *PaxosSt {
	n.RLock()
	defer n.RUnlock()

	return &PaxosSt{state: n.paxState, patiCfg: n.partiCfg}
}

func (n *DpoaMgr) updatePartions(timeoutCount int, blkNum uint64) error {
	n.paxState = PaxStateInit
	stars := n.store.GetCurStars()
	ebgHeight := n.store.EpochBegin()
	lastBlk, _ := n.store.getSealedBlock(blkNum - 1)
	var times int
	if blkNum-1 == ebgHeight {
		times = timeoutCount
	} else {
		times = int(lastBlk.GetViews()) + timeoutCount + 1
	}
	partiNums, epochViews := CalcStellar(float64(len(stars)))
	if times > epochViews-1 {
		return errors.New(fmt.Sprintf("genblock completee, wait earth block"))
	}

	if blkNum-1 == ebgHeight || n.partiCfg.BlkNum < ebgHeight {
		n.partiCfg = &comm.ParticiConfig{
			PartiRaw:   stars,
			ProcNodes:  make([]string, 0),
			ObserNodes: make([]string, 0),
			FailsNodes: make([]string, 0),
		}
		var blkData *comm.Block
		if ebgHeight == 0 { // 创世
			blkData, _ = n.store.getGeisisBlock()
		} else {
			blkData, _ = n.store.getSealedBlock(ebgHeight)
		}
		vrfValue := getParticipantSelectionSeed(blkData)
		if vrfValue.IsNil() {
			return errors.New(fmt.Sprintf("DpoaMgr updatePartions getParticipantSelectionSeed vrf is nil"))
		}
		n.partiCfg.StarsSorted = calcParticipantPeers(vrfValue, stars)
	}

	log.Info("DpoaMgr updatePartions",
		"timeoutCount", timeoutCount, "blkNum", blkNum, "partiNums", partiNums, "epochViews",epochViews, "times",times, "ebgHeight",ebgHeight, "len(n.partiCfg.StarsSorted)",len(n.partiCfg.StarsSorted))
	n.partiCfg.BlkNum = blkNum
	n.partiCfg.View = uint32(times)
	n.partiCfg.ProcNodes = n.partiCfg.StarsSorted[times*partiNums : (times+1)*partiNums]
	n.partiCfg.ObserNodes = n.partiCfg.StarsSorted[(times+1)*partiNums:]
	fmt.Println("---------------------procNodes", n.partiCfg.ProcNodes)
	fmt.Println("---------------------obserNodes", n.partiCfg.ObserNodes)
	n.annPid.Tell(n.partiCfg.ProcNodes)

	if err := n.updateRole(); err != nil {
		//log.Error("DpoaMgr startwork updateRole err:%v", err)
		return err
	}

	return nil
}

func (n *DpoaMgr) genBlockData() *comm.Block {
	timeoutSigs := make([][]byte, 0)
	for view, data := range n.msgpool.GetTimeoutMsgs(n.store.db.GetCurrentBlockHeight() + 1) {
		if data.established {
			for publicKey, sig := range data.tmSigs {
				var d []byte
				d = append(d, int2Byte(uint16(view))...)
				idx, _ := GetIndex(n.store.GetCurStars(), publicKey)
				fmt.Println("----------timeout", publicKey, idx, view, n.store.db.GetCurrentBlockHeight()+1)
				d = append(d, int2Byte(uint16(idx))...)
				d = append(d, sig...)
				timeoutSigs = append(timeoutSigs, d)
			}
		}
	}

	prevBlk, _ := n.store.getSealedBlock(n.store.db.GetCurrentBlockHeight())
	//fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$4", n.store.db.GetCurrentBlockHeight(), prevBlk.Block.Header.Hash().String())
	if prevBlk == nil {
		//return nil, fmt.Errorf("failed to get prevBlock (%d)", blkNum-1)
	}
	blocktimestamp := uint64(time.Now().Unix())
	if prevBlk.Block.Header.Timestamp >= blocktimestamp {
		blocktimestamp = prevBlk.Block.Header.Timestamp + 1
	}
	vrfValue, vrfProof, err := computeVrf(n.cfg.account.(*account.NormalAccountImpl).PrivKey, n.store.db.GetCurrentBlockHeight()+1, prevBlk.GetVrfValue())
	if err != nil {

		//return nil, fmt.Errorf("failed to get vrf and proof: %s", err)
		log.Error("computeVrf failed to get vrf and proof", "err", err)
		return nil
	}

	lastConfigBlkNum := prevBlk.Info.LastConfigBlockNum
	if prevBlk.Info.NewChainConfig != nil {
		lastConfigBlkNum = prevBlk.GetBlockNum()
	}

	vbftBlkInfo := &comm.VbftBlockInfo{
		View:               n.partiCfg.View,
		Miner:              n.cfg.accountStr,
		VrfValue:           vrfValue,
		VrfProof:           vrfProof,
		LastConfigBlockNum: lastConfigBlkNum,
	}
	consensusPayload, _ := json.Marshal(vbftBlkInfo)
	log.Info("func dpoa genBlockData begin")

	//  []map[common.Address]uint64
	// GenerateMainTx（[]map[common.Address]uint64）
	blkinfo := n.txpool.Execute()

	log.Info("func dpoa genBlockData", "blockHeight", blkinfo.Block.Header.Height, "txlen", blkinfo.Block.Transactions.Len())
	fmt.Println("func dpoa genBlockData", "blockHeight", blkinfo.Block.Header.Height, "txlen", blkinfo.Block.Transactions.Len(), blkinfo.Block.Header.Hash().String())

	blkinfo.Block.Header.Timestamp = blocktimestamp
	blkinfo.Block.Header.ConsensusPayload = consensusPayload
	blkinfo.Block.Sigs = &types.SigData{TimeoutSigs: timeoutSigs, FailerSigs: make([][]byte, 0), ProcSigs: make([][]byte, 0)}
	//fmt.Println("========--------->>>>>>>>>>", blkinfo.Block.Header.Height, n.store.getLatestBlockNumber())
	block, _ := comm.InitVbftBlock(blkinfo.Block)

	return block
}

func (n *DpoaMgr) startwork(timeoutCount int, blkNum uint64, stopCh chan struct{}) {
	if err := n.updatePartions(timeoutCount, blkNum); err != nil {
		//log.Error("DpoaMgr startwork updatePartions err:%v", err)
		return
	}
	notice := &comm.ConsenseNotify{BlkNum: n.partiCfg.BlkNum, ProcNodes: n.partiCfg.ProcNodes, Istart: true}
	n.p2pPid.Tell(notice)

	go func() {
		switch n.role {
		case PaxRoleProcesser:
			//n.handleProcess()
			n.paxosIns.partiCfg = n.partiCfg
			n.paxState = PaxStateProcessing
			log.Info("func dpoa dpoamgr startwork")
			n.paxosIns.ConsensProcess(n.genBlockData(), 100, n.partiCfg, stopCh)
			n.paxState = PaxStateTimeout
		case PaxRoleObserver:
			n.paxState = PaxStateObserving
			select {
			case <-stopCh:
				n.paxState = PaxStateTimeout
			}
		case PaxRoleFail:
		default:
			log.Info("DpoaMgr handle default", "acc", n.cfg.accountStr, "blknum", n.partiCfg.BlkNum, "view", n.partiCfg.View)
		}
	}()
}

func (n *DpoaMgr) updateRole() (err error) {
	defer func() {
		if err == nil {
			log.Info("DpoaMgr updateRole ", "accountStr", n.cfg.accountStr, "BlkNum", n.partiCfg.BlkNum, "View", n.partiCfg.View, "role", n.role)
		}
	}()

	publicStr := n.cfg.accountStr
	n.role = PaxRoleMax

	for _, v := range n.partiCfg.ObserNodes {
		if v == publicStr {
			n.role = PaxRoleObserver
			return
		}
	}

	for _, v := range n.partiCfg.ProcNodes {
		if v == publicStr {
			n.role = PaxRoleProcesser
			return
		}
	}

	for _, v := range n.partiCfg.FailsNodes {
		if v == publicStr {
			n.role = PaxRoleFail
			return
		}
	}

	err = errors.New(fmt.Sprintf("DpoaMgr updateRole unknow role, obser:%v proc:%v fail:%v ",
		n.partiCfg.ObserNodes, n.partiCfg.ProcNodes, n.partiCfg.FailsNodes))
	return
}
