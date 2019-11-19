package mock

import (
	"math/big"
	"time"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
	maintypes "github.com/sixexorg/magnetic-ring/core/mainchain/types"
	orgtypes "github.com/sixexorg/magnetic-ring/core/orgchain/types"
	"github.com/sixexorg/magnetic-ring/log"
)

func init() {
	InitTransactions()
	InitBlocks()
	InitLeagueDataAndConfirmTx()
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StderrHandler))
}

var (
	OrgTimeSpan    = uint64(time.Date(2018, 12, 13, 0, 0, 0, 0, time.Local).UnixNano())
	CreateLeagueTx *maintypes.Transaction
	Address_1, _   = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	Address_2, _   = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdb")
	Address_3, _   = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdc")
	Address_4, _   = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdd")
	Address_5, _   = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfde")
	/*txHash_1, _    = common.StringToHash("53m4x7z349sjd857zwpfu2d8e1v6rgh6jk1yg9t7fq7vxds6rnx1====")
	txHash_2, _    = common.StringToHash("pevhqgd4827rg441y4g5bwxx32zexr5v5ryup5duzcpgs76pg5u1====")
	txHash_3, _    = common.StringToHash("u2kehc55estec4kt9dvzafnrnqwdc16pvxj893dw8fprfd3ty1vh====")
	txHash_4, _    = common.StringToHash("av362nhkksdpbd8w9dw9g24c92tqxuxu8nzy3umed8xy3ste7awh====")
	txHash_5, _    = common.StringToHash("9umzce1r3qxedj2qvs1ewtj4j1fjedwt6rems8rnvmbqqw78p821====")*/

	//for createLeagueTx create
	LeagueAddress1, _ = common.ToAddress("ct2hA77vSd7JFUU71ZVtjcx3ttnh5d654qH")

	//mainchain txs,mock mainchain tx
	TxList                  map[common.Hash]*maintypes.Transaction
	Tx1, Tx2, Tx3, Tx4, Tx5 *maintypes.Transaction

	OrgTx1            *orgtypes.Transaction
	CreateLeagueTxMap map[common.Address]*maintypes.Transaction

	Block1 *orgtypes.Block
	Block2 *orgtypes.Block
	Block3 *orgtypes.Block
	Block4 *orgtypes.Block
	Block5 *orgtypes.Block
	Block6 *orgtypes.Block
	Block7 *orgtypes.Block

	Blocks []*orgtypes.Block

	NodeId, _ = common.ToAddress("ct3qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	///////for league
	//LeagueData1 *league_data.LeagueData
	ConfirmTx *maintypes.Transaction
)

func InitLeagueDataAndConfirmTx() {
	hashes := make(common.HashArray, 0, 2)
	hashes = append(hashes, Block1.Hash(), Block2.Hash())
	headers := make([]*orgtypes.Header, 0, 2)
	txMap := make(map[uint64]common.HashArray)
	for k, v := range Blocks[:2] {
		headers = append(headers, Blocks[k].Header)
		txMap[v.Header.Height] = make(common.HashArray, 0, v.Transactions.Len())
		for _, vi := range v.Transactions {
			txMap[v.Header.Height] = append(txMap[v.Header.Height], vi.Hash())
		}

	}
	//LeagueData1 = league_data.NewLeagueData(0, 1, hashes.GetHashRoot(), headers, txMap)

	ConfirmTx = &maintypes.Transaction{
		Version: 0x01,
		TxType:  maintypes.ConsensusLeague,
		TxData: &maintypes.TxData{
			StartHeight: 1,
			EndHeight:   2,
			BlockRoot:   hashes.GetHashRoot(),
			LeagueId:    LeagueAddress1,
			Fee:         big.NewInt(10),
		},
	}
}
func InitBlocks() {
	var (
		otxs                         orgtypes.Transactions
		otx1, otx2, otx3, otx4, otx5 *orgtypes.Transaction
	)

	Block1 = &orgtypes.Block{
		Header: &orgtypes.Header{
			Version:       0x01,
			PrevBlockHash: common.Hash{},
			LeagueId:      LeagueAddress1,
			TxRoot:        common.Hash{},
			StateRoot:     common.Hash{},
			ReceiptsRoot:  common.Hash{},
			Difficulty:    big.NewInt(10),
			Height:        1,
			Timestamp:     OrgTimeSpan,
		},
		Transactions: nil,
	}

	otx1 = &orgtypes.Transaction{
		Version: 0x01,
		TxType:  orgtypes.EnergyFromMain,
		TxData: &orgtypes.TxData{
			TxHash: Tx1.Hash(),
			From:   Tx1.TxData.From,
			Energy:   Tx1.TxData.Energy,
		},
	}
	OrgTx1 = otx1
	common.DeepCopy(&otx2, otx1)
	otx2.TxType = orgtypes.Join
	otx2.TxData.TxHash = Tx2.Hash()
	otx2.TxData.From = Tx2.TxData.From

	common.DeepCopy(&otx3, otx1)
	otx3.TxType = orgtypes.Join
	otx3.TxData.TxHash = Tx3.Hash()
	otx3.TxData.From = Tx3.TxData.From

	common.DeepCopy(&otx4, otx1)
	otx4.TxType = orgtypes.EnergyFromMain
	otx4.TxData.TxHash = Tx4.Hash()
	otx4.TxData.From = Tx4.TxData.From
	otx4.TxData.Energy = Tx4.TxData.Energy

	common.DeepCopy(&otx5, otx1)
	otx5.TxType = orgtypes.EnergyFromMain
	otx5.TxData.TxHash = Tx5.Hash()
	otx5.TxData.From = Tx5.TxData.From
	otx5.TxData.Energy = Tx5.TxData.Energy

	otxs = append(otxs, otx1, otx2, otx3, otx4, otx5)
	//receiptRoot :=
	//stateRoot :=

	Block2 = &orgtypes.Block{
		Header: &orgtypes.Header{
			Version:       0x01,
			PrevBlockHash: Block1.Hash(),
			LeagueId:      LeagueAddress1,
			TxRoot:        otxs.GetHashRoot(),
			StateRoot:     common.Hash{},
			ReceiptsRoot:  common.Hash{},
			Difficulty:    big.NewInt(10),
			Height:        2,
			Timestamp:     uint64(time.Now().Unix()),
		},
		Transactions: otxs,
	}

	Blocks = append(Blocks, Block1, Block2, Block3, Block4, Block5, Block6, Block7)
}
func InitTransactions() {
	TxList = make(map[common.Hash]*maintypes.Transaction)
	CreateLeagueTxMap = make(map[common.Address]*maintypes.Transaction)
	CreateLeagueTx = &maintypes.Transaction{
		Version: 0x01,
		TxType:  maintypes.CreateLeague,
		TxData: &maintypes.TxData{
			From:    Address_1,
			Nonce:   1,
			Fee:     big.NewInt(10),
			Rate:    1000,
			MinBox:  10,
			MetaBox: big.NewInt(2000),
			NodeId:  NodeId,
			Private: false,
		},
	}

	CreateLeagueTxMap[ToLeagueAddress(CreateLeagueTx)] = CreateLeagueTx
	var (
		maintxs maintypes.Transactions
	)

	txd1 := &maintypes.TxData{
		LeagueId: LeagueAddress1,
		Fee:      big.NewInt(10),
		Energy:     big.NewInt(2000),
		From:     Address_1,
		Nonce:    1,
	}
	Tx1 = &maintypes.Transaction{
		Version: 0x01,
		TxType:  maintypes.EnergyToLeague,
		TxData:  txd1,
	}

	common.DeepCopy(&Tx2, Tx1)
	Tx2.TxType = maintypes.JoinLeague
	Tx2.TxData.LeagueNonce = 10
	Tx2.TxData.From = Address_2
	Tx2.TxData.Account = Address_2
	Tx2.TxData.Nonce = 1
	Tx2.TxData.MinBox = CreateLeagueTx.TxData.MinBox

	common.DeepCopy(&Tx3, Tx1)
	Tx3.TxType = maintypes.JoinLeague
	Tx3.TxData.LeagueNonce = 11
	Tx3.TxData.From = Address_3
	Tx3.TxData.Account = Address_3
	Tx3.TxData.Nonce = 1
	Tx3.TxData.MinBox = CreateLeagueTx.TxData.MinBox

	common.DeepCopy(&Tx4, Tx1)
	Tx4.TxType = maintypes.EnergyToLeague
	Tx4.TxData.LeagueNonce = 11
	Tx4.TxData.From = Address_2
	Tx4.TxData.Nonce = 2
	Tx4.TxData.Energy = big.NewInt(1500)

	common.DeepCopy(&Tx5, Tx1)
	Tx5.TxType = maintypes.EnergyToLeague
	Tx5.TxData.LeagueNonce = 12
	Tx5.TxData.From = Address_2
	Tx5.TxData.Nonce = 3
	Tx5.TxData.Energy = big.NewInt(2000)

	maintxs = append(maintxs, Tx1, Tx2, Tx3, Tx4, Tx5)
	for k, v := range maintxs {
		TxList[v.Hash()] = maintxs[k]
	}
	/*
		53m4x7z349sjd857zwpfu2d8e1v6rgh6jk1yg9t7fq7vxds6rnx1====
		pevhqgd4827rg441y4g5bwxx32zexr5v5ryup5duzcpgs76pg5u1====
		u2kehc55estec4kt9dvzafnrnqwdc16pvxj893dw8fprfd3ty1vh====
		av362nhkksdpbd8w9dw9g24c92tqxuxu8nzy3umed8xy3ste7awh====
		9umzce1r3qxedj2qvs1ewtj4j1fjedwt6rems8rnvmbqqw78p821====
	*/
}

func GetTransaction(txId common.Hash) (*maintypes.Transaction, uint64, error) {
	return TxList[txId], 0, nil
}

func ToLeagueAddress(tx *maintypes.Transaction) common.Address {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteAddress(tx.TxData.From)
	sk.WriteAddress(tx.TxData.NodeId)
	cp, _ := sink.BigIntToComplex(tx.TxData.MetaBox)
	sk.WriteComplex(cp)
	sk.WriteUint64(tx.TxData.MinBox)
	sk.WriteUint32(tx.TxData.Rate)
	sk.WriteBool(tx.TxData.Private)
	league, _ := common.CreateLeagueAddress(sk.Bytes())
	return league
}
