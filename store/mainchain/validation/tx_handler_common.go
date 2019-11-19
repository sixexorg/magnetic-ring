package validation

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

func commonTxRoute(tx *types.Transaction) ([]*OpLog, bool) {
	switch tx.TxType {
	case types.TransferBox:
		return AnalysisTansferBox(tx), true
	case types.TransferEnergy:
		return AnalysisTansferEnergy(tx), true
	case types.CreateLeague:
		return AnalysisCreateLeague(tx), true
	case types.JoinLeague:
		return AnalysisJoinLeague(tx), true
	case types.RecycleMinbox:
		return AnalysisRecycleMinbox(tx), true
	case types.EnergyToLeague:
		return AnalysisEnergyToLeague(tx), true
	}
	return nil, false
}

//data: from+fee
func AnalysisMSG(tx *types.Transaction) []*OpLog {
	op1 := oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume)
	op2 := oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add)
	return []*OpLog{op1, op2}
}

//data:len from
func AnalysisTansferBox(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, len(tx.TxData.Froms.Tis)+len(tx.TxData.Tos.Tos)+1)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	for _, v := range tx.TxData.Froms.Tis {
		oplogs = append(oplogs, oplogPraseAddressBigInt(v.Address, v.Amount, Account_box_sub))
		oplogs = append(oplogs, oplogPraseAddresUint64(v.Address, v.Nonce, Account_nonce_add))
	}
	for _, v := range tx.TxData.Tos.Tos {
		oplogs = append(oplogs, oplogPraseAddressBigInt(v.Address, v.Amount, Account_box_add))
	}
	return oplogs
}

//data:len from
func AnalysisTansferEnergy(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, len(tx.TxData.Froms.Tis)+len(tx.TxData.Tos.Tos)+1)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	for _, v := range tx.TxData.Froms.Tis {
		oplogs = append(oplogs, oplogPraseAddressBigInt(v.Address, v.Amount, Account_energy_sub))
		oplogs = append(oplogs, oplogPraseAddresUint64(v.Address, v.Nonce, Account_nonce_add))
	}
	for _, v := range tx.TxData.Tos.Tos {
		oplogs = append(oplogs, oplogPraseAddressBigInt(v.Address, v.Amount, Account_energy_add))
	}
	return oplogs
}
func AnalysisCreateLeague(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 5)
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.MetaBox, Account_box_sub))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, big.NewInt(0).SetUint64(uint64(tx.TxData.MinBox)), Account_box_sub))
	leagueId := types.ToLeagueAddress(tx)
	oplogs = append(oplogs, oplogPraseCreateLeague(tx.TxData.From, leagueId, tx.TxData.NodeId, tx.TxData.MetaBox, tx.TxData.MinBox, tx.TxData.Rate, tx.TxData.Private, tx.Hash(), tx.TxData.Symbol))

	return oplogs
}
func AnalysisJoinLeague(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 5)
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, big.NewInt(0).SetUint64(tx.TxData.MinBox), Account_box_sub))
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.LeagueId, tx.TxData.MinBox, League_minbox))
	oplogs = append(oplogs, oplogPraseAddresAddress(tx.TxData.Account, tx.TxData.LeagueId, League_member_apply))
	return oplogs
}

func AnalysisEnergyToLeague(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 3)
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Energy, Account_energy_sub))
	return oplogs
}

func AnalysisRecycleMinbox(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 5)
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	oplogs = append(oplogs, oplogPraseAddresAddress(tx.TxData.From, tx.TxData.LeagueId, League_member_remove))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, big.NewInt(0).SetUint64(tx.TxData.MinBox), Account_box_add))
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.LeagueId, tx.TxData.MinBox, League_minbox))
	return oplogs
}
