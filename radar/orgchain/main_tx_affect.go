package orgchain

import (
	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
)

const (
	applyInterval = 3
)

type MainChainCfm struct {
	Start     uint64
	End       uint64
	BlockRoot common.Hash
	Locked    bool
}
type ResultOp struct {
	Err   chan error
	MCC   chan *MainChainCfm
	ORGTX chan *orgtypes.Transaction
}

func (this *Radar) CheckCfm(tx *maintypes.Transaction) error {
	if this.beVerified {
		return errors.ERR_RADAR_VERIFIED
	}
	if this.cfmEndHeight+1 > tx.TxData.StartHeight {
		return errors.ERR_RADAR_MAINTX_BEFORE
	}
	if this.cfmEndHeight+1 < tx.TxData.StartHeight {
		return errors.ERR_RADAR_MAINTX_AFTER
	}
	return nil
}

func (this *Radar) AffectConsensusLeague(tx *maintypes.Transaction) (*MainChainCfm, error) {
	err := this.CheckCfm(tx)
	if err != nil {
		return nil, err
	}
	this.cfmStartHeight = tx.TxData.StartHeight
	this.cfmEndHeight = tx.TxData.EndHeight
	this.blockRoot = tx.TxData.BlockRoot
	this.beVerified = false
	mcc := &MainChainCfm{
		Start:     tx.TxData.StartHeight,
		End:       tx.TxData.EndHeight,
		BlockRoot: tx.TxData.BlockRoot,
		Locked:    false,
	}
	return mcc, nil
}
func (this *Radar) AffectLockLeague(tx *maintypes.Transaction) (*MainChainCfm, error) {
	err := this.CheckCfm(tx)
	if err != nil {
		return nil, err
	}
	this.cfmStartHeight = tx.TxData.StartHeight
	this.cfmEndHeight = tx.TxData.EndHeight
	this.blockRoot = tx.TxData.BlockRoot
	this.leagueLocked = true
	mcc := &MainChainCfm{
		Start:     tx.TxData.StartHeight,
		End:       tx.TxData.EndHeight,
		BlockRoot: tx.TxData.BlockRoot,
		Locked:    true,
	}
	return mcc, nil
}

func (this *Radar) AffectEnergyToLeague(tx *maintypes.Transaction) (*orgtypes.Transaction, error) {
	fmt.Println("🔆 💸 ⭕️ The main chain turns energy into the circle")
	/*	cp, err := sink.BytesToComplex(tx.TxData.LeagueRaw)
		if err != nil {
			return nil, err
		}
		orgTx := &orgtypes.Transaction{}
		source := sink.NewZeroCopySource(cp.Data)
		if err = orgTx.Deserialization(source); err != nil {
			return nil, err
		}
		if orgTx.TxData.TxHash != tx.Hash() || orgTx.TxData.From != tx.TxData.From || orgTx.TxData.Energy != tx.TxData.Energy {
			return nil, errors.ERR_RADAR_TX_MOT_MATCH
		}*/
	orgTx := &orgtypes.Transaction{
		Version: orgtypes.TxVersion,
		TxType:  orgtypes.EnergyFromMain,
		TxData: &orgtypes.TxData{
			TxHash: tx.Hash(),
			From:   tx.TxData.From,
			Energy:   tx.TxData.Energy,
		},
	}
	return orgTx, nil
}

func (this *Radar) AffectJoinLeague(tx *maintypes.Transaction) (*orgtypes.Transaction, error) {
	fmt.Println("🔆 🏃 ⭕️ Main chain members join the circle,this.isPrivate=",this.isPrivate)
	//TODO receiptcheck等
	var orgTx *orgtypes.Transaction
	if this.isPrivate {
		nextHeight := this.ledger.GetCurrentBlockHeight() + 1
		orgTx = &orgtypes.Transaction{
			Version: orgtypes.TxVersion,
			TxType:  orgtypes.VoteApply,
			TxData: &orgtypes.TxData{
				TxHash: tx.Hash(),
				From:   tx.TxData.Account,
				Start:  nextHeight,
				End:    nextHeight + applyInterval,
			},
		}
		fmt.Println("⭕️ voteReply hash:", orgTx.Hash().String())
	} else {
		orgTx = &orgtypes.Transaction{
			Version: orgtypes.TxVersion,
			TxType:  orgtypes.Join,
			TxData: &orgtypes.TxData{
				TxHash: tx.Hash(),
				From:   tx.TxData.Account,
			},
		}
	}
	return orgTx, nil
}

func (this *Radar) AffectRaiseUT(tx *maintypes.Transaction) (*orgtypes.Transaction, error) {
	ltx, _, err := this.mainLedger.GetTxByLeagueId(this.leagueId)
	if err != nil {
		return nil, err
	}
	orgTx := &orgtypes.Transaction{
		Version: orgtypes.TxVersion,
		TxType:  orgtypes.RaiseUT,
		TxData: &orgtypes.TxData{
			MetaBox: tx.TxData.MetaBox,
			TxHash:  tx.Hash(),
			Rate:    ltx.TxData.Rate,
		},
	}
	return orgTx, nil
}
