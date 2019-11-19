package dpoa

import (
	"encoding/json"
	"math/big"
	"time"

	//"github.com/sixexorg/magnetic-ring/core/types"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/consense/dpoa/comm"

	//"github.com/sixexorg/magnetic-ring/core/signature"
	//"github.com/sixexorg/magnetic-ring/consensus/vbft/config"
	"github.com/sixexorg/magnetic-ring/log"

	comm2 "github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto"
)

// Paxos instance
type Paxos struct {
	index          int
	totalGas       int
	publicKey      string
	active         bool        // active leader
	ballot         comm.Ballot // highest ballot number
	ab             comm.Ballot // highest ballot number
	av             []byte
	quorumPhase1   *comm.Quorum // phase 1 quorum
	quorumPhase2   *comm.Quorum // phase 2 quorum
	candidateBlock *comm.Block
	dataPhase1     *comm.P1a //
	dataPhase2     *comm.P2a
	done           bool
	account        account.Account
	sendCh         chan interface{}
	partiCfg       *comm.ParticiConfig
	sigsMap        map[uint16][]byte
}

func NewPaxos(publicKey string, acc account.Account, sendCh chan interface{}) *Paxos {
	return &Paxos{
		sendCh:    sendCh,
		publicKey: publicKey,
		account:   acc,
	}
}

func (p *Paxos) ConsensProcess(block *comm.Block, totalGas int, partiCfg *comm.ParticiConfig, stopCh chan struct{}) {

	log.Info("func dpoa paxos ConsensProcess", "blockHeight", block.Block.Header.Height, "txlen", block.Block.Transactions.Len())

	//fmt.Println("============>>>>>>>>>ConsensProcess", block.GetProposer(), block.GetBlockNum(), block.GetViews(), block.Block.Sigs == nil)
	p.partiCfg = partiCfg
	p.candidateBlock = block
	p.totalGas = totalGas

	{
		mm, _ := p.candidateBlock.Serialize()
		bk := &comm.Block{}
		bk.Deserialize(mm)

		//log.Info("ConsensProcess print info bk", "block.txs.len", bk.Block.Transactions.Len())

		//fmt.Println("11111111111============>>>>>>>>>ConsensProcess", bk.GetProposer(), bk.GetBlockNum(), bk.GetViews(), bk.Block.Sigs == nil)
	}

	p.ballot = comm.Ballot(0)
	p.quorumPhase1 = comm.NewQuorum(len(p.partiCfg.ProcNodes))
	p.quorumPhase2 = comm.NewQuorum(len(p.partiCfg.ProcNodes))
	p.index, _ = GetIndex(p.partiCfg.PartiRaw, p.publicKey)
	p.sigsMap = make(map[uint16][]byte)
	p.P1a()

	for {
		select {
		case <-time.After(time.Second * 2):
			if p.done {
				return
			}

			p.P1a()
			log.Info("func dpoa paxos ConsensProcess")
		case <-stopCh:
			p.clean()
			return
		}
	}
}

// P1a starts phase 1 prepare
func (p *Paxos) P1a() {
	p.active = false
	p.ballot.Next(comm.NewID(p.totalGas, p.index))
	p.quorumPhase1.Reset()
	p.quorumPhase2.Reset()
	p.quorumPhase1.ACK(comm.NewID(0, p.index)) //
	//fmt.Println("=======================>>>>>>>>>>>>>>>>>>>>>>>>>>P1a", p.index, p.ballot.String(), p.candidateBlock.Block.Hash())
	p1a := comm.P1a{Ballot: p.ballot, ID: uint32(p.index), BlockNumber: p.partiCfg.BlkNum, View: p.partiCfg.View, TotalGas: big.NewInt(int64(p.totalGas))}
	bp1a, _ := json.Marshal(p1a)
	sig, err := p.account.Sign(bp1a)

	if err != nil {
		log.Error("sign paxos 105 occured error", "error", err)
		return
	}
	log.Info("func dpoa paxos P1a sendCh msg")
	p.sendCh <- &SendMsgEvent{ToPeer: comm.BroadCast, Msg: &comm.PrepareMsg{Msg: p1a, Sig: sig}}
}

func (p *Paxos) p1aCheck(m comm.P1a) error {


	return nil
}

// HandleP1a handles P1a message
func (p *Paxos) HandleP1a(m comm.P1a) {
	//fmt.Printf("success--------------->>>>>>>>>>>>>>>>>>>>>>>>>>HandleP1a\n", m)
	/*
	Serial number check
	Verification signature
	Height check
	View check
	Round check
	Txhash check
	*/
	if p.done {
		if m.Ballot > p.ballot { // accept
			if p.p1aCheck(m) != nil {
				return
			}

			m := comm.P1b{
				Ballot:      m.Ballot,
				ID:          uint32(p.index),
				Ab:          p.ab,
				Av:          p.av,
				BlockNumber: p.partiCfg.BlkNum,
				View:        p.partiCfg.View,
			}
			b, _ := json.Marshal(&m)
			sig, err := p.account.Sign(b)

			if err != nil {
				log.Error("sign paxos 160 occured error", "error", err)
				return
			}

			p.sendCh <- &SendMsgEvent{ToPeer: comm.BroadCast, Msg: &comm.PromiseMsg{Msg: m, Sig: sig}}
		}
		return
	}

	if p.p1aCheck(m) != nil {
		return
	}

	if m.Ballot > p.ballot { // accept
		p.ballot = m.Ballot
		p.active = false
		if len(p.av) == 0 {
			p.dataPhase1 = &m
		}
		m := comm.P1b{
			Ballot:      p.ballot,
			ID:          uint32(p.index),
			Ab:          p.ab,
			Av:          p.av,
			BlockNumber: p.partiCfg.BlkNum,
			View:        p.partiCfg.View,
		}
		b, _ := json.Marshal(&m)
		sig, err := p.account.Sign(b)
		if err != nil {
			log.Error("sign pasos190 occured error", "error", err)
			return
		}
		p.sendCh <- &SendMsgEvent{ToPeer: comm.BroadCast, Msg: &comm.PromiseMsg{Msg: m, Sig: sig}}
	}
	return
}

// HandleP1b handles P1b message
func (p *Paxos) HandleP1b(m comm.P1b) {
	//fmt.Println("==================>>>>>>>>>>>>>>>>>>>>>>>>>>HandleP1b", m.ID, m.Ballot.String(), p.ballot.String(), p.index, p.ab, p.done)
	//fmt.Println("==================>>>>>>>>>>>>>>>>>>>>>>>>>>HandleP1bm", m)
	if p.done || m.Ballot < p.ballot {
		return
	}

	if m.Ab > p.ab && len(m.Av) > 0 {
		p.ab = m.Ab
		p.av = m.Av
	}

	if m.Ballot > p.ballot {
		p.ballot = m.Ballot
		p.active = false // not necessary
	}

	if m.Ballot.ID().Node() == p.index && m.Ballot == p.ballot {
		p.quorumPhase1.ACK(comm.NewID(0, int(m.ID)))
		if p.quorumPhase1.Q1() {
			p.active = true
			//p.quorump2.Reset()
			p.quorumPhase2.ACK(comm.NewID(0, p.index))
			//fmt.Printf("success---------------===>>>>>>>>>>>>>>>>>>>>>>>>>>HandleP1b lenth %v\n", len(p.av))
			if len(p.av) == 0 {
				p.ab = p.ballot
				p.av, _ = p.candidateBlock.Serialize()

				tmpblock := new(comm.Block)

				err := tmpblock.Deserialize(p.av)
				if err != nil {
					log.Error("tmpblock Deserialize occured err", "error", err)
				} else {
					log.Debug("block", "height", tmpblock.Block.Header.Height, "block", tmpblock, "block.txlen", tmpblock.Block.Transactions.Len())
				}

				log.Info("func dpoa HandleP1b", "txlen", p.candidateBlock.Block.Transactions.Len())
			}
			m := comm.P2a{
				Ballot:      p.ballot,
				Av:          p.av,
				BlockNumber: p.partiCfg.BlkNum,
				View:        p.partiCfg.View,
				ID:          uint32(p.index),
			}

			b, _ := json.Marshal(&m)
			{ //test by rennbon
				m2 := &comm.P2a{}
				err := json.Unmarshal(b, m2)
				if err != nil {
					bl := &comm.Block{}
					err = bl.Deserialize(m2.Av)
					if err != nil {
						log.Info("func dpoa paxos HandleP1b", "blockHeight", bl.Block.Header.Height, "txlen", bl.Block.Transactions.Len())
					} else {
						log.Info("func dpoa paxos HandleP1b", "error", err)
					}

				}
			}
			sig, err := p.account.Sign(b)
			if err != nil {
				log.Error("sign paxos 264 occured error", "error", err)
				return
			}
			p.sendCh <- &SendMsgEvent{ToPeer: comm.BroadCast, Msg: &comm.ProposerMsg{Msg: m, Sig: sig}}
		}
	}
}

func (p *Paxos) p2aCheck(m comm.P2a) error {

	return nil
}

func (p *Paxos) HandleP2a(m comm.P2a) {
	//fmt.Printf("===================->>>>>>>>>>>>>>>>>HandleP2a\n", p.index, p.done, m.Ballot == p.ballot, m.Ballot.ID().Node() == p.index)
	if p.done {
		if m.Ballot > p.ballot { // accept
			if p.p2aCheck(m) != nil {
				return
			}

			m1 := comm.P2b{
				Ballot:      m.Ballot,
				ID:          uint32(p.index),
				BlockNumber: p.partiCfg.BlkNum,
				View:        p.partiCfg.View,
			}
			b, _ := json.Marshal(&m1)
			sig, err := p.account.Sign(b)
			if err != nil {
				log.Error("sign paxos 304 occured error", "error", err)
				return
			}
			p.sendCh <- &SendMsgEvent{ToPeer: comm.BroadCast, Msg: &comm.AcceptMsg{Msg: m1, Sig: sig}}
		}
		return
	}

	if p.p2aCheck(m) != nil {
		return
	}

	if m.Ballot == p.ballot {
		//p.dataPhase1.TxHashs
		blk := &comm.Block{}
		err := blk.Deserialize(m.Av)
		//blk.Block.Transactions, p.dataPhase1.TxHashs

		if err != nil {
			log.Error("handle p2a tmpblock Deserialize occured err", "error", err)
		} else {
			log.Debug("handle p2a block", "", blk.Block.Header.Height, "block", blk, "block.txlen", blk.Block.Transactions.Len())
		}

		p.dataPhase2 = &m
		p.ab = m.Ballot
		p.av = m.Av
		h := blk.Block.Hash()
		bsig, err := p.account.Sign(h[:])
		if err != nil {
			log.Error("sign paxos 335 occured error", "error", err)
			return
		}
		m1 := comm.P2b{
			Ballot:      p.ballot,
			ID:          uint32(p.index),
			BlockNumber: p.partiCfg.BlkNum,
			View:        p.partiCfg.View,
			Signature:   bsig,
		}
		b, _ := json.Marshal(&m1)
		sig, err := p.account.Sign(b)
		if err != nil {
			log.Error("sign paxos348 occured error", "error", err)
			return
		}
		//fmt.Printf("===>>>>>>>>>>>>>>>>>HandleP2a\n", m.Ballot, p.ballot, bsig)
		p.sendCh <- &SendMsgEvent{ToPeer: comm.BroadCast, Msg: &comm.AcceptMsg{Msg: m1, Sig: sig}}
	}
}

// HandleP2b handles P2b message
func (p *Paxos) HandleP2b(m comm.P2b) {
	//fmt.Printf("success--------------->>>>>>>>>>>>>>>>>>>>>>>>>>HandleP2b\n", p.index, p.done, m.Ballot == p.ballot, m.Ballot.ID().Node() == p.index, p.av == nil)
	if p.done {
		return
	}
	log.Info("func dpoa HandleP2b 01")
	if m.Ballot > p.ballot {
		p.ballot = m.Ballot
		p.active = false
	}
	log.Info("func dpoa HandleP2b 02")
	if m.Ballot.ID().Node() == p.index && m.Ballot == p.ballot { // id 和序号校验
		p.quorumPhase2.ACK(comm.NewID(0, int(m.ID)))
		p.sigsMap[uint16(m.ID)] = m.Signature
		if p.quorumPhase2.Q2() {
			p.done = true
			blk := &comm.Block{}
			blk.Deserialize(p.av)
			log.Info("func dpoa HandleP2b 03", "blockHeight", blk.Block.Header.Height, "txlen", blk.Block.Transactions.Len())
			if len(blk.Block.Sigs.ProcSigs) == 0 {
				for idx, sig := range p.sigsMap {
					//fmt.Println(">>$$$$$$$$$$$$$$$$$idx", idx)
					var d []byte
					//dd := int2Byte(uint16(idx))
					d = append(d, int2Byte(uint16(idx))...)
					//fmt.Println("1>>$$$$$$$$$$$$$$$$$idx", len(d), byte2Int(d[0:2]), byte2Int(dd))
					d = append(d, sig...)
					blk.Block.Sigs.ProcSigs = append(blk.Block.Sigs.ProcSigs, d)
					///////////
					//fmt.Println("2>>$$$$$$$$$$$$$$$$$idx", byte2Int(d[0:2]))
					pubStr, _ := GetNode(p.partiCfg.PartiRaw, int(byte2Int(d[0:2])))

					buf, err := comm2.Hex2Bytes(pubStr)
					if err != nil {
						return
					}

					publicKey, _ := crypto.UnmarshalPubkey(buf)
					//pubKey, _ := vconfig.Pubkey(pubStr)
					h := blk.Block.Hash()
					if fg, err := publicKey.Verify(h[:], d[2:]); !fg && err != nil {
						//fmt.Errorf("1111111111111111$$$$$$$$$$$$$$$$$idx:%v err:%v", pubStr, err)
					} else {
						//fmt.Println(">>$$$$$$$$$$$$$$$$$idx------->done", d)
					}
					//if err := signature.Verify(pubKey, h[:], d[2:]); err != nil{
					//	fmt.Errorf("1111111111111111$$$$$$$$$$$$$$$$$idx:%v err:%v", pubStr, err)
					//	//return err
					//}else {
					//	fmt.Println(">>$$$$$$$$$$$$$$$$$idx------->done", d)
					//}
				}
				var d []byte
				h := blk.Block.Hash()
				bsig, err := p.account.Sign(h[:])
				if err != nil {
					log.Error("sign paxos408 block.hash() occured error", "error", err)
					return
				}
				d = append(d, int2Byte(uint16(p.index))...)
				d = append(d, bsig...)
				blk.Block.Sigs.ProcSigs = append(blk.Block.Sigs.ProcSigs, d)
			}
			//fmt.Println("success---------------!!!!HandleP2b", blk.GetBlockNum(), len(blk.Block.Sigs.TimeoutSigs), len(blk.Block.Sigs.ProcSigs), len(blk.Block.Sigs.FailerSigs), len(blk.Block.Transactions))
			log.Info("func dpoa HandleP2b 04", "blockHeight", blk.GetBlockNum(), "timeout", len(blk.Block.Sigs.TimeoutSigs), "proc", len(blk.Block.Sigs.ProcSigs), "failer", len(blk.Block.Sigs.FailerSigs), "txlen", len(blk.Block.Transactions))
			p.sendCh <- blk
			blockData, _ := blk.Serialize()

			sig, err := p.account.Sign(blockData)
			if err != nil {
				log.Error("sign paxos422 blockData occured error", "error", err)
				return
			}
			p.sendCh <- &SendMsgEvent{ToPeer: comm.BroadCast, Msg: &comm.ConsedoneMsg{BlockNumber: p.partiCfg.BlkNum, View: p.partiCfg.View, BlockData: blockData, PublicKey: p.publicKey, SigData: sig}}
			//fmt.Printf("success HandleP2b||||||||||p.ID():%v, p.ballot:%v, p.ab:%v,time:%v\n", p.index, p.ballot, p.ab, time.Now().String())
		}
	}
}

func (p *Paxos) clean() {
	time.Sleep(time.Millisecond * 500)
	//fmt.Println("------------------------->>>>>>>>>>>>>>>>>>>>>>>>>>>clean 22222222")
	p.done = false
	p.active = false
	p.candidateBlock = nil
	p.partiCfg = nil
	p.ballot = 0
	p.dataPhase1 = nil
	p.dataPhase2 = nil
	p.totalGas = 0
	p.ab = 0
	p.av = nil
	p.sigsMap = make(map[uint16][]byte)
	if p.quorumPhase1 != nil {
		p.quorumPhase1.Reset()
	}
	if p.quorumPhase2 != nil {
		p.quorumPhase2.Reset()
	}
}
