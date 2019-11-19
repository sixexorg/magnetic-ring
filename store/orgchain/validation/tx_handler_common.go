package validation

import "github.com/sixexorg/magnetic-ring/core/orgchain/types"

func commonTxRoute(tx *types.Transaction) ([]*OpLog, bool) {
	switch tx.TxType {
	case types.TransferUT:
		return AnalysisTansferUT(tx), true
	case types.EnergyToMain:
		return AnalysisEnergyToMain(tx), true
	case types.MSG:
		AnalysisMSG(tx)
	}
	return nil, false
}

//data: from+fee
func AnalysisMSG(tx *types.Transaction) []*OpLog {
	op1 := oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume)
	op2 := oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add)
	//TODO
	return []*OpLog{op1, op2}
}

//data:len from
func AnalysisTansferUT(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, len(tx.TxData.Froms.Tis)+len(tx.TxData.Tos.Tos)+1)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	for _, v := range tx.TxData.Froms.Tis {
		oplogs = append(oplogs, oplogPraseAddressBigInt(v.Address, v.Amount, Account_ut_sub))
		oplogs = append(oplogs, oplogPraseAddresUint64(v.Address, v.Nonce, Account_nonce_add))
	}
	for _, v := range tx.TxData.Tos.Tos {
		oplogs = append(oplogs, oplogPraseAddressBigInt(v.Address, v.Amount, Account_ut_add))
	}
	return oplogs
}

func AnalysisEnergyToMain(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 3)
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Energy, Account_energy_sub))
	return oplogs
}
