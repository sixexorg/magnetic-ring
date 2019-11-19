package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"github.com/sixexorg/magnetic-ring/account"
	"github.com/sixexorg/magnetic-ring/common"
	pasw "github.com/sixexorg/magnetic-ring/common/password"
	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/gomobile/walletutil"
	httpcli "github.com/sixexorg/magnetic-ring/http/cli"
	"github.com/sixexorg/magnetic-ring/http/req"
	"github.com/sixexorg/magnetic-ring/signer/mainchain/signature"
	orgsignature "github.com/sixexorg/magnetic-ring/signer/orgchain/signature"
)

var (
	SignCommand = cli.Command{
		Action:      cli.ShowSubcommandHelp,
		Name:        "tx",
		Usage:       "about transaction",
		ArgsUsage:   "[arguments...]",
		Description: `transaction actions`,
		Subcommands: []cli.Command{
			{
				Action:    signRawTransaction,
				Name:      "sign",
				Usage:     "sign a transaction",
				ArgsUsage: "[sub-command options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "account,a",
						Value: "",
						Usage: "sign transaction with account",
					},
				},
				Description: ` Add a new account to keystore.`,
			},
			{
				Action:    signOrgRawTransaction,
				Name:      "orgsign",
				Usage:     "sign an org transaction",
				ArgsUsage: "[sub-command options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "account,a",
						Value: "",
						Usage: "sign transaction with account",
					},
				},
				Description: ` Add a new account to keystore.`,
			},
			{
				Action:    pushTransaction,
				Name:      "pushtx",
				Usage:     "push a transaction to block chain network",
				ArgsUsage: "[sub-command options]",
				Flags:     []cli.Flag{
					//cli.StringFlag{
					//	Name:  "account,a",
					//	Value: "",
					//	Usage: "sign transaction with account",
					//},
				},
				Description: ` Add a new account to keystore.`,
			},
			{
				Action:    createMsgTx,
				Name:      "createmsgtx",
				Usage:     "createmsgtx -a=addr -orgaddr=orgaddr",
				ArgsUsage: "[sub-command options]",
				Flags:     []cli.Flag{
					cli.StringFlag{
						Name:  "account,a",
						Value: "",
						Usage: "who create this tx",
					},
					cli.StringFlag{
						Name:  "orgaddr",
						Value: "",
						Usage: "which org to create this tx",
					},
					cli.StringFlag{
						Name:  "pubk",
						Value: "",
						Usage: "your account pubk",
					},
					cli.StringFlag{
						Name:  "to",
						Value: "",
						Usage: "point an address you want to tell hims",
					},
					cli.StringFlag{
						Name:  "msg",
						Value: "",
						Usage: "message,can be a word,a sentence or a hash",
					},
				},
				Description: "create a tx for sending msg or hash to chain",
			},
		},
	}
)

func signRawTransaction(ctx *cli.Context) {
	password, err := pasw.GetPassword()
	if err != nil {
		fmt.Printf("input password error:%v\n", err)
		return
	}

	mgr, err := account.NewManager()
	if err != nil {
		fmt.Printf("create wallet error:%v\n", err)
		return
	}

	acctaddressstr := ctx.String("account")

	comaddress, err := common.ToAddress(acctaddressstr)
	if err != nil {
		fmt.Printf("toaddress error=%v\n",err)
		return
	}

	passwordstr := string(password)

	rawTx := ctx.Args().First()
	txbuf, err := common.Hex2Bytes(rawTx)
	if err != nil {
		fmt.Printf("hex2bytes error =%v\n", err)
		return
	}
	buff := bytes.NewBuffer(txbuf)
	tx := &types.Transaction{}
	sk := sink.NewZeroCopySource(buff.Bytes())
	err = tx.Deserialization(sk)
	if err != nil {
		fmt.Printf("can not deserialize data to transaction,err=%v\n", err)
		return
	}

	err = signature.SignTransaction(mgr, comaddress, tx, passwordstr)
	if err != nil {
		fmt.Printf("sign transaction error=%v\n", err)
		return
	}

	b, e := tx.VerifySignature()
	if e != nil {
		fmt.Printf("error=%v\n", e)
		return
	}

	fmt.Printf("b=%t\n", b)

	serbufer := new(bytes.Buffer)
	err = tx.Serialize(serbufer)

	if err != nil {
		fmt.Printf("Serialize transaction error=%v\n", err)
		return
	}

	outbuf := serbufer.Bytes()

	returnstr := common.Bytes2Hex(outbuf)

	reverse, err := common.Hex2Bytes(returnstr)
	if err != nil {
		fmt.Printf("hex2bytes error =%v\n", err)
		return
	}

	txrev := new(types.Transaction)

	source := sink.NewZeroCopySource(reverse)
	err = txrev.Deserialization(source)
	if err != nil {
		fmt.Printf("transaction Deserialization error=", err)
		return
	}

	b2, e2 := txrev.VerifySignature()
	if e2 != nil {
		fmt.Printf("e2=%v\n", e2)
		return
	}

	fmt.Printf("b2=%v\n", b2)

	fmt.Printf("%s\n", returnstr)

}

func pushTransaction(ctx *cli.Context) {
	baseurl := fmt.Sprintf("http://127.0.0.1:%d", config.GlobalConfig.SysCfg.HttpPort)
	rawTx := ctx.Args().First()
	ra, err := common.Hex2Bytes(rawTx)
	if err != nil {
		fmt.Printf("hex2bytes error =%v\n", err)
		return
	}
	req := req.ReqSendTx{
		Raw: ra,
	}
	fullurl := fmt.Sprintf("%s/block/sendtx", baseurl)
	resp, err := httpcli.HttpSend(fullurl, req)
	if err != nil {
		fmt.Printf("push transaction to net work error:=%v\n", err)
		return
	}
	fmt.Printf("%s\n", resp)
}


func getNonceInLeague(orgid,account string) (string,error){
	baseurl := fmt.Sprintf("http://127.0.0.1:%d", config.GlobalConfig.SysCfg.HttpPort)


	req := req.ReqUngetBonus{
		OrgId:orgid,
		Account:account,
	}
	fullurl := fmt.Sprintf("%s/orgapi/getleaguenonce", baseurl)
	return httpcli.HttpSend(fullurl, req)
	//if err != nil {
	//	//fmt.Printf("push transaction to net work error:=%v\n", err)
	//	return resp,err
	//}
	//fmt.Printf("%s\n", resp)
}


func createMsgTx(ctx *cli.Context) {

	//orgid := ctx.String("orgaddr")
	//account := ctx.String("account")

	pubkstr := ctx.String("pubk")
	tostr := ctx.String("to")
	msgstr := ctx.String("msg")
	//feestr := ctx.Int64("fee")

	//resp,err := getNonceInLeague(orgid,account)
	//if err != nil {
	//	fmt.Printf("push transaction to net work error:=%v\n", err)
	//	return
	//}
	//
	type Noncer struct {
		Nonce uint64 `json:"nonce"`
	}
	nc := new(Noncer)
	nc.Nonce=1
	//
	//err = json.Unmarshal([]byte(resp),nc)
	//if err != nil {
	//	fmt.Printf("decode nonce error:=%v\n", err)
	//	return
	//}


	txhash := walletutil.PushMsg(pubkstr,tostr,msgstr,10000000000,int64(nc.Nonce+1))

	fmt.Printf("%s\n", txhash)
}

func createMsgTx2(ctx *cli.Context) {

	orgid := ctx.String("orgaddr")
	account := ctx.String("account")

	pubkstr := ctx.String("pubk")
	tostr := ctx.String("to")
	msgstr := ctx.String("msg")
	//feestr := ctx.Int64("fee")

	resp,err := getNonceInLeague(orgid,account)
	if err != nil {
		fmt.Printf("push transaction to net work error:=%v\n", err)
		return
	}

	type Noncer struct {
		Nonce uint64 `json:"nonce"`
	}
	nc := new(Noncer)

	err = json.Unmarshal([]byte(resp),nc)
	if err != nil {
		fmt.Printf("decode nonce error:=%v\n", err)
		return
	}


	txhash := walletutil.PushMsg(pubkstr,tostr,msgstr,10000000000,int64(nc.Nonce+1))

	fmt.Printf("%s\n", txhash)
}


func signOrgRawTransaction(ctx *cli.Context) {
	password, err := pasw.GetPassword()
	if err != nil {
		fmt.Printf("input password error:%v\n", err)
		return
	}

	mgr, err := account.NewManager()
	if err != nil {
		fmt.Printf("create wallet error:%v\n", err)
		return
	}

	acctaddressstr := ctx.String("account")

	comaddress, err := common.ToAddress(acctaddressstr)
	if err != nil {
		fmt.Printf("toaddress error=%v\n",err)
		return
	}

	passwordstr := string(password)

	rawTx := ctx.Args().First()
	txbuf, err := common.Hex2Bytes(rawTx)
	if err != nil {
		fmt.Printf("hex2bytes error =%v\n", err)
		return
	}
	buff := bytes.NewBuffer(txbuf)
	tx := &orgtypes.Transaction{}
	sk := sink.NewZeroCopySource(buff.Bytes())
	err = tx.Deserialization(sk)
	if err != nil {
		fmt.Printf("can not deserialize data to transaction,err=%v\n", err)
		return
	}

	err = orgsignature.SignTransaction(mgr, comaddress, tx, passwordstr)
	if err != nil {
		fmt.Printf("sign transaction error 777 =%v\n", err)
		return
	}

	_, e := tx.VerifySignature()
	if e != nil {
		fmt.Printf("error=%v\n", e)
		return
	}

	//fmt.Printf("verify signature err=%t\n",b)

	serbufer := new(bytes.Buffer)
	err = tx.Serialize(serbufer)

	if err != nil {
		fmt.Printf("Serialize transaction error=%v\n", err)
		return
	}

	outbuf := serbufer.Bytes()

	returnstr := common.Bytes2Hex(outbuf)

	reverse, err := common.Hex2Bytes(returnstr)
	if err != nil {
		fmt.Printf("hex2bytes error =%v\n", err)
		return
	}

	txrev := new(orgtypes.Transaction)

	source := sink.NewZeroCopySource(reverse)
	err = txrev.Deserialization(source)
	if err != nil {
		fmt.Printf("transaction Deserialization error=", err)
		return
	}

	_, e2 := txrev.VerifySignature()
	if e2 != nil {
		fmt.Printf("e2=%v\n", e2)
		return
	}

	//fmt.Printf("b2=%v\n", b2)

	fmt.Printf("%s\n", returnstr)

}