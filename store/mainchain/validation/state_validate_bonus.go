package validation

import (
	"math/big"

	"sync"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/mainchain/account_level"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

//just sub balance
func (s *StateValidate) verifyBonus(oplog *OpLog) (extra []*OpLog, err error) {
	s.M.Lock()
	defer s.M.Unlock()
	{
		if oplog.method == Account_bonus_fee {
			aa := oplog.data.(*OPAddressBigInt)
			err := s.getAccountState(aa.Address)
			if err != nil {
				return nil, err
			}
			amount, bonusHeight, err := calculateReward(s.ledgerStore, aa.Address, s.MemoAccountState[aa.Address].Data.BonusHeight, s.currentBlockHeight)
			if err != nil {
				return nil, err
			}
			if amount.Cmp(aa.Num) == -1 {
				return nil, errors.ERR_STATE_BONUS_NOT_ENOUGH
			}
			amount.Sub(amount, aa.Num)
			ops := analysisBonusAfter(aa.Address, bonusHeight, amount)
			return ops, nil

		}
		return nil, nil
	}
}

func (s *StateValidate) calculateReward(account common.Address, bonusLeft uint64) (amount *big.Int, bonusHeight uint64, err error) {

	bonusLeft = bonusLeft + types.HWidth
	// get account state between bonusHeight+1 and s.targetHeight
	// get headlvs
	// get height level of account
	ass, err := s.ledgerStore.GetAccountRange(bonusLeft, s.currentBlockHeight, account)
	if err != nil {
		return nil, 0, err
	}
	as, _ := s.ledgerStore.GetAccountByHeight(bonusLeft, account)
	if ass.Len() > 0 {
		if as.Height < ass[0].Height {
			ass = append(states.AccountStates{as}, ass...)
		}
	} else {
		ass = states.AccountStates{as}
	}
	assFTree := make([]common.FTreer, 0, len(ass))
	for _, v := range ass {
		assFTree = append(assFTree, v)
	}

	lvlFTree := s.ledgerStore.GetAccountLvlRange(bonusLeft, s.currentBlockHeight, account)
	lvl := s.ledgerStore.GetAccountLevel(bonusLeft, account)

	if len(lvlFTree) > 0 {
		if bonusLeft < lvlFTree[0].GetHeight() {
			lvlFTree = append([]common.FTreer{&account_level.HeightLevel{Height: bonusLeft, Lv: lvl}}, lvlFTree...)
		}
	} else {
		lvlFTree = []common.FTreer{&account_level.HeightLevel{Height: bonusLeft, Lv: lvl}}
	}
	hbonus := s.ledgerStore.GetBouns()
	amount, bonusHeight = aggregateBonus(assFTree, lvlFTree, hbonus, bonusLeft, s.currentBlockHeight, types.HWidth)
	return
}

func aggregateBonus(assFTree, lvlFTree []common.FTreer, hbonus map[uint64][]uint64, start, end, lapWidth uint64) (amount *big.Int, bonusLeft uint64) {
	if start == 0 {
		start += lapWidth
	}
	l := (end-start)/lapWidth + 1
	balances := make([]*big.Int, 0, l)
	levels := make([]account_level.EasyLevel, 0, l)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		//fmt.Println("㊗️ ㊗️ ㊗️ aggregateBonus", start, end, lapWidth, len(assFTree))
		aft := common.NewForwardTree(lapWidth, start, end, assFTree)
		for !aft.Next() {
			balances = append(balances, aft.Val().(*big.Int))
		}
		wg.Done()
	}()
	go func() {
		lft := common.NewForwardTree(lapWidth, start, end, lvlFTree)
		for !lft.Next() {
			levels = append(levels, lft.Val().(account_level.EasyLevel))
		}
		wg.Done()
	}()
	wg.Wait()
	sum := big.NewInt(0)
	sp := big.NewInt(0)
	bonBig := big.NewInt(0)
	/*for k, v := range hbonus {
		fmt.Println("🚫 🚫 🚫 🚫 🚫 🚫 🚫 🚫", k, v, start)
	}*/

	for k, v := range levels {
		if v > 1 {
			//fmt.Println(start, int(v-1))
			bal := balances[k]
			bon := hbonus[start][int(v)-1]
			sp.Mul(bal, bonBig.SetUint64(bon))
			sum.Add(sum, sp)
		}
		bonusLeft = start
		start += lapWidth
	}
	return sum, bonusLeft
}
func CalculateBonus(ledger *storages.LedgerStoreImp, account common.Address) (amount *big.Int, bonusHeight uint64, err error) {
	as, err := ledger.GetAccountByHeight(ledger.GetCurrentBlockHeight(), account)
	if err != nil {
		return nil, 0, err
	}
	return calculateReward(ledger, account, as.Data.BonusHeight, ledger.GetCurrentBlockHeight())
}
func calculateReward(ledger *storages.LedgerStoreImp, account common.Address, bonusLeft, curHeight uint64) (amount *big.Int, bonusHeight uint64, err error) {
	bonusLeft = bonusLeft + types.HWidth
	// get account state between bonusHeight+1 and s.targetHeight
	// get headlvs
	// get height level of account
	ass, err := ledger.GetAccountRange(bonusLeft, curHeight, account)
	if err != nil {
		return nil, 0, err
	}
	as, _ := ledger.GetAccountByHeight(bonusLeft, account)
	if ass.Len() > 0 {
		if as.Height < ass[0].Height {
			ass = append(states.AccountStates{as}, ass...)
		}
	} else {
		ass = states.AccountStates{as}
	}
	assFTree := make([]common.FTreer, 0, len(ass))
	for _, v := range ass {
		assFTree = append(assFTree, v)
	}

	lvlFTree := ledger.GetAccountLvlRange(bonusLeft, curHeight, account)
	lvl := ledger.GetAccountLevel(bonusLeft, account)

	if len(lvlFTree) > 0 {
		if bonusLeft < lvlFTree[0].GetHeight() {
			lvlFTree = append([]common.FTreer{&account_level.HeightLevel{Height: bonusLeft, Lv: lvl}}, lvlFTree...)
		}
	} else {
		lvlFTree = []common.FTreer{&account_level.HeightLevel{Height: bonusLeft, Lv: lvl}}
	}
	hbonus := ledger.GetBouns()
	amount, bonusHeight = aggregateBonus(assFTree, lvlFTree, hbonus, bonusLeft, curHeight, types.HWidth)
	return
}
