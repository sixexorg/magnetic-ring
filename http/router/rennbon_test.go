package router_test

import (
	"fmt"
	"testing"

	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/http/cli"
	"github.com/sixexorg/magnetic-ring/http/req"
	"github.com/sixexorg/magnetic-ring/mock"
)

func TestMainChain(t *testing.T) {
	var (
		//from = mock.AccountNormal_1.Addr.ToString()
		from = "ct1j95gw45wnmdazmgpux7wc25tTrT5ExXg"
		to   = "ct1b3cznr6mAmE7Ab4cewd6xe6rtctjpgfv"

		nonce, leagueNonce uint64 = 1, 2

		tp types.TransactionType = types.JoinLeague
	)
	/*TransferBox
	GetBonus
	RecycleMinbox
	TransferEnergy
	CreateLeague
	JoinLeague
	EnergyToLeague */
	//nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType)
	tx := transactionBuild(nonce, leagueNonce, from, to, leagueId, tp)
	fmt.Printf("☺️ ☺️ ☺️ ☺️ ☺️ -->%s\n",leagueId)
	err := tx.ToRaw()
	if err != nil {
		panic(err)
	}
	req := req.ReqSendTx{
		Raw: tx.Raw,
	}
	fullurl := fmt.Sprintf("%s/block/sendtx", baseurl)
	resp, err := cli.HttpSend(fullurl, req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("txhash=%s,\n", resp)
}
func TestOrgChain(t *testing.T) {
	var (
		nonce     uint64                   = 4
		from                               = mock.AccountNormal_1.Addr.ToString()
		to                                 = "ct0fHcR19s2nDaHxR3CVdUyYxHgsjzp8wrt"
		tp        orgtypes.TransactionType = orgtypes.ReplyVote
		voteReply                          = orgtypes.Agree
		txhash                             = "52fcc88ab3521b1fbcc0cf2e7132cb67540c3826b390dcbc0fa6c7d744c7d149"
		//"52fcc88ab3521b1fbcc0cf2e7132cb67540c3826b390dcbc0fa6c7d744c7d149" //"eb7e7a23e70cc64c2e9d5fc8a854dd5d5ebc2c195c12eb9f9db0d220c84e258d"
	)
	/*TransferUT
	ReplyVote
	VoteIncreaseUT
	GetBonus
	Join
	Leave
	EnergyFromMain
	EnergyToMain   */
	//nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType)
	tx := orgTransactionBuild(nonce, from, to, tp, txhash, voteReply)
	err := tx.ToRaw()
	fmt.Println("hash ", tx.Hash())
	sk := sink.NewZeroCopySource(tx.Raw)
	txTmp := &orgtypes.Transaction{}
	err = txTmp.Deserialization(sk)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err != nil {
		t.Errorf("toraw err-->%v\n", err)
		return
	}
	//ct2aQyCjVfJ5vgh8x8xb2yRbTu6741Yg8u7
	fmt.Printf("••••••••••••••••••••••••••••herhe-->%s\n", leagueId)

	req := req.ReqSendTx{
		LeagueId: leagueId,
		Raw:      tx.Raw,
	}
	fullurl := fmt.Sprintf("%s/block/sendsubtx", baseurl)
	resp, err := cli.HttpSend(fullurl, req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("txhash=%s,\n", resp)
}
