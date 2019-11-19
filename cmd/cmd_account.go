package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/sixexorg/magnetic-ring/account"
	pasw "github.com/sixexorg/magnetic-ring/common/password"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/crypto"
	httpcli "github.com/sixexorg/magnetic-ring/http/cli"
	"github.com/sixexorg/magnetic-ring/http/req"
)

var (
	AccountCommand = cli.Command{
		Action:    cli.ShowSubcommandHelp,
		Name:      "account",
		Usage:     "Manage accounts",
		ArgsUsage: "[arguments...]",
		Description: `Wallet management commands can be used to add, view, modify, delete, import account, and so on.
					  You can use ./go-crystal account --help command to view help information of wallet management command.`,
		Subcommands: []cli.Command{
			{
				Action:    accountCreate,
				Name:      "add",
				Usage:     "Add a new account",
				ArgsUsage: "[sub-command options]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "type,t",
						Value: "normal",
						Usage: "account type:[normal,node]",
					},
				},
				Description: ` Add a new account to keystore.`,
			},
			{
				Action:      getBoxBalance,
				Name:        "getbalance",
				Usage:       "get balance",
				ArgsUsage:   "[sub-command options]",
				Flags:       []cli.Flag{},
				Description: ` Add a new account to keystore.`,
			},
			{
				Action:    oneKeyAccountCreate,
				Name:      "create",
				Usage:     "create a new account",
				ArgsUsage: "[sub-command options]",
				Flags:     []cli.Flag{
					cli.StringFlag{
						Name:  "dir,d",
						Value: "normal",
						Usage: "point keystore file path",
					},
				},
				Description: ` create a new account to keystore.`,
			},
		},
	}
)

func accountCreate(ctx *cli.Context) {
	password, err := pasw.GetConfirmedPassword()
	if err != nil {
		fmt.Printf("input password error:%v\n", err)
		return
	}

	mgr, err := account.NewManager()
	if err != nil {
		fmt.Printf("create wallet error:%v\n", err)
		return
	}

	accountType := ctx.String("type")

	var acct account.NormalAccount

	if accountType == "normal" {
		acct, err = mgr.GenerateNormalAccount(string(password))
		if err != nil {
			fmt.Printf("create normal account error:%v\n", err)
			return
		}
	} else if accountType == "node" {
		acct, err = mgr.GenerateNodeAccount(string(password))
		if err != nil {
			fmt.Printf("create node account error:%v\n", err)
			return
		}
	} else {
		fmt.Printf("only [normal,node] account support")
		return
	}

	mult := account.MultiAccountUnit{1, []crypto.PublicKey{acct.PublicKey()}}

	muls := make(account.MultiAccountUnits, 1)
	muls[0] = mult

	mulacct, err := mgr.CreateMultipleAccount(acct.Address(), muls)

	if err != nil {
		return
	}

	pubksrcbuf := acct.PublicKey().Bytes()

	fmt.Printf("public key:%x\n", pubksrcbuf)
	fmt.Printf("address:%s\n", acct.Address().ToString())
	fmt.Printf("default multiple address:%s\n", mulacct.Address().ToString())

}

func oneKeyAccountCreate(ctx *cli.Context) {
	password, err := pasw.GetConfirmedPassword()
	if err != nil {
		fmt.Printf("input password error:%v\n", err)
		return
	}

	mgr, err := account.NewManager()
	if err != nil {
		fmt.Printf("create wallet error:%v\n", err)
		return
	}

	multact, err := mgr.CreateOneKeyAccount(password)
	if err != nil {
		fmt.Printf("create account error=%v\n", err)
		return
	}

	impl := multact.(*account.MultipleAccountImpl)

	pubksrcbuf := impl.OnlyPubkey().Bytes()

	fmt.Printf("public key:%x\n", pubksrcbuf)
	fmt.Printf("address:%s\n", multact.Address().ToString())
	//fmt.Printf("default multiple address:%s\n", mulacct.Address().ToString())

}

func getBoxBalance(ctx *cli.Context) {
	baseurl := fmt.Sprintf("http://127.0.0.1:%d", config.GlobalConfig.SysCfg.HttpPort)
	addr := ctx.Args().First()
	orgId := ctx.Args().Get(1)
	req := req.ReqGetBalance{
		orgId,
		addr,
	}
	fullurl := fmt.Sprintf("%s/block/getbalance", baseurl)
	resp, err := httpcli.HttpSend(fullurl, req)
	if err != nil {
		fmt.Printf("push transaction to net work error:=%v\n", err)
		return
	}
	fmt.Printf("%s\n", resp)

}
