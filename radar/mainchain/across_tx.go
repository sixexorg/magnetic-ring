package mainchain

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
)

type LeagueEntiy struct {
	Name    string
	Rate    uint32
	MinBox  uint64
	Creator common.Address
	MetaBox *big.Int
}

func checkOrgTx(mainTx *maintypes.Transaction, orgTx *orgtypes.Transaction) error {
	if orgTx.TxType == orgtypes.Join &&
		mainTx.TxType == maintypes.JoinLeague &&
		mainTx.TxData.Account == orgTx.TxData.From {
		return nil

	} else if orgTx.TxType == orgtypes.EnergyFromMain &&
		mainTx.TxType == maintypes.EnergyToLeague &&
		mainTx.TxData.From == orgTx.TxData.From &&
		mainTx.TxData.Energy.Cmp(orgTx.TxData.Energy) == 0 {
		return nil
	}
	return errors.ERR_EXTERNAL_TX_REFERENCE_WRONG
}

//todo Currently, the nonce value of the cross-chain is set to 0, and the value is automatically assigned when the main chain txpool is verified.
// The logic is not implemented at present, and the later implementation is implemented.
func orgTxBirthMainTx(tx *orgtypes.Transaction, height uint64, leagueId common.Address, f func(leagueId common.Address) (rate uint32)) *maintypes.Transaction {
	var mainTx *maintypes.Transaction
	switch tx.TxType {
	case orgtypes.EnergyToMain:
		mainTx = &maintypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  maintypes.EnergyToMain,
			TxData: &maintypes.TxData{
				From:     tx.TxData.From,
				LeagueId: leagueId,
				Energy:     tx.TxData.Energy,
			},
		}
	case orgtypes.VoteIncreaseUT:
		mainTx = &maintypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  maintypes.RaiseUT,
			TxData: &maintypes.TxData{
				From:     tx.TxData.From,
				LeagueId: leagueId,
				MetaBox:  tx.TxData.MetaBox,
			},
		}
	case orgtypes.VoteApply:
		mainTx = &maintypes.Transaction{
			Version: maintypes.TxVersion,
			TxType:  maintypes.ApplyPass,
			TxData: &maintypes.TxData{
				From:     tx.TxData.From,
				LeagueId: leagueId,
				TxHash:   tx.Hash(),
			},
		}

	}
	return mainTx
}
