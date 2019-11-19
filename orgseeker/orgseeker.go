package orgseeker

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/central/entity"
	"github.com/sixexorg/magnetic-ring/central/mongo/service/impl"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/log"
	orgstorages "github.com/sixexorg/magnetic-ring/store/orgchain/storages"
	"gopkg.in/mgo.v2"
	"time"
)

type Seeker struct {
	leagueId common.Address
	ledger   *orgstorages.LedgerStoreImp

	MemberAddChan chan string
	MemberOutChan chan string
	txchan        chan *types.Transaction
}

func (seeker *Seeker) dealMemberAddOrOut() {
	for {
		select {
		case addr := <-seeker.MemberAddChan:
			seeker.addMemeber(addr)
		case addr := <-seeker.MemberOutChan:
			seeker.delMemeber(addr)
		case tx := <-seeker.txchan:
			var uttotal uint64 = 0
			if tx.TxType == types.RaiseUT {
				uttotal = seeker.ledger.GetUTByHeight(tx.TxData.Start,common.Address{}).Uint64()
			}
			votedetail := &entity.VoteDetail{
				Id:tx.TransactionHash.String(),
				Hoster:tx.TxData.From.ToString(),
				UtBefore:uttotal,
				Msg:tx.TxData.Msg.String(),
				EndHeight:tx.TxData.End,
			}

			if tx.TxData.MetaBox !=nil {
				votedetail.MetaBox=tx.TxData.MetaBox.Uint64()
			}

			err := impl.NewVoteDetailImpl(votedetail,seeker.leagueId.ToString()).Insert()
			if err != nil {
				log.Error("orgseeker.go insert votedetail","error",err)
			}
		}
	}
}

func NewSeeker(leagueId common.Address, ledger *orgstorages.LedgerStoreImp) *Seeker {
	return &Seeker{
		leagueId:      leagueId,
		ledger:        ledger,
		MemberAddChan: make(chan string, 1024),
		MemberOutChan: make(chan string, 1024),
		txchan:        make(chan *types.Transaction, 1024),
	}
}

func (seeker *Seeker) Start() {
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				seeker.seekBlock()
			}
		}
	}()
	go seeker.dealMemberAddOrOut()
}

func (seeker *Seeker) seekBlock() {

	lhservice := impl.NewLeagueHeightImpl(nil, seeker.leagueId)
	htor, err := lhservice.GetHeightor()
	if err != nil {
		log.Info("seek block failed,pnt=lhservice.GetHeightor()", "error", err)

		if err != mgo.ErrNotFound {
			return
		}

		htor = &entity.LeagueHeightor{seeker.leagueId.ToString(), 1}
		lhservice = impl.NewLeagueHeightImpl(htor, seeker.leagueId)
		err := lhservice.Insert()
		if err != nil {
			log.Error("seek block insert heightor failed", "error", err)
			return
		}
	}
	height := htor.Height
	fmt.Printf("🐞 leagueaddr=%s last height=%d\n",seeker.leagueId.ToString(), height)

	block, err := seeker.ledger.GetBlockByHeight(height)
	if err != nil {
		log.Error("seek block seeker.ledger.GetBlockByHeight(height) failed", "error", err)
		return
	}

	err = seeker.analysisBlock(block)
	if err != nil {
		log.Error("seek block analysisBlock(block) failed", "error", err)
		return
	}

	height++
	err = lhservice.UpdateById(seeker.leagueId,height)

	if err != nil {
		fmt.Printf("👁 err=%v\n",err)
	}

}

func (seeker *Seeker) analysisBlock(block *types.Block) error {

	blockTime := block.Header.Timestamp
	txns := block.Transactions

	txs := make([]interface{}, 0)

	chats := make([]interface{}, 0)

	for _, tx := range txns {

		if tx.TxType == types.RaiseUT || tx.TxType == types.Join || tx.TxType == types.Leave {//投票交易落库
			seeker.txchan<-tx
		}

		switch tx.TxType {
		case types.TransferUT, types.EnergyFromMain, types.EnergyToMain:
			tmp := &entity.OrgTxDto{
				Id:              tx.TransactionHash.String(),
				TxType:          tx.TxType,
				TransactionHash: tx.TransactionHash,
				BlockTime:       blockTime,
				TxData: &entity.OrgTransferDataDto{
					From:   tx.TxData.From,
					Froms:  tx.TxData.Froms,
					Tos:    tx.TxData.Tos,
					Msg:    tx.TxData.Msg,
					Amount: tx.TxData.Amount,
					Fee:    tx.TxData.Fee,
					Energy:   tx.TxData.Energy,
				},
			}
			txs = append(txs, tmp)

		case types.MSG, types.VoteIncreaseUT, types.ReplyVote, types.RaiseUT, types.Join:
			tmpct := &entity.OrgChatDto{
				Id:      tx.TransactionHash.String(),
				Account: tx.TxData.From.ToString(),
				Type:    tx.TxType,
				Msg:     tx.TxData.Msg.String(),
				Hash:    tx.TransactionHash.String(),
				Result:  0,
			}
			chats = append(chats, tmpct)
		}
	}

	if len(txs) > 0 {
		err := impl.NewOrgTxnsImpl(txs, seeker.leagueId.ToString()).Insert()
		if err != nil {
			return err
		}
	}

	if len(chats) > 0 {
		err := impl.NewOrgChatMsgImpl(chats, seeker.leagueId.ToString()).Insert()
		if err != nil {
			return err
		}
	}

	return nil
}

func (seeker *Seeker) addMemeber(memberAddress string) error {
	member := &entity.Member{memberAddress}
	return impl.NewMemberImpl(member, seeker.leagueId.ToString()).Insert()
}

func (seeker *Seeker) delMemeber(memberAddress string) error {
	member := &entity.Member{memberAddress}
	return impl.NewMemberImpl(member, seeker.leagueId.ToString()).Delete()
}
