package router_test

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
)

func orgTransactionBuild(nonce uint64, from, to string, tp orgtypes.TransactionType, txhash string, reply orgtypes.VoteAttitude) *orgtypes.Transaction {
	fromAddr, _ := common.ToAddress(from)
	toAddr, _ := common.ToAddress(to)
	switch tp {
	case orgtypes.TransferUT:
		txins := &common.TxIns{}
		txin := common.TxIn{
			Address: fromAddr,
			Amount:  big.NewInt(100),
			Nonce:   nonce,
		}
		txins.Tis = append(txins.Tis, &txin)

		totxins := &common.TxOuts{}
		totxin := common.TxOut{
			Address: toAddr,
			Amount:  big.NewInt(100),
		}
		totxins.Tos = append(totxins.Tos, &totxin)
		txdata := &orgtypes.TxData{
			From:  fromAddr,
			Froms: txins,
			Tos:   totxins,
			Fee:   big.NewInt(10000000),
			Nonce:nonce,
		}
		tx := &orgtypes.Transaction{
			Version: orgtypes.TxVersion,
			TxType:  tp,
			TxData:  txdata,
		}
		return tx
	case orgtypes.GetBonus:

		tx := &orgtypes.Transaction{
			Version: orgtypes.TxVersion,
			TxType:  tp,
			TxData: &orgtypes.TxData{
				From:  fromAddr,
				Fee:   big.NewInt(1),
				Nonce: nonce,
			},
		}
		return tx
	case orgtypes.VoteIncreaseUT:
		tx := &orgtypes.Transaction{
			Version: orgtypes.TxVersion,
			TxType:  tp,
			TxData: &orgtypes.TxData{
				From:     fromAddr,
				Fee:      big.NewInt(1),
				Nonce:    nonce,
				LeagueId: leagueIdAddress,
				MetaBox:  big.NewInt(20),
				Start:    3,
				End:      5,
			},
		}
		return tx
	case orgtypes.ReplyVote:
		th, _ := common.StringToHash(txhash)
		tx := &orgtypes.Transaction{
			Version: orgtypes.TxVersion,
			TxType:  tp,
			TxData: &orgtypes.TxData{
				From:      fromAddr,
				Fee:       big.NewInt(1),
				Nonce:     nonce,
				TxHash:    th,
				VoteReply: uint8(reply),
			},
		}
		return tx
	}
	return nil
}
