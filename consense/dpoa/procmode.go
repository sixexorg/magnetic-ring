package dpoa

import (
	"encoding/json"
	"fmt"
	"time"

	//"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/log"

	//"github.com/sixexorg/magnetic-ring/core/ledger"
	//"github.com/sixexorg/magnetic-ring/core/signature"
	//"github.com/ontio/ontology-crypto/keypair"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	//"github.com/sixexorg/magnetic-ring/consensus/vbft/config"
	"sync"

	"github.com/sixexorg/magnetic-ring/account"
)

type ProcMode struct {
	sync.RWMutex
	dpoaMgr     *DpoaMgr
	msgpool     *MsgPool
	cfg         *Config
	notifyBlock *Feed
	notifyState *Feed
	state       ServerState
	prestate    ServerState
	trans       *TransAction
	excuteSet   map[comm.ExcuteType]*SyncExec
	quitC       chan struct{}
	stopNewHtCh chan struct{}
}

func NewprocMode(dpoaMgr *DpoaMgr, msgpool *MsgPool, trans *TransAction, notifyBlock, notifyState *Feed, cfg *Config) *ProcMode {
	return &ProcMode{dpoaMgr: dpoaMgr, msgpool: msgpool, notifyBlock: notifyBlock, notifyState: notifyState, cfg: cfg, trans: trans,
		stopNewHtCh: make(chan struct{}), excuteSet: make(map[comm.ExcuteType]*SyncExec, comm.MaxExcuteType)}
}

func (srv *ProcMode) Start() {
	subCh1 := make(chan stateChange, 100)
	srv.notifyState.Subscribe(subCh1)

	go func() {
		//fmt.Println("***************ProcMode Start", srv.prestate, srv.state)
		for {
			select {
			case st := <-subCh1:
				fmt.Println("==========---------->>>>>>>>>state", srv.prestate, srv.state, st.currentState)
				srv.Lock()
				srv.prestate = srv.state
				srv.state = st.currentState
				srv.Unlock()
				if st.currentState >= SyncReady {
					if srv.prestate < SyncReady {
						srv.Excute(comm.StartNewHeight, srv.NewEpoch)
					}
				}
			case <-srv.quitC:
				return
			}
		}
	}()

	for e := comm.MinExcuteType + 1; e < comm.MaxExcuteType; e++ {
		srv.excuteSet[e] = newExcutePool(5)
	}

	if !srv.dpoaMgr.store.isEarth() {
		go func() {
			for {
				select {
				case m := <-srv.dpoaMgr.RecvCh():
					//log.Info("func dpoa procmode start 01", "recvch type", reflect.TypeOf(m))
					switch v := m.(type) {
					case *SendMsgEvent:
						//log.Info("func dpoa procmode start 02","msg type", v.Msg.Type())

						srv.trans.sendMsg(v)
					case *comm.Block:
						log.Info("func dpoa procmode start 03", "blockHeight", v.Block.Header.Height, "txlen", v.Block.Transactions.Len())
						srv.dpoaMgr.store.sealBlock(v)
					default:
						//fmt.Println("b1.(type):", "other", v)
					}
				case <-srv.quitC:
					return
				}
			}
		}()
	}
}

func (srv *ProcMode) currentState() ServerState {
	srv.RLock()
	defer srv.RUnlock()
	return srv.state
}

func (srv *ProcMode) Process() {
	log.Info("func dpoa procmode Process", "isearth", srv.dpoaMgr.store.isEarth())
	srv.dpoaMgr.Run()
	if srv.dpoaMgr.store.isEarth() {
		srv.Excute(comm.EarthProcess, srv.earthProcess)
	}
}

func (srv *ProcMode) earthProcess() {
	var exitDesc string
	//srv.stopNewHtCh = make(chan struct{})
	subCh := make(chan types.Block, 100)
	subIns := srv.notifyBlock.Subscribe(subCh)
	log.Info("Server earthProcess start")
	defer func() {
		log.Info("Server earthProcess exit due to", "desc", exitDesc)
		subIns.Unsubscribe()
	}()

	for {
		//log.Info("-----------------&&&&&&&&&&&&&&&&&&&&&&", "currentState", srv.currentState(), "time", time.Now().String())
		time.Sleep(time.Second)
		stars := srv.dpoaMgr.store.GetCurStars()
		ebgHeight := srv.dpoaMgr.store.EpochBegin()
		blkData, _ := srv.dpoaMgr.store.getSealedBlock(ebgHeight)
		partiNums, epochView := CalcStellar(float64(len(stars)))
		endtime := time.Unix(int64(blkData.Block.Header.Timestamp), 0).Add(time.Duration(srv.cfg.earthCfg.duration) * time.Second).Add(time.Duration(epochView*2*srv.cfg.earthCfg.duration) * time.Second)
		vrfValue := getParticipantSelectionSeed(blkData)
		if vrfValue.IsNil() {
			log.Error("ProcMode earthProcess", "err", fmt.Sprintf("StateMgr earth vrf is nil"))
			return
		}

		if srv.currentState() <= WaitNetworkReady {
			continue
		}

		delay := endtime.Sub(time.Now().Add(time.Duration(srv.cfg.earthCfg.duration) * time.Second))
		log.Info("Server earthProcess wait duration",
			"delay", delay, "ebgHeight", ebgHeight, "partiNums", partiNums, "epochView", epochView, "endtime", endtime, "time", time.Unix(int64(blkData.Block.Header.Timestamp), 0).String(), "view", blkData.GetViews())
		fmt.Println("Server earthProcess wait duration",
			"delay", delay, "ebgHeight", ebgHeight, "partiNums", partiNums, "epochView", epochView, "endtime", endtime, "time", time.Unix(int64(blkData.Block.Header.Timestamp), 0).String(), "view", blkData.GetViews())
		select {
		case <-subIns.Err():
			break
		case <-time.After(delay):
			log.Info("Server earthProcess start search earthsigs", "ebgHeight", ebgHeight, "view", blkData.GetViews())
			fmt.Println("Server earthProcess start search earthsigs", "time", time.Now().String(), "ebgHeight", ebgHeight, "view", blkData.GetViews())
			select {
			case <-time.After(time.Second * time.Duration(srv.cfg.earthCfg.duration)):
				log.Info("---.Server earthProcess construct block due to epoch timeout", "ebgHeight", ebgHeight, "getLatestBlockNumber", srv.dpoaMgr.store.getLatestBlockNumber()+1)
				fmt.Println("---.Server earthProcess construct block due to epoch timeout", "time", time.Now().String(), "ebgHeight", ebgHeight, "getLatestBlockNumber", srv.dpoaMgr.store.getLatestBlockNumber()+1)
				failerSigs := make([][]byte, 0)
				for publicKey, sig := range srv.msgpool.GetEarthMsgs(blkData.Block.Hash()) {
					var d []byte
					idx, _ := GetIndex(srv.dpoaMgr.store.GetCurStars(), publicKey)
					//fmt.Println("----------fairSigs", publicKey, idx, blkData.GetBlockNum(), blkData.Block.Hash())
					d = append(d, int2Byte(uint16(idx))...)
					d = append(d, sig...)
					failerSigs = append(failerSigs, d)
				}
				//fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", ebgHeight, time.Now().String(), srv.dpoaMgr.store.getLatestBlockNumber()+1)
				blk, err := srv.dpoaMgr.constructEmptyBlock(srv.dpoaMgr.store.getLatestBlockNumber()+1, &types.SigData{TimeoutSigs: make([][]byte, 0), FailerSigs: failerSigs, ProcSigs: make([][]byte, 0)})
				if err != nil {
					log.Error("Server earthProcess constructBlockMsg", "GetBlockNum", blk.GetBlockNum(), "EmptyBlockNum", "err", err)
					continue
				}
				fmt.Println("🌏  Empty block product,the height is ", blk.Block.Header.Height)
				if err := srv.dpoaMgr.store.sealBlock(blk); err != nil {
					log.Error("Server earthProcess sealBlock", "GetBlockNum", blk.GetBlockNum(), "err", err)
				}
				<-subCh
			}
		case <-srv.quitC:
			exitDesc = "server quit"
			return
		}
	}
}

func (self *DpoaMgr) constructEmptyBlock(blkNum uint64, sigData *types.SigData) (*comm.Block, error) {
	prevBlk, _ := self.store.getSealedBlock(blkNum - 1)
	//fmt.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$4", blkNum, prevBlk.Block.Header.Hash().String())
	if prevBlk == nil {
		return nil, fmt.Errorf("failed to get prevBlock (%d)", blkNum-1)
	}
	blocktimestamp := uint64(time.Now().Unix())
	if prevBlk.Block.Header.Timestamp >= blocktimestamp {
		blocktimestamp = prevBlk.Block.Header.Timestamp + 1
	}
	vrfValue, vrfProof, err := computeVrf(self.cfg.account.(*account.NormalAccountImpl).PrivKey, blkNum, prevBlk.GetVrfValue())
	if err != nil {
		return nil, fmt.Errorf("failed to get vrf and proof: %s", err)
	}

	lastConfigBlkNum := prevBlk.Info.LastConfigBlockNum
	if prevBlk.Info.NewChainConfig != nil {
		lastConfigBlkNum = prevBlk.GetBlockNum()
	}

	vbftBlkInfo := &comm.VbftBlockInfo{
		View:               self.partiCfg.View,
		Miner:              self.cfg.accountStr,
		VrfValue:           vrfValue,
		VrfProof:           vrfProof,
		LastConfigBlockNum: lastConfigBlkNum,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil, err
	}

	//fmt.Println("@@@@@@@@@@###########$$$$$$$$$$$$$", string(consensusPayload))
	blkHeader := &types.Header{
		PrevBlockHash: prevBlk.Block.Hash(),
		//TxRoot: txRoot,
		//BlockRoot:        blockRoot,
		Timestamp: blocktimestamp,
		Height:    blkNum,
		//ConsensusData:    common.GetNonce(),
		ConsensusPayload: consensusPayload,
	}
	blk := &types.Block{
		Header: blkHeader,
		Sigs:   sigData,
	}

	return &comm.Block{
		Block: blk,
		Info:  vbftBlkInfo,
	}, nil
}

func (srv *ProcMode) epochPorcess(endtime time.Time, epochViews int, lastBlk, curEpoch *comm.Block) {
	var (
		t0, t1, t2   time.Time
		timeOutCount int
		descStr      string
		blkNum       uint64 = lastBlk.GetBlockNum() + 1
		preNum       uint64 = lastBlk.GetBlockNum()
		//lastBlkTime         = time.Unix(int64(lastBlk.Block.Header.Timestamp), 0)
		curEpochTime = time.Unix(int64(curEpoch.Block.Header.Timestamp), 0)
	)
	subCh := make(chan types.Block, 100)
	subIns := srv.notifyBlock.Subscribe(subCh)

	defer func() {
		log.Debug("--Server epochPorcess exit blknum, cause", "blkNum", blkNum, "cause", descStr)
		subIns.Unsubscribe()
		notice := &comm.ConsenseNotify{BlkNum: srv.dpoaMgr.partiCfg.BlkNum, ProcNodes: srv.dpoaMgr.partiCfg.ProcNodes, Istart: false}
		srv.dpoaMgr.p2pPid.Tell(notice)
	}()

	log.Info("Server epochPorcess start curblknum, epoch begtime  endtime , epochViews ", "blkNum", blkNum, "begtime", curEpochTime.String(), "endtime", endtime.String(), "epochViews", epochViews)

	for {
		if srv.trans.stateMgr.getState() != SyncReady {
			descStr = "state is not ready"
			log.Error("Server epochPorcess state is not SyncReady or Synced", "err is %v", srv.trans.stateMgr.getState())
			break
		}

		if time.Now().After(endtime) {
			descStr = fmt.Sprintf("out epoch endtime %v", endtime.String())
			break
		}

		if blkNum != srv.dpoaMgr.store.getLatestBlockNumber()+1 {
			lastBlk, _ = srv.dpoaMgr.store.getSealedBlock(srv.dpoaMgr.store.db.GetCurrentBlockHeight())
			//lastBlkTime = time.Unix(int64(lastBlk.Block.Header.Timestamp), 0)
			/*fmt.Println("****************..................update waitEvent", blkNum, lastBlkTime, time.Now().Sub(lastBlkTime).Seconds(),
			srv.dpoaMgr.store.getLatestBlockNumber()+1, srv.dpoaMgr.store.getLatestBlockNumber(), srv.dpoaMgr.store.db.GetCurrentBlockHeight())*/
			blkNum = srv.dpoaMgr.store.getLatestBlockNumber() + 1
		}

		timeOutCount = srv.msgpool.TimeoutCount(blkNum)
		if blkNum-1 == curEpoch.GetBlockNum() {
			if timeOutCount == epochViews {
				descStr = fmt.Sprintf("generate block compeletes, blknum %v", blkNum)
				break
			}
			t0 = curEpochTime.Add(time.Duration(comm.V_One*srv.cfg.starsCfg.duration) * time.Second).Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration*timeOutCount) * time.Second)
			t1 = curEpochTime.Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration) * time.Second).Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration*timeOutCount) * time.Second)
			t2 = curEpochTime.Add(time.Duration(comm.V_Three*srv.cfg.starsCfg.duration) * time.Second).Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration*timeOutCount) * time.Second)
		} else {
			t0 = curEpochTime.Add(time.Duration(comm.V_One*srv.cfg.starsCfg.duration) * time.Second).Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration*int(lastBlk.GetViews()+uint32(timeOutCount)+1)) * time.Second)
			t1 = curEpochTime.Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration) * time.Second).Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration*int(lastBlk.GetViews()+uint32(timeOutCount)+1)) * time.Second)
			t2 = curEpochTime.Add(time.Duration(comm.V_Three*srv.cfg.starsCfg.duration) * time.Second).Add(time.Duration(comm.V_Two*srv.cfg.starsCfg.duration*int(lastBlk.GetViews()+uint32(timeOutCount)+1)) * time.Second)
		}

		/*	log.Info("==Server epochPorcess loop latest", "blknum", lastBlk.GetBlockNum(), "view", lastBlk.GetViews(), "blktime", lastBlkTime,
			"t0", t0, "t1", t1, "t2", t2, "curblknum", srv.dpoaMgr.store.getLatestBlockNumber()+1, "timeOutCount", timeOutCount)
		*/
		if srv.dpoaMgr.store.inFailers(curEpoch.GetBlockNum(), srv.cfg.accountStr) {
			if time.Now().Before(endtime) && endtime.Sub(time.Now()).Seconds() < float64(srv.cfg.starsCfg.duration) { // 如果本节点是failer 发送签名超时给地球
				//fmt.Println("----------case3 >>>>>>>>>>>>send earthsigs")
				blk, _ := srv.dpoaMgr.store.getLatestBlock()
				hash := blk.Block.Hash()
				sig, err := srv.cfg.account.Sign(hash[:])
				if err != nil {
					log.Error("sign 352 blk.block.hash() occured error", "error", err)
					return
				}
				msg := constructEarthSigsFetchRspMsg(srv.dpoaMgr.store.getLatestBlockNumber()+1, srv.cfg.accountStr, sig, hash)
				srv.trans.sendMsg(&SendMsgEvent{
					ToPeer: srv.dpoaMgr.store.earthNode(),
					Msg:    msg,
				})
				break
			}
		}
		/*
		Case 1 10-20s
		Case 2 0-10s waiting for the block
		Case 3 >20 send timeout
		*/
		// case 1 starts consensus
		if time.Now().After(t0) && time.Now().Before(t1) {
			stopCh := make(chan struct{})
			srv.dpoaMgr.startwork(timeOutCount, blkNum, stopCh)
		Flag:
			delay := t1.Sub(time.Now())
			//fmt.Println(" ---------case1 update waitEvent", delay, timeOutCount)
			select {
			case <-time.After(delay):
				//fmt.Println(" ---------case1 after")
				select {
				case <-stopCh:
				default:
					close(stopCh)
				}
				//fmt.Println(" ---------case1 after!")
				continue
				//case <-srv.store.epochNotify:
				//case <-srv.annNewBlock:
			case <-subCh:
				if blkNum == srv.dpoaMgr.store.getLatestBlockNumber() { //
					//fmt.Println("提前停止!!!!!!!!!!!!!!!")
					select {
					case <-stopCh:
					default:
						close(stopCh)
					}
				}
				goto Flag
			case <-srv.quitC:
			case <-srv.stopNewHtCh:
				descStr = "quitC or stopNewHtCh"
				close(stopCh)
				return
			}
		}

		// case 2 waiting
		if time.Now().Before(t0) {
			delay := t0.Sub(time.Now()) //time.Now().Sub(t1)
			//fmt.Println(" ---------case2 update waitEvent", delay, timeOutCount, time.Now())
			select {
			case <-time.After(delay):
				//fmt.Println(" 1111---------case2 update waitEvent", delay, timeOutCount, time.Now())
				continue
			case <-srv.quitC:
				//fmt.Println(" 4444---------case2 update waitEvent", delay, timeOutCount, time.Now())
				return
			case <-srv.stopNewHtCh:
				//fmt.Println(" 5555---------case2 update waitEvent", delay, timeOutCount, time.Now())
				descStr = "quitC or stopNewHtCh"
				return
			}
			//fmt.Println(" 77777---------case2 update waitEvent", delay, timeOutCount, time.Now())
		}

		// case3 timeout
		if time.Now().After(t1) && time.Now().Before(t2) {
			if preNum == srv.dpoaMgr.store.db.GetCurrentBlockHeight() {
				//fmt.Println(" ---------case3 1update waitEvent", preNum, srv.dpoaMgr.store.db.GetCurrentBlockHeight())
				phData := &comm.ViewData{BlkNum: srv.dpoaMgr.partiCfg.BlkNum, View: srv.dpoaMgr.partiCfg.View}
				pb, _ := json.Marshal(phData)
				b, err := srv.cfg.account.Sign(pb)
				if err != nil {
					log.Error("sign procmode433 occured error", "error", err)
					return
				}
				srv.trans.sendMsg(&SendMsgEvent{
					ToPeer: comm.BroadCast,
					Msg:    &comm.ViewtimeoutMsg{RawData: phData, Signature: b, PubKey: srv.cfg.accountStr},
				})
			} else {
				preNum = srv.dpoaMgr.store.db.GetCurrentBlockHeight()
				//fmt.Println(" ---------case3 2update waitEvent", preNum, srv.dpoaMgr.store.db.GetCurrentBlockHeight())
			}
			delay := t2.Sub(time.Now())
			//fmt.Println(" ---------case3 update waitEvent", timeOutCount, delay)

			select {
			//case <-srv.annNewBlock:
			//	fmt.Println("$$$$$$$$$$$$$$srv.annNewBlock", lastBlkTime, srv.GetCurrentBlockNo(), srv.GetCurrentBlockNo(), srv.store.getLatestBlockNumber(), srv.store.db.GetCurrentBlockHeight())
			case <-time.After(delay):
				//fmt.Println("$$$$$$$$$$$$$$time.After", lastBlkTime, srv.dpoaMgr.store.getLatestBlockNumber()+1, srv.dpoaMgr.store.getLatestBlockNumber()+1, srv.dpoaMgr.store.getLatestBlockNumber(), srv.dpoaMgr.store.db.GetCurrentBlockHeight())
				continue
			case <-srv.quitC:
			case <-srv.stopNewHtCh:
				descStr = "quitC or stopNewHtCh"
				return
			}
		}

		if time.Now().After(t2) {
			//fmt.Println(" ---------case4 update waitEvent", t2, time.Now())
			select {
			case <-time.After(comm.V_Three * time.Second):
				//fmt.Println("wait timeout or new block", time.Now(), t2, lastBlk.GetViews()+uint32(timeOutCount)+1)
				continue
			case <-srv.quitC:
			case <-srv.stopNewHtCh:
				descStr = "quitC or stopNewHtCh"
				return
			}
		}
	}
}

func (srv *ProcMode) NewEpoch() {
	var (
		desc        string
		blkNum      uint64 = srv.dpoaMgr.store.getLatestBlockNumber() + 1
		lastBlk, _         = srv.dpoaMgr.store.getSealedBlock(blkNum - 1)
		lastBlkTime        = time.Unix(int64(lastBlk.Block.Header.Timestamp), 0)
		stars              = srv.dpoaMgr.store.GetCurStars()
		ebgHeight          = srv.dpoaMgr.store.EpochBegin()
	)

	defer func() {
		log.Info("Server NewEpoch exit", "blkNum", blkNum, "cause", desc)
		select {
		case <-srv.stopNewHtCh:
		default:
			close(srv.stopNewHtCh)
		}
	}()

	blkData, _ := srv.dpoaMgr.store.getSealedBlock(ebgHeight)
	_, epochViews := CalcStellar(float64(len(stars)))
	endtime := time.Unix(int64(blkData.Block.Header.Timestamp), 0).Add(time.Duration(srv.cfg.starsCfg.duration) * time.Second).Add(2 * time.Duration(srv.cfg.starsCfg.duration) * time.Second * time.Duration(epochViews))
	log.Info("Server NewEpoch waitEvent curblkbum, lastblktime, range", "blkNum", blkNum, "lastBlkTime", lastBlkTime, "time", time.Now().Sub(lastBlkTime).Seconds())
	fmt.Println("Server NewEpoch waitEvent curblkbum, lastblktime, range", "blkNum", blkNum, "lastBlkTime", lastBlkTime, "time", time.Now().Sub(lastBlkTime).Seconds())
	srv.stopNewHtCh = make(chan struct{})

	if srv.trans.stateMgr.getState() != SyncReady {
		log.Error("Server NewEpoch state is not SyncReady or Synced", "err is", srv.trans.stateMgr.getState())
		desc = fmt.Sprintf("Server NewEpoch state is not SyncReady or Synced", "err is", srv.trans.stateMgr.getState())
		return
	}

	log.Info("Server NewEpoch loop latest blktime, epoch begin height, epoch endtime", "lastBlkTime", lastBlkTime, "ebgHeight", ebgHeight, "endtime", endtime)

	if time.Now().After(endtime) {
		log.Info("Server NewEpoch waitEvent blknum, latest blknum, out of range", "blkNum", blkNum, "lastBlkTime", lastBlkTime, "time", time.Now().Sub(lastBlkTime).Seconds())
		desc = fmt.Sprintf("Server NewEpoch waitEvent blknum %v, latest blknum %v, out of range %v (s)", blkNum, lastBlkTime, time.Now().Sub(lastBlkTime).Seconds())
		return
	}

	srv.epochPorcess(endtime, epochViews, lastBlk, blkData)
}
