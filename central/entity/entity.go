package entity

import (
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"math/big"
	"time"
)

type VoteUnit struct {
	Type           types.TransactionType `json:"type"`
	HostAddress    string        `json:"host_address"`
	EndBlockHeight uint64                `json:"end_block_height"`
	VoteId         string           `json:"vote_id"`

	Account string `json:"account"`
	VoteReply  uint8 `json:"vote_reply"`
	Quntity uint64 `json:"quntity"`
}

type VoteDetail struct {
	Id        string `json:"id"`
	MetaBox   uint64 `json:"meta_box"`
	Hoster    string `json:"hoster"`
	UtBefore  uint64 `json:"ut_before"`
	Msg       string `json:"msg"`
	EndHeight uint64 `bson:"endheight" json:"end_height"`

	MyUt  uint64 `bson:"-" json:"my_ut"`
	State bool   `bson:"-" json:"state"`
	IfIVoted bool `json:"if_i_voted"`

	AgreePer   float64 `bson:"-" json:"agree_per"`
	AgainstPer float64 `bson:"-" json:"against_per"`
	GiveUpPer  float64 `bson:"-" json:"give_up_per"`
}

func NewVoteUnit(tp types.TransactionType, hostAddr common.Address, endH uint64, voteId string,account string,reply uint8,quntity uint64) *VoteUnit {
	return &VoteUnit{tp, hostAddr.ToString(), endH, voteId,account,reply,quntity}
}

type LeagueHeightor struct {
	Id     string `json:"id"`
	Height uint64 `bson:"height"`
}

type OrgTxDto struct {
	Id              string                `json:"id"`
	TxType          types.TransactionType `json:"tx_type"`
	TransactionHash common.Hash           `json:"transaction_hash"`
	BlockTime       uint64                `json:"block_time"`
	TxData          *OrgTransferDataDto   `json:"tx_data"`
}

type OrgTransferDataDto struct {
	From   common.Address `json:"from"`
	Froms  *common.TxIns  `json:"froms"`
	Tos    *common.TxOuts `json:"tos"`
	Msg    common.Hash    `json:"msg"`
	Amount uint64         `json:"amount"`
	Fee    *big.Int       `json:"fee"`
	Energy   *big.Int       `json:"b_gas"`
}

type OrgChatDto struct {
	Id        string                `json:"id"`
	Account   string                `json:"account"`
	Type      types.TransactionType `json:"type"`
	Msg       string                `json:"msg"`
	Hash      string                `json:"hash"`
	Status    bool                  `json:"status"`
	Result    int                   `json:"result"`
	BlockTime time.Time             `bson:"blockTime" json:"block_time"`

}

type Member struct {
	Addr string `json:"addr"`
}
