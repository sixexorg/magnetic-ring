package validation

import "github.com/sixexorg/magnetic-ring/core/orgchain/types"

func voteTxRoute(tx *types.Transaction) ([]*OpLog, bool) {
	switch tx.TxType {
	case types.VoteIncreaseUT:
		return AnalysisVoteIncreaseUT(tx), true
	case types.VoteApply:
		return AnalysisVoteApply(tx), true
	case types.ReplyVote:
		return AnalysisReplyVote(tx), true
	}

	return nil, false
}

func AnalysisVoteIncreaseUT(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 2)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	return oplogs
}

func AnalysisReplyVote(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0, 3)
	oplogs = append(oplogs, oplogPraseAddressBigInt(tx.TxData.From, tx.TxData.Fee, Account_energy_consume))
	oplogs = append(oplogs, oplogPraseAddresUint64(tx.TxData.From, tx.TxData.Nonce, Account_nonce_add))
	m := Vote_against
	if tx.TxData.VoteReply == 1 {
		m = Vote_agree
	} else if tx.TxData.VoteReply == 2 {
		m = Vote_abstention
	}
	oplogs = append(oplogs, oplogPraseHashAddress(tx.TxData.TxHash, tx.TxData.From, m))
	return oplogs
}
func AnalysisVoteApply(tx *types.Transaction) []*OpLog {
	oplogs := make([]*OpLog, 0)
	return oplogs
}

func voteTypeContains(typ types.TransactionType) bool {
	if typ == types.VoteApply || typ == types.VoteIncreaseUT {
		return true
	}
	return false
}
