package validation

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
)

func systemTxRoute(tx *types.Transaction) ([]*OpLog, bool) {
	switch tx.TxType {
	case types.EnergyFromMain:
		return AnalysisEnergyFromMain(tx), true
	case types.RaiseUT:
		return AnalsisRaiseUT(tx), true
	}
	return nil, false
}

func AnalysisEnergyFromMain(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 1)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Energy, Account_energy_add))
	return oplogs
}
func AnalsisRaiseUT(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 2)
	ut := big.NewInt(int64(tx.TxData.Rate))
	ut.Mul(ut, tx.TxData.MetaBox)
	oplogs = append(oplogs, oplogPraseAddressBigInt(common.Address{}, ut, League_Raise_ut))
	return oplogs
}
func AnalysisJoin(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 1)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Energy, Account_energy_add))
	return oplogs
}
