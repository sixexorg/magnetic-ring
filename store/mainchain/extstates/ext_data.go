package extstates

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/store/storelaw"
)

type ExtData struct {
	Height        uint64
	LeagueBlock   *LeagueBlockSimple //block easy info
	MainTxUsed    common.HashArray   //main tx used by the current block height
	AccountStates []*EasyLeagueAccount
}
type EasyLeagueAccount struct {
	Acc common.Address
	*Account
}

func (this *ExtData) GetOrgBlockInfo() (blkInfo *storelaw.OrgBlockInfo, mainTxUsed common.HashArray) {
	lass := make(storelaw.AccountStaters, 0, len(this.AccountStates))
	for _, v := range this.AccountStates {
		lass = append(lass, &LeagueAccountState{
			Address:  v.Acc,
			LeagueId: this.LeagueBlock.LeagueId,
			Height:   this.LeagueBlock.Height,
			Data:     v.Account,
		})
	}
	obi := &storelaw.OrgBlockInfo{
		Block: &types.Block{
			Header: this.LeagueBlock.Header,
		},
		AccStates: lass,
		FeeSum:    this.LeagueBlock.EnergyUsed,
	}
	return obi, this.MainTxUsed
}
