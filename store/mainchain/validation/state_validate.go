package validation

import (
	"sync"

	"math/big"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

var (
	feeLimit = big.NewInt(0)
	objTypes = make(map[types.TransactionType]struct{})
)

func init() {
	objTypes[types.ConsensusLeague] = struct{}{}
	objTypes[types.LockLeague] = struct{}{}
	objTypes[types.RaiseUT] = struct{}{}
}
func containObjTypes(key types.TransactionType) bool {
	_, ok := objTypes[key]
	return ok
}

// todo how to do with the nonce
type StateValidate struct {
	TargetHeight       uint64
	currentBlockHeight uint64
	currentBlockHash   common.Hash
	//Only contain changes state
	CacheLMemberState map[common.Address]map[common.Address]*states.LeagueMember

	MemoAccountState map[common.Address]*states.AccountState
	MemoLeagueState  map[common.Address]*states.LeagueState

	TargetLMemberState map[common.Address]map[common.Address]*states.LeagueMember

	//for one tx ,Take the lower limit，only the subtraction
	DirtyAccountState map[common.Address]*states.AccountState
	DirtyLeagueState  map[common.Address]*states.LeagueState
	//one by one,an account can only perform one operation about add or exit
	DirtyLMemberState map[common.Address]map[common.Address]*states.LeagueMember

	//if league is newly created,it can't accept anything else
	LeagueForCreate []*storages.LeagueKV
	SymbolForCreate map[common.Symbol]struct{}
	Blacklist       map[common.Address]bool

	ParentBonusHeight uint64
	WaitOplogs        []*OpLog //Transactional submission

	objTxs      types.Transactions
	receipts    types.Receipts
	txs         types.Transactions
	ledgerStore *storages.LedgerStoreImp

	LeagueGas *big.Int
	M         *sync.Mutex
}

func NewStateValidate(ledgerStore *storages.LedgerStoreImp) *StateValidate {
	log.Info("func validation NewStateValidate")

	curHeight, curHash := ledgerStore.GetCurrentBlock()
	AccountNonceInstance.reset(curHeight)
	s := &StateValidate{
		TargetHeight:       curHeight + 1,
		currentBlockHash:   curHash,
		currentBlockHeight: curHeight,
		MemoAccountState:   make(map[common.Address]*states.AccountState),
		MemoLeagueState:    make(map[common.Address]*states.LeagueState),
		DirtyAccountState:  make(map[common.Address]*states.AccountState),
		DirtyLeagueState:   make(map[common.Address]*states.LeagueState),
		LeagueForCreate:    make([]*storages.LeagueKV, 0),
		TargetLMemberState: make(map[common.Address]map[common.Address]*states.LeagueMember), //league account
		CacheLMemberState:  make(map[common.Address]map[common.Address]*states.LeagueMember),
		DirtyLMemberState:  make(map[common.Address]map[common.Address]*states.LeagueMember),
		Blacklist:          make(map[common.Address]bool),
		SymbolForCreate:    make(map[common.Symbol]struct{}),
		ledgerStore:        ledgerStore,
		WaitOplogs:         []*OpLog{}, //Transactional submission
		LeagueGas:          new(big.Int),
		M:                  new(sync.Mutex),
	}

	return s
}
func (s *StateValidate) GetTxLen() int {
	if s == nil {
		return -1
	}
	return s.txs.Len()
}
func (s *StateValidate) TxInheritance() <-chan *types.Transaction {
	ch := make(chan *types.Transaction, 20)
	for k, _ := range s.txs {
		if !s.ledgerStore.ContainTx(s.txs[k].Hash()) {
			ch <- s.txs[k]
		}
	}
	close(ch)
	return ch
}

//ExecuteOplogs is perform the verified transaction calculation and generate the status data
func (s *StateValidate) ExecuteOplogs() (
	blockInfo *storages.BlockInfo) {
	s.M.Lock()
	defer s.M.Unlock()
	{
		log.Info("func ExecuteOplogs", "blockTargetHeight", s.TargetHeight, "txlen", s.txs.Len())
		var (
			aa  *OPAddressBigInt
			an  *OPAddressUint64
			ad  *OPAddressAdress
			fee = big.NewInt(0)
		)
		/*for k, v := range s.MemoAccountState {
			fmt.Println("ex  you know", k.ToString(), v.Address.ToString(), v.Data.Nonce)

		}*/
		for _, optmp := range s.WaitOplogs {
			v := optmp
			switch v.method {
			case Account_box_add:
				aa = v.data.(*OPAddressBigInt)
				s.MemoAccountState[aa.Address].Data.Balance.Add(s.MemoAccountState[aa.Address].Data.Balance, aa.Num)
				break
			case Account_energy_add:
				aa = v.data.(*OPAddressBigInt)
				s.MemoAccountState[aa.Address].Data.EnergyBalance.Add(s.MemoAccountState[aa.Address].Data.EnergyBalance, aa.Num)
				break
			case Account_box_sub:
				aa = v.data.(*OPAddressBigInt)
				s.MemoAccountState[aa.Address].Data.Balance.Sub(s.MemoAccountState[aa.Address].Data.Balance, aa.Num)
				break
			case Account_energy_sub, Account_energy_consume:
				aa = v.data.(*OPAddressBigInt)
				s.MemoAccountState[aa.Address].Data.EnergyBalance.Sub(s.MemoAccountState[aa.Address].Data.EnergyBalance, aa.Num)
				if v.method == Account_energy_consume {
					fee.Add(fee, aa.Num)
				}
				break
			case Account_bonus_left:
				an = v.data.(*OPAddressUint64)
				s.MemoAccountState[an.Address].Data.BonusHeight = an.Num
				break
			case Account_nonce_add:
				an = v.data.(*OPAddressUint64)
				s.MemoAccountState[an.Address].Data.Nonce = an.Num
				break
			case League_create:
				modelCreate := v.data.(*OPCreateLeague)
				leagueState := &states.LeagueState{
					Address: modelCreate.League,
					Height:  s.TargetHeight,
					MinBox:  modelCreate.Minbox,
					Creator: modelCreate.Creator,
					Rate:    modelCreate.Rate,
					Private: modelCreate.Private,
					Data: &states.League{
						Nonce:      1,
						FrozenBox:  modelCreate.FrozenBox,
						MemberRoot: common.Hash{}, //todo merkle root
						Name:       common.Hash{},
					},
				}
				s.MemoLeagueState[leagueState.Address] = leagueState
				break
			case League_member_add, League_member_apply:
				ad = v.data.(*OPAddressAdress)
				if s.TargetLMemberState[ad.Second] == nil {
					s.TargetLMemberState[ad.Second] = make(map[common.Address]*states.LeagueMember)
				}
				s.TargetLMemberState[ad.Second][ad.First] = s.DirtyLMemberState[ad.Second][ad.First]
				fmt.Printf("🔆 ExecuteOplogs dirtyLMemberState %v\n", s.DirtyLMemberState[ad.Second][ad.First])
				break
			case League_member_remove:
				ad = v.data.(*OPAddressAdress)
				if s.TargetLMemberState[ad.Second] == nil {
					s.TargetLMemberState[ad.Second] = make(map[common.Address]*states.LeagueMember)
				}
				s.TargetLMemberState[ad.Second][ad.First] = s.DirtyLMemberState[ad.Second][ad.First]
				fmt.Printf("🔆 ExecuteOplogs dirtyLMemberState %v\n", s.DirtyLMemberState[ad.Second][ad.First])
			case League_nonce_add:
				an = v.data.(*OPAddressUint64)
				s.MemoLeagueState[an.Address].Data.Nonce = an.Num
				//fmt.Println("execute 2", an.Address.ToString(), an.Num)
				break
			case League_gas_destroy:
				aa = v.data.(*OPAddressBigInt)
				s.LeagueGas.Add(s.LeagueGas, aa.Num)
				break
			case League_raise:
				aa = v.data.(*OPAddressBigInt)
				s.MemoLeagueState[aa.Address].AddMetaBox(aa.Num)
				break
			}
		}
		accountStates := make(states.AccountStates, 0, len(s.MemoAccountState))
		for _, v := range s.MemoAccountState {
			/*	log.Info("--------------------------accountState------------start--------------")
				fmt.Printf("energy:%d,crystal:%d,nonce:%d,height:%d,address:%s\n", v.Data.EnergyBalance.Uint64(), v.Data.Balance.Uint64(), v.Data.Nonce, v.Height, v.Address.ToString())
				as, _ := s.ledgerStore.GetAccountByHeight(v.Height, v.Address)
				if as != nil {

					fmt.Printf("accountState ====> address :%s,nonce:%d,crystal:%d,energy:%d,height:%d \n", as.Address.ToString(), as.Data.Nonce, as.Data.Balance.Uint64(), as.Data.EnergyBalance.Uint64(), as.Height)
					log.Info("--------------------------accountState------------end--------------")
				}*/
			as := v
			accountStates = append(accountStates, as)
			//fmt.Println("execute state rennbon", as.Address.ToString(), as.Data.Nonce)
			/*
				fmt.Println("key", k.ToString())*/
		}

		//todo calculate memberRoot and fill it into memoLeagueState
		leagueMembers := make(states.LeagueMembers, 0)
		for k, v := range s.TargetLMemberState {
			lms := make(states.LeagueMembers, 0)
			for _, vi := range v {
				member := vi
				lms = append(lms, member)
				hashRoot := lms.GetHashRoot()
				s.MemoLeagueState[k].Data.MemberRoot = hashRoot
				leagueMembers = append(leagueMembers, member)
			}
		}
		leagueStates := make(states.LeagueStates, 0, len(s.MemoLeagueState))
		for k, _ := range s.MemoLeagueState {
			leagueStates = append(leagueStates, s.MemoLeagueState[k])
		}
		//todo calculate memberRoot and fill it into memoLeagueState
		fmt.Println("🔨 🔧  ExecuteOplogs txlen:", s.txs.Len())
		txs := make(types.Transactions, 0, len(s.txs)+len(s.objTxs))
		txs = append(txs, s.txs...)
		txs = append(txs, s.objTxs...)
		block := &types.Block{
			Transactions: txs,
			Header: &types.Header{
				Height:        s.TargetHeight,
				PrevBlockHash: s.currentBlockHash,
				Version:       0x01,
				TxRoot:        txs.GetHashRoot(),
				StateRoot:     accountStates.GetHashRoot(),
				LeagueRoot:    leagueStates.GetHashRoot(),
				ReceiptsRoot:  s.receipts.GetHashRoot(),
			},
			Sigs: &types.SigData{},
		}
		// Temporary cancellation, account recovery system is not designed
		//if s.TargetHeight%types.HWidth == 0 {
		//	lvs, err := s.ledgerStore.GetNextHeaderProperty(s.TargetHeight)
		//	if err != nil {
		//		panic(err)
		//	}
		//	fmt.Println("🔨 🔧  ExecuteOplogs pack block head bonus ", lvs)
		//	for k, v := range lvs {
		//		switch k {
		//		case 0:
		//			block.Header.Lv1 = v
		//			break
		//		case 1:
		//			block.Header.Lv2 = v
		//			break
		//		case 2:
		//			block.Header.Lv3 = v
		//			break
		//		case 3:
		//			block.Header.Lv4 = v
		//			break
		//		case 4:
		//			block.Header.Lv5 = v
		//			break
		//		case 5:
		//			block.Header.Lv6 = v
		//			break
		//		case 6:
		//			block.Header.Lv7 = v
		//			break
		//		case 7:
		//			block.Header.Lv8 = v
		//			break
		//		case 8:
		//			block.Header.Lv9 = v
		//			break
		//		}
		//	}
		//}
		feeSum := big.NewInt(0)
		for _, v := range s.receipts {
			feeSum.Add(feeSum, big.NewInt(0).SetUint64(v.GasUsed))
		}

		/*gasDestroy1 := new(big.Int)
		hundred := big.NewInt(100)
		gasDestroy1.Mul(feeSum, hundred)
		gasDestroy1.Div(gasDestroy1, big.NewInt(75))

		gasDestroy2 := new(big.Int)
		gasDestroy2.Mul(s.LeagueGas, hundred)
		gasDestroy2.Div(gasDestroy2, big.NewInt(30))

		gasDestroy1.Add(gasDestroy1, gasDestroy2)*/
		blockkInfo := &storages.BlockInfo{
			Block:         block,
			Receipts:      s.receipts,
			AccountStates: accountStates,
			LeagueStates:  leagueStates,
			Members:       leagueMembers,
			FeeSum:        feeSum,
			LeagueKVs:     s.LeagueForCreate,
			ObjTxs:        s.objTxs,
			//GasDestroy:    gasDestroy1,
		}
		//return block, s.receipts, accountStates, leagueStates, leagueMembers, fee
		return blockkInfo
	}
}

//VerifyTx is verify that the transaction is valid and return the processing plan
//returns:
// 	int   ------   -1: throw tx,0: back to queue, 1: efficient transaction
//  error ------   Existing warning
// if verify pass,tx is going to change StateValidate，append WaitOplogs,to affect dirtyState
func (s *StateValidate) VerifyTx(tx *types.Transaction) (int, error) {
	//log.Info("func VerifyTx 01", "tx.Hash", tx.Hash().String())
	//log.Info("func VerifyTx 02", "txlen", s.txs.Len())
	//fmt.Println("bbbb", tx.TxData.Froms.Tis[0].Nonce)
	err := s.CheckSign(tx)
	if err != nil {
		return -1, err
	}

	err = s.CheckFee(tx)
	if err != nil {
		return -1, err

	}
	//fmt.Println("---------tx start--------")
	//defer fmt.Println("---------tx end--------")
	oplogs := HandleTransaction(tx)
	l := len(oplogs)
	//route oplogs to its homestate
	accountOps := make([]*OpLog, 0, l)
	leagueOps := make([]*OpLog, 0, l)
	bonusOps := make([]*OpLog, 0, 3)
	//feeOp := &OpLog{}
	gused := uint64(0)
	if tx.TxData.Fee != nil {
		gused = tx.TxData.Fee.Uint64()
	}
	receipt := &types.Receipt{
		GasUsed: gused,
		TxHash:  tx.Hash(),
	}
	//check exits in dirtystate,if not then get
	for _, v := range oplogs {
		opTmp := v
		if v.method < AccountMaxLine {
			accountOps = append(accountOps, opTmp)
			/*if v.method == Account_energy_consume {
				feeOp = opTmp
			}*/
		} else if v.method > LeagueMinLine && v.method < LeagueMaxLine {
			leagueOps = append(leagueOps, opTmp)
		} else if v.method == Account_bonus_fee {
			bonusOps = append(bonusOps, opTmp)
		}
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
		/*	err = s.verifyAccount(feeOp)
			if err != nil {
				return 0, err
			}
			s.appendWaitOplogs(feeOp)
			s.appendReceipt(receipt)
			s.appendTx(tx)
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
	opIndes2, err := s.VerifyLeague(leagueOps)
	if err != nil {
		//TODO if league has something that has to be done
		s.RollbackAccount(accountOps, opIndes1)
		s.RollbackLeague(leagueOps, opIndes2)
		return -1, err
		/*if err == errors.ERR_STATE_LEAGUE_NONCE_BIGGER {
			return 0, err
		} else {*/

		//}
		/*	err = s.verifyAccount(feeOp)
			s.appendWaitOplogs(feeOp)
			s.appendReceipt(receipt)
			s.appendTx(tx)
			return 1, nil*/
	}

	receipt.Status = true

	s.appendReceipt(receipt)
	s.appendWaitOplogs(accountOps...)
	s.appendWaitOplogs(leagueOps...)
	s.appendWaitOplogs(bonusOps...)
	s.appendTx(tx)

	//log.Info("func VerifyTx 03", "appendTx", tx.Hash().String())

	if tx.TxType == types.EnergyToLeague {
		fmt.Println("转energy到圈子")
	}
	return 1, nil
}

func (s *StateValidate) appendWaitOplogs(oplogs ...*OpLog) {
	s.WaitOplogs = append(s.WaitOplogs, oplogs...)
}

func (s *StateValidate) appendReceipt(receipt *types.Receipt) {
	s.receipts = append(s.receipts, receipt)
}
func (s *StateValidate) appendTx(tx *types.Transaction) {
	if containObjTypes(tx.TxType) {
		s.objTxs = append(s.objTxs, tx)
	} else {
		s.txs = append(s.txs, tx)
	}
}

func (s *StateValidate) CheckSign(tx *types.Transaction) error {
	if !containObjTypes(tx.TxType) {
		b, err := tx.VerifySignature()
		if err != nil {
			return err
		}
		if !b {
			return errors.ERR_TX_SIGN
		}
	}
	return nil
}

//todo
func (s *StateValidate) CheckFee(tx *types.Transaction) error {
	return nil
	/*	if containObjTypes(tx.TxType) {
			return nil
		} else {
			feeLimit.SetUint64(2)
			if tx.TxData.Fee.Cmp(feeLimit) == -1 {
				return errors.NewErr("fee not enought")
			}
			return nil
		}*/
}
func (s *StateValidate) CheckNodeId(nodeId common.Address) error {
	return nil
}

func (s *StateValidate) RemoveElementInLeagueForCreate(address common.Address) {
	index := -1
	for k, v := range s.LeagueForCreate {
		if address == v.K {
			index = k
		}
	}
	if index != -1 {
		if index == len(s.LeagueForCreate) {
			s.LeagueForCreate = s.LeagueForCreate[:index]
		} else {
			s.LeagueForCreate = append(s.LeagueForCreate[:index], s.LeagueForCreate[index+1:]...)
		}
	}
}
