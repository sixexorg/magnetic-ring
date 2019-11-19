package orgchain

import (
	"fmt"

	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/errors"
)

func (this *Radar) RoutingMainTx(tx *maintypes.Transaction) *ResultOp {
	fmt.Println("♻️ Recycling the main chain transaction to the circle")
	re := &ResultOp{}


	fmt.Printf("tx.txdata.leagueId=%s,this.leagueId=%s\n",tx.TxData.LeagueId.ToString(),this.leagueId.ToString())

	if !containMainType(tx.TxType) && tx.TxData.LeagueId != this.leagueId {
		errch := make(chan error, 1)
		errch <- errors.ERR_RADAR_TX_USELESS
		close(errch)
		re.Err = errch
		return re
	}
	switch tx.TxType {
	case maintypes.ConsensusLeague:
		mcc, err := this.AffectConsensusLeague(tx)
		if err != nil {
			errch := make(chan error, 1)
			errch <- err
			close(errch)
			re.Err = errch

		} else {
			mccch := make(chan *MainChainCfm, 1)
			mccch <- mcc
			close(mccch)
			re.MCC = mccch
		}
		return re
	case maintypes.LockLeague:
		mcc, err := this.AffectLockLeague(tx)
		if err != nil {
			errch := make(chan error, 1)
			errch <- err
			close(errch)
			re.Err = errch

		} else {
			mccch := make(chan *MainChainCfm, 1)
			mccch <- mcc
			close(mccch)
			re.MCC = mccch
		}
		return re
	case maintypes.JoinLeague:
		fmt.Printf("🎒 🎒 🎒 🎒 🎒 🎒 🎒 🎒 🎒 🎒 --->txhash=%s\n",tx.TransactionHash)
		orgtx, err := this.AffectJoinLeague(tx)
		if err != nil {
			errch := make(chan error, 1)
			errch <- err
			close(errch)
			re.Err = errch
		} else {
			txcch := make(chan *orgtypes.Transaction, 1)
			txcch <- orgtx
			close(txcch)
			re.ORGTX = txcch
		}
		return re

	case maintypes.ApplyPass://TODO



	case maintypes.EnergyToLeague:
		orgtx, err := this.AffectEnergyToLeague(tx)
		if err != nil {
			errch := make(chan error, 1)
			errch <- err
			close(errch)
			re.Err = errch

		} else {
			txcch := make(chan *orgtypes.Transaction, 1)
			txcch <- orgtx
			close(txcch)
			re.ORGTX = txcch
		}
		return re
	case maintypes.RaiseUT:
		orgtx, err := this.AffectRaiseUT(tx)
		if err != nil {
			errch := make(chan error, 1)
			errch <- err
			close(errch)
			re.Err = errch

		} else {
			txcch := make(chan *orgtypes.Transaction, 1)
			txcch <- orgtx
			close(txcch)
			re.ORGTX = txcch
		}
		return re
	}
	return re
}
