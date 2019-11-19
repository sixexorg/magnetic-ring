package genesis

import (
	"math/big"

	"encoding/json"

	"encoding/hex"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
)

type Genesis struct {
	Config *config.GenesisConfig
}

var genesis *Genesis

func InitGenesis(conf *config.GenesisConfig) (*Genesis, error) {
	genesis = &Genesis{
		Config: conf,
	}
	return genesis, nil
}
func GetGenesisBlockState() (states.AccountStates, types.Transactions) {
	return genesis.genFirst()
}

type VbftBlockInfo struct {
	VrfValue []byte `json:"vrf_value"`
	VrfProof []byte `json:"vrf_proof"`
}

//privatekey must input from cmd
func (g *Genesis) GenesisBlock(ledger *storages.LedgerStoreImp) error {
	ass, txs := g.genFirst()
	//todo hello world msg
	block := &types.Block{
		Header: &types.Header{
			Height:        1,
			Version:       0x01,
			PrevBlockHash: common.Hash{},
			LeagueRoot:    common.Hash{},
			ReceiptsRoot:  common.Hash{},
			StateRoot:     ass.GetHashRoot(),
			TxRoot:        txs.GetHashRoot(),
			Timestamp:     g.Config.Timestamp,
		},
		Transactions: txs,
	}
	blkInfo := &VbftBlockInfo{}
	blkInfo.VrfValue = []byte("1c9810aa9822e511d5804a9c4db9dd08497c31087b0daafa34d768a3253441fa20515e2f30f81741102af0ca3cefc4818fef16adb825fbaa8cad78647f3afb590e")
	blkInfo.VrfProof = []byte("c57741f934042cb8d8b087b44b161db56fc3ffd4ffb675d36cd09f83935be853d8729f3f5298d12d6fd28d45dde515a4b9d7f67682d182ba5118abf451ff1988")
	block.Header.ConsensusPayload, _ = json.Marshal(blkInfo)
	blockInfo := &storages.BlockInfo{
		Block:         block,
		AccountStates: ass,
		GasDestroy:    big.NewInt(0),
	}
	err := ledger.SaveAll(blockInfo)
	return err
}

func (g *Genesis) genFirst() (states.AccountStates, types.Transactions) {
	official, _ := common.ToAddress(g.Config.Official)
	ass := make(states.AccountStates, 0, 1+len(g.Config.Stars))
	txs := make(types.Transactions, 0, len(g.Config.Stars))
	as := &states.AccountState{
		Address: official,
		Height:  1,
		Data: &states.Account{
			Nonce:       1,
			Balance:     big.NewInt(0).SetUint64(g.Config.crystal),
			EnergyBalance: big.NewInt(0).SetUint64(g.Config.Energy),
		},
	}
	ass = append(ass, as)
	for _, v := range g.Config.Stars {

		addr, err := common.ToAddress(v.Account)
		if err != nil {
			panic(err)
		}
		asNode := &states.AccountState{
			Address: addr,
			Height:  1,
			Data: &states.Account{
				Nonce:   1,
				Balance: big.NewInt(0).SetUint64(v.crystal),
			},
		}
		ass = append(ass, asNode)

		pubuf, err := hex.DecodeString(v.Nodekey)
		if err != nil {
			continue
		}
		tx := &types.Transaction{

			Version: types.TxVersion,
			TxType:  types.AuthX,
			TxData: &types.TxData{
				Nonce:   1,
				From:    addr,
				NodePub: pubuf,
				Fee:     big.NewInt(0),
			},
		}
		txs = append(txs, tx)
	}
	return ass, txs
}
