package validation

import (
	"math/big"
	"sync"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

var (
	ratU = big.NewRat(1e8, 1)
	bigU = big.NewInt(1e8)
)

// todo how to do with the nonce
type StateValidate struct {
	TargetHeight       uint64
	currentBlockHeight uint64
	currentBlockHash   common.Hash
	MemoAccountState   map[common.Address]storelaw.AccountStater
	//for one tx ,Take the lower limit，only the subtraction
	DirtyAccountState map[common.Address]storelaw.AccountStater

	//vote instant state
	MemoVoteState map[common.Hash]*storelaw.VoteState

	DirtyVote map[common.Hash]map[common.Address]struct{}
	Blacklist map[common.Address]bool

	receipts          types.Receipts
	txs               types.Transactions
	ParentBonusHeight uint64
	WaitOplogs        []*OpLog //Transactional submission

	//the ut amount when the vote start
	voteUT map[common.Hash]*TxUT
	//new vote pool
	voteNew       []*storelaw.VoteState
	accountVoteds map[common.Hash]map[common.Address]struct{}
	ledgeror      storelaw.Ledger4Validation
	leagueId      common.Address

	m sync.Mutex
}

func NewStateValidate(ledgeror storelaw.Ledger4Validation, leadgeId common.Address) *StateValidate {

	curHeight := ledgeror.GetCurrentBlockHeight4V(leadgeId)
	s := &StateValidate{
		TargetHeight:       curHeight + 1,
		currentBlockHash:   ledgeror.GetCurrentHeaderHash4V(leadgeId),
		currentBlockHeight: curHeight,
		MemoAccountState:   make(map[common.Address]storelaw.AccountStater),
		DirtyAccountState:  make(map[common.Address]storelaw.AccountStater),
		Blacklist:          make(map[common.Address]bool),
		MemoVoteState:      make(map[common.Hash]*storelaw.VoteState),
		voteUT:             make(map[common.Hash]*TxUT),
		accountVoteds:      make(map[common.Hash]map[common.Address]struct{}),
		DirtyVote:          make(map[common.Hash]map[common.Address]struct{}),
		WaitOplogs:         []*OpLog{}, //Transactional submission
		ledgeror:           ledgeror,
		leagueId:           leadgeId,
	}
	return s
}

func (s *StateValidate) TxInheritance() <-chan *types.Transaction {
	ch := make(chan *types.Transaction, 20)
	for k, _ := range s.txs {
		if !s.ledgeror.ContainTx4V(s.txs[k].Hash()) {
			ch <- s.txs[k]
		}
	}
	close(ch)
	return ch
}

//ExecuteOplogs is perform the verified transaction calculation and generate the status data
func (s *StateValidate) ExecuteOplogs() *storelaw.OrgBlockInfo {
	s.m.Lock()
	defer s.m.Unlock()
	{
		//fmt.Println("⛏ ⛏  ExecuteOplogs txlen", s.txs.Len())
		var (
			aa *OPAddressBigInt
			an *OPAddressUint64
			ha *OPHashAddress
			//ad *OPAddressAdress
			fee = big.NewInt(0)
		)
		ut := s.ledgeror.GetUTByHeight(s.currentBlockHeight, s.leagueId)
		//the vote hasn't passed
		afterVote := make(map[common.Hash]struct{})
		for k, v := range s.MemoVoteState {
			//fmt.Println("⭕️ ExecuteOplogs vote hash ", k.String(), len(s.MemoVoteState))
			if !v.GetResult(s.voteUT[k].UT, 67) {
				//fmt.Println("⭕️ ExecuteOplogs vote result ", s.voteUT[k].UT)
				afterVote[k] = struct{}{}
			}
		}
		for _, v := range s.WaitOplogs {
			switch v.method {
			case Account_ut_add:
				aa = v.data.(*OPAddressBigInt)
				s.MemoAccountState[aa.Address].BalanceAdd(aa.Num)
				break
			case Account_energy_add:
				aa = v.data.(*OPAddressBigInt)
				//fmt.Println("func ExecuteOplogs address:", aa.Address.ToString())
				s.MemoAccountState[aa.Address].EnergyAdd(aa.Num)
				break
			case Account_ut_sub:
				aa = v.data.(*OPAddressBigInt)
				s.MemoAccountState[aa.Address].BalanceSub(aa.Num)
				break
			case Account_energy_sub, Account_energy_consume:
				aa = v.data.(*OPAddressBigInt)
				s.MemoAccountState[aa.Address].EnergySub(aa.Num)
				if v.method == Account_energy_consume {
					fee.Add(fee, aa.Num)
				}
				break
			case Account_bonus_left:
				an = v.data.(*OPAddressUint64)
				s.MemoAccountState[an.Address].BonusHSet(an.Num)
				break
			case Account_nonce_add:
				an = v.data.(*OPAddressUint64)
				s.MemoAccountState[an.Address].NonceSet(an.Num)
				break
			case Vote_abstention:
				ha = v.data.(*OPHashAddress)
				h := s.MemoVoteState[ha.Hash].Height
				as := s.GetPrevAccountState(ha.Address, h)
				s.MemoVoteState[ha.Hash].AddAbstention(as.Balance())
			case Vote_against:
				ha = v.data.(*OPHashAddress)
				h := s.MemoVoteState[ha.Hash].Height
				as := s.GetPrevAccountState(ha.Address, h)
				s.MemoVoteState[ha.Hash].AddAgainst(as.Balance())
			case Vote_agree:
				ha = v.data.(*OPHashAddress)
				h := s.MemoVoteState[ha.Hash].Height
				as := s.GetPrevAccountState(ha.Address, h)
				s.MemoVoteState[ha.Hash].AddAgree(as.Balance())
			case League_Raise_ut:
				aa = v.data.(*OPAddressBigInt)
				ut.Add(ut, aa.Num)
			}
		}
		accountStates := storelaw.AccountStaters{}
		for _, v := range s.MemoAccountState {
			accountStates = append(accountStates, v)
		}

		//todo calculate memberRoot and fill it into memoLeagueState
		block := &types.Block{
			Transactions: s.txs,
			Header: &types.Header{
				Height:        s.TargetHeight,
				PrevBlockHash: s.currentBlockHash,
				Version:       0x01,
				TxRoot:        s.txs.GetHashRoot(),
				StateRoot:     accountStates.GetHashRoot(),
				ReceiptsRoot:  s.receipts.GetHashRoot(),
				LeagueId:      s.leagueId,
			},
		}

		//fmt.Println("♌️ ♌️ ♌️ ♌️ ♌️ ut️:", ut, " leagueId:", s.leagueId.ToString())
		if s.TargetHeight%types.HWidth == 0 {
			energyUsed := s.ledgeror.GetHeaderBonus(s.leagueId)
			//fmt.Printf("♐️ ♐️ league ExecuteOplogs destroyEnergy:%d ut:%d\n ", energyUsed.Uint64(), ut.Uint64())
			rat1 := new(big.Rat).SetInt(ut)
			rat2 := new(big.Rat).SetInt(energyUsed)
			rat1.Inv(rat1)
			rat1.Mul(rat1, rat2)
			rat1.Mul(rat1, ratU)
			rfTmp := big.NewFloat(0).SetRat(rat1)
			integer, _ := rfTmp.Uint64()
			block.Header.Bonus = integer
			//fmt.Println("♐️ ♐️ league ExecuteOplogs bonus ", integer)

		}
		feeSum := big.NewInt(0)
		for _, v := range s.receipts {
			feeSum.Add(feeSum, big.NewInt(0).SetUint64(v.GasUsed))
		}
		blkInfo := &storelaw.OrgBlockInfo{
			Block:         block,
			Receipts:      s.receipts,
			AccStates:     accountStates,
			FeeSum:        feeSum,
			VoteStates:    make([]*storelaw.VoteState, 0, len(s.voteNew)+len(s.MemoVoteState)),
			AccountVoteds: make([]*storelaw.AccountVoted, 0, len(s.DirtyVote)),
		}
		//put new vote into votestates
		blkInfo.VoteStates = append(blkInfo.VoteStates, s.voteNew...)
		if len(s.MemoVoteState) > 0 {
			for k, v := range s.MemoVoteState {
				blkInfo.VoteStates = append(blkInfo.VoteStates, v)

				if v.GetResult(s.voteUT[k].UT, 67) {
					if _, ok := afterVote[k]; ok {
						fmt.Println("⭕️ vote first pass ", s.voteUT[k].Tx.Hash())
						blkInfo.VoteFirstPass = append(blkInfo.VoteFirstPass, s.voteUT[k].Tx)
					}
				}
			}
		}
		for k, v := range s.DirtyVote {
			av := &storelaw.AccountVoted{
				VoteId:   k,
				Height:   s.TargetHeight,
				Accounts: make([]common.Address, 0, len(v)),
			}
			for ki, _ := range v {
				av.Accounts = append(av.Accounts, ki)
			}
			blkInfo.AccountVoteds = append(blkInfo.AccountVoteds, av)
		}
		//UT TODO if the tx causes a change in the ut amount,it needs to be calculated on ut.
		blkInfo.UT = ut
		return blkInfo
	}
}
func (s *StateValidate) GetTxLen() int {
	return s.txs.Len()
}

//VerifyTx is verify that the transaction is valid and return the processing plan
//returns:
// 	int   ------   -1: throw tx,0: back to queue, 1: efficient transaction
//  error ------   Existing warning
// if verify pass,tx is going to change StateValidate，append WaitOplogs,to affect dirtyState
func (s *StateValidate) VerifyTx(tx *types.Transaction) (int, error) {
	s.m.Lock()
	defer s.m.Unlock()
	//fmt.Println("🔍 verifytx start")
	err := s.CheckSign(tx)
	if err != nil {
		return -1, err
	}
	err = s.CheckFee(tx)
	if err != nil {
		return -1, err

	}
	oplogs := HandleTransaction(tx)
	//route oplogs to its homestate
	accountOps := make([]*OpLog, 0, len(oplogs))
	bonusOps := make([]*OpLog, 0, 3)
	var (
		voteOp  *OpLog
		raiseOp *OpLog
	)
	//feeOp := &OpLog{}
	//check exits in dirtystate,if not then get
	for k, v := range oplogs {
		if v.method < AccountMaxLine {
			accountOps = append(accountOps, oplogs[k])
			/*if v.method == Account_energy_consume {
				feeOp = oplogs[k]
			}*/
		} else if v.method == Account_bonus_fee {
			bonusOps = append(bonusOps, v)
		} else if v.method == League_Raise_ut {
			raiseOp = v
		} else {
			voteOp = v
		}
	}
	fee := uint64(0)
	if tx.TxData.Fee != nil {
		fee = tx.TxData.Fee.Uint64()
	}
	receipt := &types.Receipt{
		TxHash:  tx.Hash(),
		GasUsed: fee,
	}
	//fill oplogs into WaitOplogs and FeeOplogs (group by)
	opIndes1, err := s.VerifyAccount(accountOps)
	if err != nil {
		s.RollbackAccount(accountOps, opIndes1)
		if err == errors.ERR_STATE_ACCOUNT_NONCE_DISCONTINUITY {
			return 0, err
		} else {
			return -1, err
		}
		/*err = s.verifyAccount(feeOp)
		if err != nil {
			return 0, err
		}
		s.appendWaitOplog(feeOp)
		s.appendTx(tx)
		s.appendReceipt(receipt)
		return 1, nil*/
	}
	if len(bonusOps) > 0 {
		ops, err := s.verifyBonus(bonusOps[0])
		if err != nil {
			s.RollbackAccount(accountOps, opIndes1)
			return -1, err
		}
		bonusOps = append(bonusOps, ops...)
	}
	if voteOp != nil {
		err = s.verifyVote(voteOp)
		if err != nil {
			s.RollbackAccount(accountOps, opIndes1)
			return -1, err
		}
	}

	receipt.Status = true
	s.appendReceipt(receipt)
	s.appendWaitOplogs(accountOps)
	s.appendWaitOplog(voteOp)
	s.appendWaitOplogs(bonusOps)
	s.appendWaitOplog(raiseOp)
	s.appendTx(tx)
	return 1, nil
}

func (s *StateValidate) appendWaitOplogs(oplogs []*OpLog) {
	s.WaitOplogs = append(s.WaitOplogs, oplogs...)
}
func (s *StateValidate) appendWaitOplog(oplog *OpLog) {
	if oplog != nil {
		s.WaitOplogs = append(s.WaitOplogs, oplog)
	}
}
func (s *StateValidate) appendTx(tx *types.Transaction) {
	s.txs = append(s.txs, tx)
	s.appendVoteNew(tx)
}
func (s *StateValidate) appendVoteNew(tx *types.Transaction) {
	if voteTypeContains(tx.TxType) {
		vs := &storelaw.VoteState{
			VoteId:     tx.Hash(),
			Height:     s.TargetHeight,
			Agree:      big.NewInt(0),
			Against:    big.NewInt(0),
			Abstention: big.NewInt(0),
		}
		s.voteNew = append(s.voteNew, vs)
	}
}

func (s *StateValidate) appendReceipt(receipt *types.Receipt) {
	s.receipts = append(s.receipts, receipt)
}

//
func (s *StateValidate) CheckSign(tx *types.Transaction) error {
	return nil
}

//todo
func (s *StateValidate) CheckFee(tx *types.Transaction) error {
	return nil
}
func (s *StateValidate) CheckNodeId(nodeId common.Address) error {
	return nil
}

//TODO get data from leder_store
func (s *StateValidate) checkMemberState(address, leagueId common.Address) error {
	return nil
}
