package validation

import (
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

func systemTxRoute(tx *types.Transaction) ([]*OpLog, bool) {
	switch tx.TxType {
	case types.ConsensusLeague:
		return AnalysisConsensusLeague(tx), true
	case types.RaiseUT:
		return AnalysisRaiseUT(tx), true
	case types.EnergyToMain:
		return AnalysisEnergyToMain(tx), true
	case types.ApplyPass:
		return AnalysisApplyPass(tx), true

	}
	return nil, false
}

func AnalysisConsensusLeague(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 1)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, League_gas_destroy))
	return oplogs
}
func AnalysisRaiseUT(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 4)
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.MetaBox, Account_box_sub))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.LeagueId, tx.TxData.MetaBox, League_raise))
	return oplogs
}
func AnalysisEnergyToMain(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 1)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Energy, Account_energy_add))
	//oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.LeagueId, tx.TxData.MetaBox, League_gas_sub))
	return oplogs
}
func AnalysisApplyPass(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 1)
	oplogs = append(oplogs, oplogPraseAddresAddress(tx.TxData.From, tx.TxData.LeagueId, League_member_add))
	return oplogs
}
