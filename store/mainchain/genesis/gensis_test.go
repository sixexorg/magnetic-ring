package genesis

import (
	"testing"

	"os"

	"fmt"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	_ "github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

func TestGenesis(t *testing.T) {
	dbDir := "test"
	defer os.RemoveAll(dbDir)
	ledger, _ := storages.NewLedgerStore("dbDir")
	genconf := &Genesis{
		Config: &config.GenesisConfig{
			ChainId:   "1",
			crystal:       10000000,
			Energy:      20000000,
			Timestamp: 12345678,
			Official:  "04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418e",
			Stars: []*config.StarNode{
				{
					Account: common.Address{1, 2, 3}.ToString(),
					crystal:     20000,
					Nodekey: "04dc2c38fd4985a30f31fe9ab0e1d6ffa85d33d5c8f0a5ff38c45534e8f4ffe751055e6f1bbec984839dd772a68794722683ed12994feb84a13b8086162591418f",
				},
			},
		},
	}
	genconf.GenesisBlock(ledger)

	ass, _ := genconf.genFirst()
	ast, err := ledger.GetAccountByHeight(2, ass[0].Address)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log(ast.Data.Balance)
	block, err := ledger.GetBlockByHeight(1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	fmt.Println(block)
}
