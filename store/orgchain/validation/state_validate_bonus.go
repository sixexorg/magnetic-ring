package validation

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

//just sub balance
func (s *StateValidate) verifyBonus(oplog *OpLog) (extra []*OpLog, err error) {
	if oplog.method == Account_bonus_fee {
		aa := oplog.data.(*OPAddressBigInt)
		s.getAccountState(aa.Address)
		amount, bonusHeight, err := calculateReward(s.ledgeror, aa.Address, s.leagueId, s.MemoAccountState[aa.Address].BonusHeight(), s.currentBlockHeight)
		//s.calculateReward(aa.Address, s.MemoAccountState[aa.Address].BonusHeight())
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

func (s *StateValidate) calculateReward(account common.Address, bonusLeft uint64) (amount *big.Int, bonusHeight uint64, err error) {
	bonusLeft = bonusLeft + types.HWidth
	// get account state between bonusHeight+1 and s.targetHeight
	// get headlvs
	// get height level of account
	ass, err := s.ledgeror.GetAccountRange(bonusLeft, s.currentBlockHeight, account, s.leagueId)
	if err != nil {
		return nil, 0, err
	}
	as, _ := s.ledgeror.GetPrevAccount4V(bonusLeft, account, s.leagueId)
	if ass.Len() > 0 {
		if as.GetHeight() < ass[0].GetHeight() {
			ass = append(storelaw.AccountStaters{as}, ass...)
		}
	} else {
		ass = storelaw.AccountStaters{as}
	}
	assFTree := make([]common.FTreer, 0, len(ass))
	for _, v := range ass {
		assFTree = append(assFTree, v)
	}
	hbonus := s.ledgeror.GetBonus(s.leagueId)
	amount, bonusHeight = aggregateBonus(assFTree, hbonus, bonusLeft, s.currentBlockHeight, types.HWidth)
	return
}

func aggregateBonus(assFTree []common.FTreer, hbonus map[uint64]uint64, start, end, lapWidth uint64) (amount *big.Int, bonusLeft uint64) {
	if start == 0 {
		start += lapWidth
	}
	l := (end-start)/lapWidth + 1
	balances := make([]*big.Int, 0, l)
	aft := common.NewForwardTree(lapWidth, start, end, assFTree)
	for !aft.Next() {
		balances = append(balances, aft.Val().(*big.Int))
	}
	sum := big.NewInt(0)
	sp := big.NewInt(0)
	for _, v := range balances {
		sp.Mul(v, new(big.Int).SetUint64(hbonus[start]))
		sum.Add(sum, sp)
		bonusLeft = start
		start += lapWidth
		sp.SetUint64(0)
	}
	sum.Div(sum, bigU)
	return sum, bonusLeft
}
func CalculateBonus(ledger *storages.LedgerStoreImp, account, leagueId common.Address) (amount *big.Int, bonusHeight uint64, err error) {
	curHeight := ledger.GetCurrentBlockHeight()
	as, err := ledger.GetPrevAccount4V(curHeight, account, leagueId)
	if err != nil {
		return nil, 0, err
	}
	return calculateReward(ledger, account, leagueId, as.BonusHeight(), curHeight)
}
func calculateReward(ledger storelaw.Ledger4Validation, account, leagueId common.Address, bonusLeft, curHeight uint64) (amount *big.Int, bonusHeight uint64, err error) {
	bonusLeft = bonusLeft + types.HWidth
	// get account state between bonusHeight+1 and s.targetHeight
	// get headlvs
	// get height level of account
	ass, err := ledger.GetAccountRange(bonusLeft, curHeight, account, leagueId)
	if err != nil {
		return nil, 0, err
	}
	as, _ := ledger.GetPrevAccount4V(bonusLeft, account, leagueId)
	if ass.Len() > 0 {
		if as.GetHeight() < ass[0].GetHeight() {
			ass = append(storelaw.AccountStaters{as}, ass...)
		}
	} else {
		ass = storelaw.AccountStaters{as}
	}
	assFTree := make([]common.FTreer, 0, len(ass))
	for _, v := range ass {
		assFTree = append(assFTree, v)
	}
	hbonus := ledger.GetBonus(leagueId)
	amount, bonusHeight = aggregateBonus(assFTree, hbonus, bonusLeft, curHeight, types.HWidth)
	return
}
