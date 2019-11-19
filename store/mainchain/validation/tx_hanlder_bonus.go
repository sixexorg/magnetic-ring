package validation

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

func bonusTxRoute(tx *types.Transaction) ([]*OpLog, bool) {
	switch tx.TxType {
	case types.GetBonus:
		return AnalysisGetBonus(tx), true
	}

	return nil, false
}

func AnalysisGetBonus(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 2)
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_bonus_fee))
	return oplogs
}

func analysisBonusAfter(account common.Address, bonusLeft uint64, amount *big.Int) []*OpLog {
	oplogs := make([]*OpLog, 0, 2)
	oplogs = append(oplogs, oplogPraseAddresUint64(account, bonusLeft, Account_bonus_left))
	oplogs = append(oplogs, oplogPraseAddressBigInt(account, amount, Account_energy_add))
	return oplogs
}
