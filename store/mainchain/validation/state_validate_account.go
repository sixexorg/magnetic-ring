package validation

import (
	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/log"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

//The most important and the most complex
//TODO Decide which oplog needs to be done
//opIndes might the oplogs which need to rollback
func (s *StateValidate) VerifyAccount(oplogs []*OpLog) (opIndes []int, err error) {
	opIndes = make([]int, 0, len(oplogs))
	for k, v := range oplogs {
		opTmp := v
		err = s.verifyAccount(opTmp)
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
	s.M.Lock()
	defer s.M.Unlock()
	{
		if oplog.method <= Account_energy_consume && oplog.method >= Account_box_add {
			aa := oplog.data.(*OPAddressBigInt)
			err := s.getAccountState(aa.Address)
			if err != nil {
				return err
			}
			switch oplog.method {
			case Account_box_add:
				return nil
			case Account_energy_add:
				return nil
			case Account_box_sub:
				crystal := s.DirtyAccountState[aa.Address].Data.Balance
				if crystal.Cmp(aa.Num) == -1 {
					return errors.ERR_STATE_LACK_BOX
				}
				s.DirtyAccountState[aa.Address].Data.Balance.Sub(crystal, aa.Num)
				return nil
			case Account_energy_sub: //special
				energy := s.DirtyAccountState[aa.Address].Data.EnergyBalance
				if energy.Cmp(aa.Num) == -1 {
					return errors.ERR_STATE_LACK_GAS
				}
				s.DirtyAccountState[aa.Address].Data.EnergyBalance.Sub(energy, aa.Num)
				return nil
			case Account_energy_consume:
				energyComsume := s.DirtyAccountState[aa.Address].Data.EnergyBalance
				if energyComsume.Cmp(aa.Num) == -1 {
					//todo blacklist to make the next nonce calculation easier
					s.Blacklist[aa.Address] = true
					return errors.ERR_STATE_GAS_OVER
				}
				s.DirtyAccountState[aa.Address].Data.EnergyBalance.Sub(energyComsume, aa.Num)
				return nil
			}
		} else if oplog.method == Account_nonce_add {
			an := oplog.data.(*OPAddressUint64)
			err := s.getAccountState(an.Address)
			if err != nil {
				return err
			}
			nonceTmp := s.DirtyAccountState[an.Address].Data.Nonce
			//fmt.Println("nonce dirty:", nonceTmp, " nonce from tx:", an.Num)
			if nonceTmp+1 == an.Num {
				s.DirtyAccountState[an.Address].Data.Nonce = an.Num
				AccountNonceInstance.SetNonce(an.Address, an.Num)
				return nil
			} else if nonceTmp >= an.Num {
				return errors.ERR_STATE_NONCE_USED
			}

			return errors.ERR_STATE_ACCOUNT_NONCE_DISCONTINUITY
		}
		return nil
	}
}

func (s *StateValidate) RollbackAccount(oplogs []*OpLog, opIndes []int) {
	for _, v := range opIndes {
		s.rollbackAccount(oplogs[v])
	}
}

//just resub balance
func (s *StateValidate) rollbackAccount(oplog *OpLog) {
	s.M.Lock()
	defer s.M.Unlock()
	{
		if oplog.method <= Account_energy_consume && oplog.method >= Account_box_sub {
			aa := oplog.data.(*OPAddressBigInt)
			switch oplog.method {
			case Account_box_sub:
				crystal := s.DirtyAccountState[aa.Address].Data.Balance
				s.DirtyAccountState[aa.Address].Data.Balance.Add(crystal, aa.Num)
				break
			case Account_energy_sub: //special
				energy := s.DirtyAccountState[aa.Address].Data.EnergyBalance
				s.DirtyAccountState[aa.Address].Data.EnergyBalance.Add(energy, aa.Num)
				break
			case Account_energy_consume:
				energyComsume := s.DirtyAccountState[aa.Address].Data.EnergyBalance
				s.DirtyAccountState[aa.Address].Data.EnergyBalance.Add(energyComsume, aa.Num)
				break
			}
			return
		} else if oplog.method == Account_nonce_add {
			aa := oplog.data.(*OPAddressUint64)
			s.DirtyAccountState[aa.Address].Data.Nonce--
			AccountNonceInstance.nonceSub(aa.Address)
		}
	}
}

//todo fill MemoAccountState and DirtyAccountState for execute verifyAccount
func (s *StateValidate) getAccountState(address common.Address) error {
	if s.DirtyAccountState[address] != nil {
		return nil
	}
	as, err := s.ledgerStore.GetAccountByHeight(s.currentBlockHeight, address)
	if err != nil {
		return err
	}
	//fmt.Println("func getAccountState", as.Data.Nonce, address.ToString(), as.Address.ToString(), as.Data.EnergyBalance.Uint64(), as.Data.Balance.Uint64())
	//as.Address = address
	log.Info(fmt.Sprintf("func getAccountState ====> address :%s,nonce:%d,crystal:%d,energy:%d,height:%d", as.Address.ToString(),
		as.Data.Nonce, as.Data.Balance.Uint64(), as.Data.EnergyBalance.Uint64(), as.Height))
	asDirty := &states.AccountState{}
	common.DeepCopy(&asDirty, as)
	s.MemoAccountState[address] = as
	s.DirtyAccountState[address] = asDirty
	return nil
}
