package dpoa

import (
	"encoding/json"
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/log"
)

func (t *TransAction) ConsenseDone(peerIdx string, msg interface{}, msgHash common.Hash) {
	if t.dpoaMgr.store.isEarth() {
		return
	}
	pMsg, ok := msg.(*comm.ConsedoneMsg)
	if !ok {
		//log.Error("invalid msg with proposal msg type")
		return
	}
	//log.Info("Server run receiveFromPeer ConsenseDone", "peerIdx", peerIdx, "GetBlockNum", pMsg.GetBlockNum(), "getLatestBlockNumber", t.dpoaMgr.store.getLatestBlockNumber()+1)
	if pMsg.BlockNumber == t.GetCurrentBlockNo() { //
		st := t.dpoaMgr.GetpaxState()
		if st.state != PaxStateProcessing {
			//log.Info("Server run receiveFromPeer ConsenseDone stateerr", "peerIdx", peerIdx, "state", st.state)
			return
		}
		if st.patiCfg.View != pMsg.View {
			//log.Info("Server run receiveFromPeer ConsenseDone viewerr", "peerIdx", peerIdx, "patiCfg.View", st.patiCfg.View, "pMsg.View", pMsg.View)
			return
		}


		buf, err := common.Hex2Bytes(pMsg.PublicKey)
		if err != nil {
			return
		}

		publicKey, _ := crypto.UnmarshalPubkey(buf)
		if _, err := publicKey.Verify(pMsg.BlockData, pMsg.SigData); err != nil {
			//log.Info("-------Server run receiveFromPeer ConsenseDone Verify err", "peerIdx", peerIdx, "err", err)
			return
		}

		block := &comm.Block{}
		block.Deserialize(pMsg.BlockData)
		if block != nil && block.Block != nil && block.Block.Header != nil {
			log.Info("func dpoa ConsenseDone yes 1", "blockHeight", block.Block.Header.Height, "txlen", block.Block.Transactions.Len())
		} else {
			log.Info("func dpoa ConsenseDone no block nil")
		}

		/*log.Info("Server run receiveFromPeer ConsenseDone", "peerIdx", peerIdx, "GetBlockNum", pMsg.GetBlockNum(),
		"len(ProcSigs)", len(block.Block.Sigs.ProcSigs), "len(FailerSigs)", len(block.Block.Sigs.FailerSigs), "len(TimeoutSigs)", len(block.Block.Sigs.TimeoutSigs))
		*/

		if t.dpoaMgr.role == PaxRoleProcesser {

			consenseblk := &comm.Block{}
			if t.dpoaMgr.paxosIns.dataPhase2 == nil {
				return
			}
			consenseblk.Deserialize(t.dpoaMgr.paxosIns.dataPhase2.Av)
			//fmt.Println("PaxRoleProcesser=====================!!!===============:", len(consenseblk.Block.Sigs.ProcSigs), len(consenseblk.Block.Sigs.FailerSigs), len(consenseblk.Block.Sigs.TimeoutSigs))
			if consenseblk.Block.Hash() != block.Block.Hash() {
				log.Info("-------Server run receiveFromPeer ConsenseDone blkhash err", "peerIdx", peerIdx, "consense hash", consenseblk.Block.Hash(), "hash", block.Block.Hash())
				return
			}

			//fmt.Println("-------------------ConsenseDone sealProposal")
			log.Info("func dpoa ConsenseDone yes 2", "blockHeight", block.Block.Header.Height, "txlen", block.Block.Transactions.Len())
			t.dpoaMgr.store.sealBlock(block)
		} else if t.dpoaMgr.role == PaxRoleObserver {

		} else if t.dpoaMgr.role == PaxRoleFail {

		}
	}
}

func (t *TransAction) ConsensePrepare(peerIdx string, msg interface{}, msgHash common.Hash) {
	pMsg, ok := msg.(*comm.PrepareMsg)
	if !ok {
		//log.Error("invalid msg with proposal msg type")
		return
	}
	p, _ := json.Marshal(pMsg)
	log.Info("-------Server run receiveFromPeer ConsensePrepare", "peerIdx", peerIdx, "msg", string(p), "CurrentBlock", t.GetCurrentBlockNo())

	msgBlkNum := pMsg.Msg.BlockNumber
	if msgBlkNum > t.GetCurrentBlockNo() {
	} else if msgBlkNum < t.GetCurrentBlockNo() {

		if msgBlkNum <= t.GetCommittedBlockNo() {
			if msgBlkNum+10 < t.GetCommittedBlockNo() {
				log.Info("server get proposal msg for block, from, current committed",
					"accountStr", t.cfg.accountStr, "from", msgBlkNum, "committed", t.GetCommittedBlockNo())
				t.heartbeat()
			}
			return
		}
	} else {
		if err := t.msgpool.AddMsg(pMsg, msgHash); err != nil {
			log.Error("failed to add proposal msg to pool", "msgBlkNum", msgBlkNum)
			return
		}
		b, _ := json.Marshal(pMsg.Msg)

		buf, err := common.Hex2Bytes(peerIdx)
		if err != nil {
			return
		}

		publicKey, _ := crypto.UnmarshalPubkey(buf)
		if fg, err := publicKey.Verify(b, pMsg.Sig); !fg && err != nil {
			log.Info("-------Server run receiveFromPeer ConsenseDone Verify", "peerIdx", peerIdx, "err", err)
			return
		}
		t.dpoaMgr.SendData(pMsg.Msg)
	}
}

func (t *TransAction) ConsensePromise(peerIdx string, msg interface{}, msgHash common.Hash) {
	pMsg, ok := msg.(*comm.PromiseMsg)
	if !ok {
		log.Error("invalid msg with proposal msg type")
		return
	}
	p, _ := json.Marshal(pMsg)
	log.Info("-------Server run receiveFromPeer ConsensePromise", "peerIdx", peerIdx, "msg", string(p))

	msgBlkNum := pMsg.Msg.BlockNumber
	if msgBlkNum > t.GetCurrentBlockNo() {
	} else if msgBlkNum < t.GetCurrentBlockNo() {

		if msgBlkNum <= t.GetCommittedBlockNo() {
			if msgBlkNum+10 < t.GetCommittedBlockNo() {
				log.Info("server %d get proposal msg for block %d, from %d, current committed %d",
					t.cfg.account.PublicKey, msgBlkNum, t.GetCommittedBlockNo())
				t.heartbeat()
			}
			return
		}
	} else {
		if err := t.msgpool.AddMsg(pMsg, msgHash); err != nil {
			log.Error("failed to add proposal msg (%d) to pool", msgBlkNum)
			return
		}
		b, _ := json.Marshal(pMsg.Msg)

		buf, err := common.Hex2Bytes(peerIdx)
		if err != nil {
			return
		}

		publicKey, _ := crypto.UnmarshalPubkey(buf)
		if _, err := publicKey.Verify(b, pMsg.Sig); err != nil {
			log.Info("-------Server run receiveFromPeer ConsenseDone Verify err", "peerIdx", peerIdx, "err", err)
			return
		}
		t.dpoaMgr.SendData(pMsg.Msg)
	}
}

func (t *TransAction) ConsenseProposer(peerIdx string, msg interface{}, msgHash common.Hash) {
	pMsg, ok := msg.(*comm.ProposerMsg)
	if !ok {
		log.Error("invalid msg with proposal msg type")
		return
	}
	p, _ := json.Marshal(pMsg)
	log.Info("-------Server run receiveFromPeer ConsenseProposer", "peerIdx", peerIdx, "msg", string(p))

	msgBlkNum := pMsg.Msg.BlockNumber
	if msgBlkNum > t.GetCurrentBlockNo() {
	} else if msgBlkNum < t.GetCurrentBlockNo() {

		if msgBlkNum <= t.GetCommittedBlockNo() {
			if msgBlkNum+10 < t.GetCommittedBlockNo() {
				log.Info("server %d get proposal msg for block %d, from %d, current committed %d",
					t.cfg.account.PublicKey, msgBlkNum, t.GetCommittedBlockNo())
				t.heartbeat()
			}
			return
		}
	} else {
		if err := t.msgpool.AddMsg(pMsg, msgHash); err != nil {
			log.Error("failed to add proposal msg (%d) to pool", msgBlkNum)
			return
		}
		b, _ := json.Marshal(pMsg.Msg)

		buf, err := common.Hex2Bytes(peerIdx)
		if err != nil {
			return
		}

		publicKey, _ := crypto.UnmarshalPubkey(buf)
		if _, err := publicKey.Verify(b, pMsg.Sig); err != nil {
			log.Info("-------Server run receiveFromPeer ConsenseDone Verify err:%v %v", peerIdx, err)
			return
		}
		t.dpoaMgr.SendData(pMsg.Msg)
	}
}

func (t *TransAction) ConsenseAccept(peerIdx string, msg interface{}, msgHash common.Hash) {
	pMsg, ok := msg.(*comm.AcceptMsg)
	if !ok {
		log.Error("invalid msg with proposal msg type")
		return
	}
	p, _ := json.Marshal(pMsg)
	log.Info("-------Server run receiveFromPeer ConsenseAccept", "peerIdx", peerIdx, "msg", string(p))

	msgBlkNum := pMsg.Msg.BlockNumber
	if msgBlkNum > t.GetCurrentBlockNo() {
	} else if msgBlkNum < t.GetCurrentBlockNo() {

		if msgBlkNum <= t.GetCommittedBlockNo() {
			if msgBlkNum+10 < t.GetCommittedBlockNo() {
				log.Info("server get proposal msg for block from current committed",
					"accountStr", t.cfg.accountStr, "block", msgBlkNum, "committed", t.GetCommittedBlockNo())
				t.heartbeat()
			}
			return
		}
	} else {
		if err := t.msgpool.AddMsg(pMsg, msgHash); err != nil {
			log.Error("failed to add proposal msg to pool", "msgBlkNum", msgBlkNum)
			return
		}
		//publicKey, _ := vconfig.Pubkey(peerIdx)
		b, _ := json.Marshal(pMsg.Msg)
		//if err := signature.Verify(publicKey, b, pMsg.Sig); err != nil{
		//	log.Info("-------Server run receiveFromPeer ConsenseAccept Verify err:%v %v", peerIdx, err)
		//	return
		//}

		buf, err := common.Hex2Bytes(peerIdx)
		if err != nil {
			return
		}

		publicKey, _ := crypto.UnmarshalPubkey(buf)
		if _, err := publicKey.Verify(b, pMsg.Sig); err != nil {
			log.Info("-------Server run receiveFromPeer ConsenseDone Verify err", "peerIdx", peerIdx, "err", err)
			return
		}
		t.dpoaMgr.SendData(pMsg.Msg)
	}
}

func (t *TransAction) ViewTimeout(peerIdx string, msg interface{}, msgHash common.Hash) {
	pMsg, ok := msg.(*comm.ViewtimeoutMsg)
	if !ok {
		log.Error("invalid msg with heartbeat msg type")
		return
	}
	log.Info("-------Server run receiveFromPeer ViewTimeout", "peerIdx", peerIdx, "GetCurrentBlockNo", t.GetCurrentBlockNo(), "GetBlockNum", pMsg.GetBlockNum(), "View", pMsg.RawData.View)
	if t.GetCurrentBlockNo() != pMsg.GetBlockNum() { // 丢弃
		return
	}

	err := CheckSigners(pMsg, peerIdx)
	if err != nil {
		log.Info("-------Server run receiveFromPeer ViewTimeout CheckSigners err", "peerIdx", peerIdx, "err", err)
		return
	}
	st := t.dpoaMgr.GetpaxState()
	log.Info("~~~~~~~~~~~~~~~~~~~~~~-------Server run receiveFromPeer ViewTimeout GetpaxState", "peerIdx", peerIdx, "state", st.state, "time", time.Now().String())
	if st.state != PaxStateTimeout || st.patiCfg.BlkNum != pMsg.GetBlockNum() || st.patiCfg.View != pMsg.RawData.View {
		log.Info("-------Server run receiveFromPeer ViewTimeout GetpaxState", "peerIdx", peerIdx, "state", st.state, "patiCfg", st.patiCfg)
		time.Sleep(2 * time.Second) // 容错
		st = t.dpoaMgr.GetpaxState()
		if st.state != PaxStateTimeout || st.patiCfg.BlkNum != pMsg.GetBlockNum() || st.patiCfg.View != pMsg.RawData.View {
			return
		}
	}
	isfound := false
	for _, pubkey := range t.dpoaMgr.partiCfg.ObserNodes {
		if pubkey == peerIdx {
			isfound = true
			break
		}
	}
	if !isfound {
		log.Info("-------Server run receiveFromPeer ViewTimeout isnotfound err", peerIdx)
		return
	}
	log.Info("-------Server run receiveFromPeer ViewTimeout add", "peerIdx", peerIdx)
	t.msgpool.AddMsg(pMsg, msgHash, t.dpoaMgr.partiCfg)
}

func (t *TransAction) EarthFetchRspSigs(peerIdx string, msg interface{}, msgHash common.Hash) {
	if !t.dpoaMgr.store.isEarth() {
		return
	}
	pMsg, ok := msg.(*comm.EarthSigsFetchRspMsg)
	if !ok {
		log.Error("invalid msg with heartbeat msg type")
		return
	}
	p, _ := json.Marshal(pMsg)
	log.Info("-------Server run receiveFromPeer", "peerIdx", peerIdx, "EarthFetchRspSigs", string(p))
	if t.dpoaMgr.store.inFailers(t.dpoaMgr.store.EpochBegin(), pMsg.PubKey) {
		t.msgpool.AddMsg(pMsg, msgHash)
	}
}

func (t *TransAction) PeerHeartbeatMessage(peerIdx string, msg interface{}, msgHash common.Hash) {
	pMsg, ok := msg.(*comm.PeerHeartbeatMsg)
	if !ok {
		log.Error("invalid msg with heartbeat msg type")
		return
	}
	if err := t.processHeartbeatMsg(peerIdx, pMsg); err != nil {
		log.Error("server failed to process heartbeat", "accountStr", t.cfg.accountStr, "peerIdx", peerIdx, "err", err)
	}
	if pMsg.CommittedBlockNumber+10 < t.GetCommittedBlockNo() {
		// delayed peer detected, response heartbeat with our chain Info
		t.heartbeat()
	}
}

func (t *TransAction) BlockFetchMessage(peerIdx string, msg interface{}, msgHash common.Hash) {
	pMsg, ok := msg.(*comm.BlockFetchMsg)
	if !ok {
		log.Error("invalid msg with blockfetch msg type")
		return
	}
	p, _ := json.Marshal(pMsg)
	log.Info("-------Server run receiveFromPeer BlockFetchMessage", "peerIdx", peerIdx, "BlockFetchMessage", string(p))
	blk, err := t.dpoaMgr.store.getSealedBlock(pMsg.BlockNum)
	if err != nil {
		log.Error("server failed to handle blockfetch from",
			"accountStr", t.cfg.accountStr, "BlockNum", pMsg.BlockNum, "peerIdx", peerIdx, "err", err)
		return
	}
	msg1, err := constructBlockFetchRespMsg(pMsg.BlockNum, blk, blk.Block.Hash())
	if err != nil {
		log.Error("server %v, failed to handle blockfetch %d from %v: %s",
			"accountStr", t.cfg.accountStr, "BlockNum", pMsg.BlockNum, "peerIdx", peerIdx, "err", err)
	} else {
		//log.Infof("server %d, handle blockfetch %d from %d",
		//	self.Index, pMsg.BlockNum, peerIdx)
		t.msgSendC <- &SendMsgEvent{
			ToPeer: peerIdx,
			Msg:    msg1,
		}
	}
}

func (t *TransAction) BlockInfoFetchMessage(peerIdx string, msg interface{}, msgHash common.Hash) {
	// handle block Info fetch msg
	pMsg, ok := msg.(*comm.BlockInfoFetchMsg)
	if !ok {
		log.Error("invalid msg with blockinfo fetch msg type")
		return
	}
	p, _ := json.Marshal(pMsg)
	log.Info("-------Server run receiveFromPeer BlockInfoFetchMessage:%v %v", peerIdx, string(p))

	maxCnt := 64
	blkInfos := make([]*comm.BlockInfo_, 0)
	targetBlkNum := t.GetCommittedBlockNo()
	for startBlkNum := pMsg.StartBlockNum; startBlkNum <= targetBlkNum; startBlkNum++ {
		blk, _ := t.dpoaMgr.store.getSealedBlock(startBlkNum)
		if blk == nil {
			break
		}
		blkInfos = append(blkInfos, &comm.BlockInfo_{
			BlockNum: startBlkNum,
			//Proposer: blk.getProposer(),
		})
		if len(blkInfos) >= maxCnt {
			break
		}
	}
	msg1, err := constructBlockInfoFetchRespMsg(blkInfos)
	if err != nil {
		//log.Errorf("server %d, failed to handle blockinfo fetch %d to %d: %s",
		//	self.Index, pMsg.StartBlockNum, peerIdx, err)
	} else {
		//log.Infof("server %d, response blockinfo fetch to %d, blk %d, len %d",
		//	self.Index, peerIdx, pMsg.StartBlockNum, len(blkInfos))
		t.msgSendC <- &SendMsgEvent{
			ToPeer: peerIdx,
			Msg:    msg1,
		}
	}
}
