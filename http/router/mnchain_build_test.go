package router_test

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"

	"testing"

	"fmt"

	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/mock"
)

var (
	nodeId        = mock.AccountNormal_2.Addr
	symbolStr, _  = common.String2Symbol("ABCD1")
	ceateLeagueTx = &maintypes.Transaction{
		Version: maintypes.TxVersion,
		TxType:  maintypes.CreateLeague,
		TxData: &maintypes.TxData{
			From:    mock.AccountNormal_1.Addr,
			Nonce:   1,
			Fee:     big.NewInt(100),
			Rate:    1000000000,
			MinBox:  10,
			MetaBox: big.NewInt(50),
			NodeId:  nodeId,
			Private: false,
			Symbol:  symbolStr,
		}}
)

func TestExecLeagueId(t *testing.T) {
	nonce := uint64(3)
	fromAddr, _ := common.ToAddress("ct0fHcR19s2nDaHxR3CVdUyYxHgsjzp8wrt")
	tx := &maintypes.Transaction{
		Version: maintypes.TxVersion,
		TxType:  maintypes.CreateLeague,
		TxData: &maintypes.TxData{
			From:    fromAddr,
			Nonce:   nonce,
			Fee:     big.NewInt(100),
			Rate:    1000000000,
			MinBox:  10,
			MetaBox: big.NewInt(50),
			NodeId:  nodeId,
			Private: true,
		},
	}
	fmt.Println(maintypes.ToLeagueAddress(tx).ToString())
}

func transactionBuild(nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType) *maintypes.Transaction {
	fromAddr, _ := common.ToAddress(from)
	toAddr, _ := common.ToAddress(to)

	leagueId, _ := common.ToAddress(league)
	var transquantity int64 = 10e11
	switch tp {
	case maintypes.TransferBox, maintypes.TransferEnergy:
		txins := &common.TxIns{}
		txout := common.TxIn{
			Address: fromAddr,
			Amount:  big.NewInt(transquantity),
			Nonce:   nonce,
		}
		txins.Tis = append(txins.Tis, &txout)

		totxins := &common.TxOuts{}
		totxin := common.TxOut{
			Address: toAddr,
			Amount:  big.NewInt(transquantity),
		}
		totxins.Tos = append(totxins.Tos, &totxin)
		txdata := maintypes.TxData{
			From:  fromAddr,
			Froms: txins,
			Tos:   totxins,
			Fee:   big.NewInt(10000000000),
		}
		tx, err := maintypes.NewTransaction(tp, maintypes.TxVersion, &txdata)
		if err != nil {
			panic(err)
		}
		return tx
	case maintypes.EnergyToLeague:
		tx := &maintypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  tp,
			TxData: &maintypes.TxData{
				From:     fromAddr,
				Nonce:    nonce,
				Fee:      big.NewInt(10),
				LeagueId: leagueId,
				Energy:     big.NewInt(10000000000),
			},
		}
		return tx
	case maintypes.CreateLeague:
		tx := ceateLeagueTx
		fmt.Printf("mock1=%s\n", mock.AccountNormal_1.Addr.ToString())
		fmt.Printf("nodeidsdafjsdajsdjsjsjsjd👠=%s\n", nodeId.ToString())

		tx.TxData.From = fromAddr
		tx.TxData.Nonce = nonce
		return tx
	case maintypes.JoinLeague:
		tx := &maintypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  tp,
			TxData: &maintypes.TxData{
				From:        fromAddr,
				Nonce:       nonce,
				Account:     fromAddr,
				LeagueNonce: leagueNonce,
				Fee:         big.NewInt(100),
				LeagueId:    leagueId,
				MinBox:      10,
			},
		}
		return tx
	case maintypes.RecycleMinbox:
		tx := &maintypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  tp,
			TxData: &maintypes.TxData{
				From:     fromAddr,
				Nonce:    nonce,
				Fee:      big.NewInt(100),
				LeagueId: leagueId,
				MinBox:   10,
			},
		}
		return tx
	case maintypes.GetBonus:
		tx := &maintypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  tp,
			TxData: &maintypes.TxData{
				From:  fromAddr,
				Nonce: nonce,
				Fee:   big.NewInt(100),
			},
		}
		return tx
	}
	return nil
}
func TestLeagueId(t *testing.T) {
	fmt.Println(getLeagueId(mock.AccountNormal_1.Addr, mock.AccountNormal_2.Addr, 2).ToString())
	fmt.Println(mock.AccountNormal_1.Addr.ToString())
}

func getLeagueId(account, node common.Address, nonce uint64) common.Address {
	tx := ceateLeagueTx
	tx.TxData.From = account
	tx.TxData.NodeId = node
	tx.TxData.Nonce = nonce
	return maintypes.ToLeagueAddress(tx)
}
func buildLeagueRaw(tx *maintypes.Transaction) []byte {
	var (
		orgTp orgtypes.TransactionType
		orgtx = &orgtypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  orgTp,
			TxData: &orgtypes.TxData{
				From:   tx.TxData.From,
				TxHash: tx.Hash(),
			},
		}
	)
	switch tx.TxType {
	case maintypes.EnergyToLeague:
		orgTp = orgtypes.EnergyFromMain
		orgtx.TxData.Energy = tx.TxData.Energy
		break
	case maintypes.JoinLeague:
		orgTp = orgtypes.Join
		break
	}
	orgtx.TxType = orgTp
	err := orgtx.ToRaw()
	if err != nil {
		return nil
	}
	return orgtx.Raw
}
