package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/sixexorg/magnetic-ring/http/cli"
	"github.com/sixexorg/magnetic-ring/signer/mainchain/signature"

	"github.com/sixexorg/magnetic-ring/mock"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/http/req"
	txpool "github.com/sixexorg/magnetic-ring/txpool/mainchain"
)

var (
	//baseurl = "http://192.168.10.166:10086"
	baseurl = "http://127.0.0.1:20086"
	//baseurl  = "http://192.168.9.123:10086"
	leagueId           = getLeagueId(mock.AccountNormal_1.Addr, mock.AccountNormal_2.Addr, 2).ToString()
	leagueIdAddress, _ = common.ToAddress(leagueId)
)

func TestAccount(t *testing.T) {
	fmt.Println(mock.AccountNormal_1.Address().ToString())
	fmt.Println(leagueId)
}

func TestTransferBox(t *testing.T) {
	TransferBox(5,t)
}

func TestTransferEnergy(t *testing.T) {
	TransferEnergy(t)
}

func TestCreateLeague(t *testing.T) {
	CreateLeague(t)
}

func TestEnergyToLeague(t *testing.T) {
	EnergyToLeague(t)
}

func TransferBox(nonce uint64,t *testing.T) {
	var (
		from = mock.AccountNormal_1.Addr.ToString()
		to   = "ct1j95gw45wnmdazmgpux7wc25tTrT5ExXg"

		//nonce, leagueNonce uint64 = 4, 2

		tp types.TransactionType = types.TransferBox
	)
	/*TransferBox
	TransferEnergy
	CreateLeague
	JoinLeague
	EnergyToLeague */
	//nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType)
	tx := transactionBuild(nonce, 2, from, to, leagueId, tp)
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

func TransferEnergy(t *testing.T) {
	var (
		from = mock.AccountNormal_1.Addr.ToString()
		to   = "ct1j95gw45wnmdazmgpux7wc25tTrT5ExXg" //"ct2j9ywuyqz8VW7ZGBBd26V4VqAWd4r43Vs"

		nonce, leagueNonce uint64 = 6, 2

		tp types.TransactionType = types.TransferEnergy
	)
	/*TransferBox
	TransferEnergy
	CreateLeague
	JoinLeague
	EnergyToLeague */
	//nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType)
	tx := transactionBuild(nonce, leagueNonce, from, to, leagueId, tp)
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

func EnergyToLeague(t *testing.T) {
	var (
		from = mock.AccountNormal_1.Addr.ToString()
		to   = "ct1j95gw45wnmdazmgpux7wc25tTrT5ExXg"

		nonce, leagueNonce uint64 = 3, 2

		tp types.TransactionType = types.TransferEnergy
	)
	/*TransferBox
	TransferEnergy
	CreateLeague
	JoinLeague
	EnergyToLeague */
	//nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType)
	tx := transactionBuild(nonce, leagueNonce, from, to, leagueId, tp)
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

func TestSendTx(t *testing.T) {
	var (
		to = "ct0fHcR19s2nDaHxR3CVdUyYxHgsjzp8wrt"

		nonce, leagueNonce uint64 = 1, 2
		from                      = mock.AccountNormal_1.Addr.ToString()

		tp types.TransactionType = types.JoinLeague
	)
	/*TransferBox
	TransferEnergy
	CreateLeague
	JoinLeague
	EnergyToLeague */
	//nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType)
	tx := transactionBuild(nonce, leagueNonce, from, to, leagueId, tp)
	tx.Sigs = new(common.SigPack)
	tx.Templt = common.NewSigTmplt()

	maue := make(common.Maus, 0)
	mau := common.Mau{}
	mau.M = 1
	mau.Pubks = make([]common.PubHash, 0)

	pubkstr := "04f52f8ccc9de7496af4460da2bfe8762fd64dfabada237542824d327b350f415c276c92c56656f2cfa30c93c852609ef7bf55234bca3106d5e588acb025929356"
	puhash := common.PubHash{}

	pkbuf, _ := common.Hex2Bytes(pubkstr)

	t.Logf("pkbuflen=%d\n", len(pkbuf))

	copy(puhash[:], pkbuf[:])

	mau.Pubks = append(mau.Pubks, puhash)

	maue = append(maue, mau)

	tx.Templt.SigData = append(tx.Templt.SigData, &maue)

	fromAddr, err := common.ToAddress(from)
	if err != nil {
		fmt.Printf("common to address error=%v\n", err)
		return
	}

	fmt.Printf("👀 %s\n", fromAddr.ToString())

	err = signature.SignTransaction(mock.Mgr, fromAddr, tx, "jjjkkk")
	if err != nil {
		fmt.Printf("sign transaction error=%v\n", err)
		return
	}

	be, eb := tx.VerifySignature()
	if eb != nil {
		fmt.Printf("eb=%v\n", eb)
		return
	}
	fmt.Printf("be-🎶 🎶 🎶 🎶 🎶 🎶 🎶->%t\n", be)

	err = tx.ToRaw()
	if err != nil {
		panic(err)
	}
	////return
	//ra := common.Hex2Bytes("0110600008f85ef85cf85af858954a3dcbc73c7f8a5a5b751078347790a2412d206300b84078d07af214b8cbad56f92fcd7d2ae5f82489678258f0c6ad1aa7bad4bea6d24dcd9305fa590584beafa42400431ec960d838981c3ac0901de0c911df3f34b2bf4e0009f84cf84af848f84601f843b84104f992c8418749fc5d4bbcaffc4ff120b20b5550f0b5744cff2dd55ad1bfa7f7db94891af2315443a35330a6527436037b7b3d6de5c41e16171fd461e73b5434500100036443d34df4118065f93c4bee56b5d61d4a3b3d6918011d0001dcdbda9543d34df4118065f93c4bee56b5d61d4a3b3d6918018203e80200000000000000000000000000000000000000000000000000000000000000001c0002dbdad9957417702301a3130edc4bdb35ef760fc47f53f2f9008203e8")
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

func CreateLeague(t *testing.T) {
	var (
		//from                      = "ct0fHcR19s2nDaHxR3CVdUyYxHgsjzp8wrt"
		to = mock.AccountNormal_1.Addr.ToString()

		nonce, leagueNonce uint64 = 2, 2
		from                      = mock.AccountNormal_1.Addr.ToString()

		tp types.TransactionType = types.CreateLeague
	)
	/*TransferBox
	TransferEnergy
	CreateLeague
	JoinLeague
	EnergyToLeague */
	//nonce, leagueNonce uint64, from, to, league string, tp maintypes.TransactionType)
	tx := transactionBuild(nonce, leagueNonce, from, to, leagueId, tp)
	err := tx.ToRaw()
	if err != nil {
		panic(err)
	}
	req := req.ReqSendTx{
		Raw: tx.Raw,
	}

	leagueId := types.ToLeagueAddress(tx)
	fmt.Printf("leagueId=%s\n", leagueId.ToString())

	fullurl := fmt.Sprintf("%s/block/sendtx", baseurl)
	resp, err := cli.HttpSend(fullurl, req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("txhash=%s,\n", resp)
}

func TestSubDynamicTxSend(t *testing.T) {
	var (
		nonce     uint64                   = 2
		from                               = mock.AccountNormal_1.Addr.ToString()
		to                                 = "ct0fHcR19s2nDaHxR3CVdUyYxHgsjzp8wrt"
		tp        orgtypes.TransactionType = orgtypes.TransferUT
		voteReply                          = orgtypes.Agree
		txhash                             = "52fcc88ab3521b1fbcc0cf2e7132cb67540c3826b390dcbc0fa6c7d744c7d149"
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

	//fmt.Println("hash ", tx.Hash())
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
	fullurl := fmt.Sprintf("%s/orgapi/sendsubtx", baseurl)
	resp, err := cli.HttpSend(fullurl, req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Logf("txhash=%s,\n", resp)
}

func TestRegistGetTxs(t *testing.T) {

	fullurl := fmt.Sprintf("%s/block/gettxs", baseurl)

	official, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")

	req := req.ReqGetTxs{
		Addr: official.ToString(),
	}

	reqbuf, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buffer := bytes.NewBuffer(reqbuf)
	request, err := http.NewRequest("GET", fullurl, buffer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("an error occured when ioutil.readall,err-->%v\n", err)
		return
	}

	t.Logf("resp:%s\n", buf)
}

func TestAddSyncOrg(t *testing.T) {

	fullurl := fmt.Sprintf("%s/block/syncorgdata", baseurl)

	leagueId := "ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda"

	req := req.ReqSyncOrgData{
		LeagueId: leagueId,
		BAdd:     true,
	}

	reqbuf, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buffer := bytes.NewBuffer(reqbuf)
	request, err := http.NewRequest("GET", fullurl, buffer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("an error occured when ioutil.readall,err-->%v\n", err)
		return
	}

	t.Logf("resp:%s\n", buf)
}

func TestSendSubTx(t *testing.T) {

	fullurl := fmt.Sprintf("%s/block/sendsubtx", baseurl)

	official := mock.Mock_Address_1
	txins := &common.TxIns{}
	txin := common.TxIn{

		Address: official,
		Amount:  big.NewInt(100),
		Nonce:   3,
	}
	txins.Tis = append(txins.Tis, &txin)

	to, _ := common.ToAddress("ct04xYt91DuMWGB7W1FhGbF3UfY7akq4bf9")
	totxins := &common.TxOuts{}
	totxin := common.TxOut{
		Address: to,
		Amount:  big.NewInt(100),
	}
	totxins.Tos = append(totxins.Tos, &totxin)

	msg, _ := common.StringToHash("hello")
	txdata := orgtypes.TxData{

		From:  official,
		Froms: txins,
		Tos:   totxins,
		Msg:   msg,
		Fee:   big.NewInt(1001),
		Nonce: 1,
	}

	tx, err := orgtypes.NewTransactionRaw(orgtypes.TransferUT, types.TxVersion, &txdata)
	err = tx.ToRaw()
	if err != nil {
		t.Errorf("toraw err-->%v\n", err)
		return
	}

	req := req.ReqSendTx{
		LeagueId: "ct2b5xc5uyaJrSqZ866gVvM5QtXybprb9sy",
		Raw:      tx.Raw,
	}

	reqbuf, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	buffer := bytes.NewBuffer(reqbuf)
	request, err := http.NewRequest("POST", fullurl, buffer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("an error occured when ioutil.readall,err-->%v\n", err)
		return
	}

	respobj := new(txpool.TxResp)
	err = json.Unmarshal(buf, respobj)

	if err != nil {
		t.Errorf("err=%v\n", err)
		return
	}

	t.Logf("txhash=%s,error info=%s\n", respobj.Hash, respobj.Desc)
}

func TestTime(t *testing.T)  {

	ok := Ok{Tm:time.Now()}

	okbuf,_ := json.Marshal(ok)
	fmt.Printf("okbuf=%s\n",okbuf)



}

type Ok struct {
	Tm time.Time `json:"tm"`
}
