package validation

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

//The most important and the most complex
//TODO Decide which oplog needs to be done
//opIndes might the oplogs which need to rollback
func (s *StateValidate) VerifyAccount(oplogs []*OpLog) (opIndes []int, err error) {
	opIndes = make([]int, 0, len(oplogs))
	for k, v := range oplogs {
		err = s.verifyAccount(v)
		if err == nil {
			//already exec
			opIndes = append(opIndes, k)
		} else {
			break
		}
	}
	return opIndes, err
}

//just sub balance
func (s *StateValidate) verifyAccount(oplog *OpLog) error {
	if oplog.method <= Account_energy_consume && oplog.method >= Account_ut_add {
		aa := oplog.data.(*OPAddressBigInt)
		s.getAccountState(aa.Address)
		switch oplog.method {
		case Account_ut_add:
			return nil
		case Account_energy_add:
			return nil
		case Account_ut_sub:
			crystal := s.DirtyAccountState[aa.Address].Balance()
			if crystal.Cmp(aa.Num) == -1 {
				return errors.ERR_STATE_LACK_BOX
			}
			s.DirtyAccountState[aa.Address].BalanceSub(aa.Num)
			return nil
		case Account_energy_sub: //special
			energy := s.DirtyAccountState[aa.Address].Energy()
			if energy.Cmp(aa.Num) == -1 {
				return errors.ERR_STATE_LACK_GAS
			}
			s.DirtyAccountState[aa.Address].EnergySub(aa.Num)
			return nil
		case Account_energy_consume:
			energyComsume := s.DirtyAccountState[aa.Address].Energy()
			if energyComsume.Cmp(aa.Num) == -1 {
				//todo blacklist to make the next nonce calculation easier
				s.Blacklist[aa.Address] = true
				return errors.ERR_STATE_GAS_OVER
			}
			s.DirtyAccountState[aa.Address].EnergySub(aa.Num)
			return nil
		}
	} else if oplog.method == Account_nonce_add {
		an := oplog.data.(*OPAddressUint64)
		s.getAccountState(an.Address)
		nextNonce := s.DirtyAccountState[an.Address].Nonce() + 1
		if nextNonce == an.Num {
			s.DirtyAccountState[an.Address].NonceSet(an.Num)
			return nil
		} else if nextNonce > an.Num {
			return errors.ERR_STATE_NONCE_USED
		} else {
			return errors.ERR_STATE_ACCOUNT_NONCE_DISCONTINUITY
		}
	}
	return nil
}

func (s *StateValidate) RollbackAccount(oplogs []*OpLog, opIndes []int) {
	for _, v := range opIndes {
		s.rollbackAccount(oplogs[v])
	}
}

//just resub balance
func (s *StateValidate) rollbackAccount(oplog *OpLog) {
	if oplog.method <= Account_energy_consume && oplog.method >= Account_ut_sub {
		aa := oplog.data.(*OPAddressBigInt)
		switch oplog.method {
		case Account_ut_sub:
			s.DirtyAccountState[aa.Address].BalanceAdd(aa.Num)
			break
		case Account_energy_sub, Account_energy_consume: //special
			s.DirtyAccountState[aa.Address].BalanceAdd(aa.Num)
			break
		}
		return
	}
	/*else if oplog.method == Account_nonce_add {
		an := oplog.data.(*OPAddressUint64)
		s.DirtyAccountState[an.Address].Data.Nonce--
	}*/

}

//todo fill MemoAccountState and DirtyAccountState for execute verifyAccount
//todo need optimize the Nil state in which a single Block retrieves data multiple times
func (s *StateValidate) getAccountState(address common.Address) {
	//fmt.Println("func GetAccountState address:", address.ToString())
	if s.DirtyAccountState[address] != nil {
		return
	}
	var (
		as  storelaw.AccountStater
		err error
	)
	as, err = s.ledgeror.GetPrevAccount4V(s.currentBlockHeight, address, s.leagueId)
	if err != nil {
		panic(err)
	}
	as.HeightSet(s.TargetHeight)
	var asDirty *states.AccountState

	common.DeepCopy(&asDirty, as)
	s.MemoAccountState[address] = as
	s.DirtyAccountState[address] = asDirty
}

func (s *StateValidate) GetPrevAccountState(account common.Address, height uint64) storelaw.AccountStater {
	as, err := s.ledgeror.GetPrevAccount4V(height, account, s.leagueId)
	if err != nil {
		panic(err)
	}
	return as
}
