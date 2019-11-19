package validation_test

import (
	"os"
	"testing"

	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/config"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/genesis"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/mainchain/validation"

	_ "github.com/sixexorg/magnetic-ring/mock"
)

var (
	stateValidate *validation.StateValidate
	tx1           *types.Transaction //crystal
	tx2           *types.Transaction //crystal
	tx3           *types.Transaction //energy
	tx4           *types.Transaction //crystal
	tx5           *types.Transaction //crystal

	tx6 *types.Transaction //createLeague

	tx7  *types.Transaction //createLeague
	tx8  *types.Transaction //createLeague
	tx9  *types.Transaction //createLeague
	tx10 *types.Transaction //join
	tx11 *types.Transaction //join
	tx12 *types.Transaction //join

	txs []*types.Transaction

	nodeId, _    = common.ToAddress("ct3qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	Address_1, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	Address_2, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdb")
	Address_3, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdc")
	Address_4, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdd")
	Address_5, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfde")

	League_1, _                                                        = common.ToAddress("ct2qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	League_1_minbox, league_1_frozenBox, league_1_nonce, league_1_rate = uint64(20), big.NewInt(2000), uint64(1), uint32(1000)

	ledger *storages.LedgerStoreImp

	gss *genesis.Genesis
)

func TestMain(m *testing.M) {

	dbdir := "test"
	ledger, _ = storages.NewLedgerStore(dbdir)
	genconf := &config.GenesisConfig{
		ChainId: "1",
		Crystal:     10000000,
		Energy:    20000000,
	}
	gss, _ = genesis.InitGenesis(genconf)
	err := gss.GenesisBlock(ledger)
	if err != nil {
		return
	}
	stateValidate = validation.NewStateValidate(ledger)

	//height := uint64(50)
	var (
		tp1, tp2, tp3, tp4, tp5, tp6    = types.TransferBox, types.TransferBox, types.TransferEnergy, types.TransferBox, types.TransferBox, types.CreateLeague
		tp7, tp8, tp9, tp10, tp11, tp12 = types.CreateLeague, types.CreateLeague, types.CreateLeague, types.JoinLeague, types.JoinLeague, types.JoinLeague

		balance1, balance2                                = big.NewInt(10000), big.NewInt(2000)
		energy1, energy2                                      = big.NewInt(5000), big.NewInt(4000)
		nonce1, n1, n2, n3, n4, n5, n6             uint64 = 1, 2, 3, 4, 5, 6, 7
		nonce2, m1, m2, m3, m4, m5                 uint64 = 1, 2, 3, 4, 5, 6
		fee1, fee2, fee3, fee4, fee5, fee6                = big.NewInt(20), big.NewInt(10), big.NewInt(2), big.NewInt(20), big.NewInt(5), big.NewInt(100)
		tx1a1, tx1a2, touta1_1, touta1_2, touta1_3        = big.NewInt(20), big.NewInt(108), big.NewInt(30), big.NewInt(50), big.NewInt(28)
		tx2a1, tx2a2, touta2_1, touta2_2, touta2_3        = big.NewInt(80), big.NewInt(300), big.NewInt(50), big.NewInt(50), big.NewInt(280)
		tx3a1, tx3a2, touta3_1, touta3_2, touta3_3        = big.NewInt(800), big.NewInt(500), big.NewInt(500), big.NewInt(300), big.NewInt(500)
		tx4a1, tx4a2, touta4_1, touta4_2, touta4_3        = big.NewInt(200), big.NewInt(20), big.NewInt(20), big.NewInt(100), big.NewInt(100)
		tx5a1, tx5a2, touta5_1, touta5_2, touta5_3        = big.NewInt(3000), big.NewInt(40), big.NewInt(1000), big.NewInt(1000), big.NewInt(1040)
		//addr1,addr1,addr2,addr1,addr1,addr2
		ln7, ln8, ln9, ln10, ln11, ln12       uint64 = 8, 9, 7, 10, 11, 8
		fee7, fee8, fee9, fee10, fee11, fee12        = big.NewInt(50), big.NewInt(50), big.NewInt(50), big.NewInt(50), big.NewInt(50), big.NewInt(50)
	)
	//league data
	var (
		rate6, rate7, rate8, rate9             uint32 = 1000, 2000, 2000, 4000
		minBox6, minBox7, minBox8, minBox9     uint64 = 100, 20, 20, 50
		metaBox6, metaBox7, metaBox8, metaBox9        = big.NewInt(1000), big.NewInt(1000), big.NewInt(1000), big.NewInt(1000)
		jminBox10, jminBox11, jminBox12        uint64 = 20, 20, 20
		leageNonce1, leageNonce2, leageNonce3  uint64 = 2, 2, 3
	)

	ass := make(states.AccountStates, 0, 5)
	ass = append(ass, &states.AccountState{
		Address: Address_1,
		Height:  0,
		Data: &states.Account{
			Nonce:       nonce1,
			Balance:     balance1,
			EnergyBalance: energy1,
			BonusHeight: 0,
		}}, &states.AccountState{
		Address: Address_2,
		Height:  0,
		Data: &states.Account{
			Nonce:       nonce2,
			Balance:     balance2,
			EnergyBalance: energy2,
			BonusHeight: 0,
		}}, &states.AccountState{
		Address: Address_3,
		Height:  0,
		Data: &states.Account{
			Nonce:       0,
			Balance:     big.NewInt(10000),
			EnergyBalance: big.NewInt(20000),
			BonusHeight: 0,
		}}, &states.AccountState{
		Address: Address_4,
		Height:  0,
		Data: &states.Account{
			Nonce:       0,
			Balance:     big.NewInt(0),
			EnergyBalance: big.NewInt(0),
			BonusHeight: 0,
		}}, &states.AccountState{
		Address: Address_5,
		Height:  0,
		Data: &states.Account{
			Nonce:       0,
			Balance:     big.NewInt(0),
			EnergyBalance: big.NewInt(0),
			BonusHeight: 0,
		}},
	)
	ledger.SaveAccounts(ass)

	stateValidate.ParentBonusHeight = 1000
	stateValidate.MemoAccountState[Address_1] = ass[0]
	stateValidate.MemoAccountState[Address_2] = ass[1]
	stateValidate.MemoAccountState[Address_3] = ass[2]
	stateValidate.MemoAccountState[Address_4] = ass[3]
	stateValidate.MemoAccountState[Address_5] = ass[4]

	/*	league1 := &states.LeagueState{
			Address: League_1,
			Height:  0,
			MinBox:  League_1_minbox,
			Creator: Address_5,
			Rate:    league_1_rate,
			Data: &states.League{
				Nonce:     league_1_nonce,
				FrozenBox: league_1_frozenBox,
				Private:   false,
			},
		}
		member1 := &states.LeagueMember{
			LeagueId: League_1,
			Height:   0,
			Data: &states.LeagueAccount{
				Account: Address_5,
				Status:  states.LAS_Normal,
			},
		}
		ledger.SaveLeagues(states.LeagueStates{league1}, states.LeagueMembers{member1})*/

	//stateValidate.MemoLeagueState[League_1] = league1
	common.DeepCopy(&stateValidate.DirtyAccountState, stateValidate.MemoAccountState)
	common.DeepCopy(&stateValidate.DirtyLeagueState, stateValidate.MemoLeagueState)
	froms := &common.TxIns{}
	froms.Tis = append(froms.Tis,
		&common.TxIn{
			Address: Address_1,
			Nonce:   n1,
			Amount:  tx1a1,
		},
		&common.TxIn{
			Address: Address_2,
			Nonce:   m1,
			Amount:  tx1a2,
		},
	)
	tos := &common.TxOuts{}
	tos.Tos = append(tos.Tos,
		&common.TxOut{
			Address: Address_3,
			Amount:  touta1_1,
		},
		&common.TxOut{
			Address: Address_4,
			Amount:  touta1_2,
		},
		&common.TxOut{
			Address: Address_5,
			Amount:  touta1_3,
		},
	)
	txd1 := &types.TxData{
		Froms: froms,
		Tos:   tos,
		Fee:   fee1,
		From:  Address_1,
	}

	tx1 = &types.Transaction{
		Version: 0x01,
		TxType:  tp1,
		TxData:  txd1,
	}
	common.DeepCopy(&tx2, tx1)
	tx2.TxType = tp2
	tx2.TxData.Froms.Tis[0].Nonce = n2
	tx2.TxData.Froms.Tis[0].Amount = tx2a1
	tx2.TxData.Froms.Tis[1].Nonce = m2
	tx2.TxData.Froms.Tis[1].Amount = tx2a2
	tx2.TxData.Tos.Tos[0].Amount = touta2_1
	tx2.TxData.Tos.Tos[1].Amount = touta2_2
	tx2.TxData.Tos.Tos[2].Amount = touta2_3
	tx2.TxData.Fee = fee2
	common.DeepCopy(&tx3, tx1)
	tx3.TxType = tp3
	tx3.TxData.Froms.Tis[0].Nonce = n3
	tx3.TxData.Froms.Tis[0].Amount = tx3a1
	tx3.TxData.Froms.Tis[1].Nonce = m3
	tx3.TxData.Froms.Tis[1].Amount = tx3a2
	tx3.TxData.Tos.Tos[0].Amount = touta3_1
	tx3.TxData.Tos.Tos[1].Amount = touta3_2
	tx3.TxData.Tos.Tos[2].Amount = touta3_3
	tx3.TxData.Fee = fee3
	common.DeepCopy(&tx4, tx1)
	tx4.TxType = tp4
	tx4.TxData.Froms.Tis[0].Nonce = n4
	tx4.TxData.Froms.Tis[0].Amount = tx4a1
	tx4.TxData.Froms.Tis[1].Nonce = m4
	tx4.TxData.Froms.Tis[1].Amount = tx4a2
	tx4.TxData.Tos.Tos[0].Amount = touta4_1
	tx4.TxData.Tos.Tos[1].Amount = touta4_2
	tx4.TxData.Tos.Tos[2].Amount = touta4_3
	tx4.TxData.Fee = fee4
	common.DeepCopy(&tx5, tx1)
	tx5.TxType = tp5
	tx5.TxData.Froms.Tis[0].Nonce = n5
	tx5.TxData.Froms.Tis[0].Amount = tx5a1
	tx5.TxData.Froms.Tis[1].Nonce = m5
	tx5.TxData.Froms.Tis[1].Amount = tx5a2
	tx5.TxData.Tos.Tos[0].Amount = touta5_1
	tx5.TxData.Tos.Tos[1].Amount = touta5_2
	tx5.TxData.Tos.Tos[2].Amount = touta5_3
	tx5.TxData.Fee = fee5

	addr_a_bal := big.NewInt(0).Set(balance1)
	addr_a_bal.Sub(addr_a_bal, tx1a1)
	addr_a_bal.Sub(addr_a_bal, tx2a1)
	addr_a_bal.Sub(addr_a_bal, tx3a1)
	addr_a_bal.Sub(addr_a_bal, tx4a1)
	addr_a_bal.Sub(addr_a_bal, tx5a1)

	addr_b_bal := big.NewInt(0).Set(balance2)
	addr_b_bal.Sub(addr_b_bal, tx1a2)
	addr_b_bal.Sub(addr_b_bal, tx2a2)
	addr_b_bal.Sub(addr_b_bal, tx3a2)
	addr_b_bal.Sub(addr_b_bal, tx4a2)
	addr_b_bal.Sub(addr_b_bal, tx5a2)

	tx6 = &types.Transaction{
		Version: 0x01,
		TxType:  tp6,
		TxData: &types.TxData{
			From:    Address_1,
			Nonce:   n6,
			Fee:     fee6,
			Rate:    rate6,
			MinBox:  minBox6,
			MetaBox: metaBox6,
			NodeId:  nodeId,
			Private: false,
		},
	}
	common.DeepCopy(&tx7, tx6)
	tx7.TxType = tp7
	tx7.TxData.From = Address_1
	tx7.TxData.Nonce = ln7
	tx7.TxData.Rate = rate7
	tx7.TxData.MetaBox = metaBox7
	tx7.TxData.MinBox = minBox7
	tx7.TxData.Private = true
	tx7.TxData.Fee = fee7

	common.DeepCopy(&tx8, tx6)
	tx8.TxType = tp8
	tx8.TxData.From = Address_1
	tx8.TxData.Nonce = ln8
	tx8.TxData.Rate = rate8
	tx8.TxData.MetaBox = metaBox8
	tx8.TxData.MinBox = minBox8
	tx8.TxData.Private = true
	tx8.TxData.Fee = fee8

	common.DeepCopy(&tx9, tx6)
	tx9.TxType = tp9
	tx9.TxData.From = Address_2
	tx9.TxData.Nonce = ln9
	tx9.TxData.Rate = rate9
	tx9.TxData.MetaBox = metaBox9
	tx9.TxData.MinBox = minBox9
	tx9.TxData.Private = true
	tx9.TxData.Fee = fee9

	common.DeepCopy(&tx10, tx6)
	tx10.TxType = tp10
	tx10.TxData.From = Address_1
	tx10.TxData.Account = Address_1
	tx10.TxData.Nonce = ln10
	tx10.TxData.LeagueNonce = leageNonce1
	tx10.TxData.MinBox = jminBox10
	tx10.TxData.LeagueId = League_1
	tx10.TxData.Fee = fee10

	common.DeepCopy(&tx11, tx6)
	tx11.TxType = tp11
	tx11.TxData.From = Address_1
	tx11.TxData.Account = Address_1
	tx11.TxData.Nonce = ln11
	tx11.TxData.LeagueNonce = leageNonce2
	tx11.TxData.MinBox = jminBox11
	tx11.TxData.LeagueId = League_1
	tx11.TxData.Fee = fee11

	common.DeepCopy(&tx12, tx6)
	tx12.TxType = tp12
	tx12.TxData.From = Address_2
	tx11.TxData.Account = Address_1
	tx12.TxData.Nonce = ln12
	tx12.TxData.LeagueNonce = leageNonce3
	tx12.TxData.MinBox = jminBox12
	tx12.TxData.LeagueId = League_1
	tx12.TxData.Fee = fee12

	txs = append(txs, tx1, tx2, tx3, tx4, tx5, tx6, tx7, tx8, tx9, tx10, tx11, tx12)

	m.Run()
	os.RemoveAll(dbdir)
}
func TestStateValidate_VerifyTx(t *testing.T) {

	for k, v := range txs[:] {
		//fmt.Println("num:", k, ",hash:", v.Hash())
		result, err := stateValidate.VerifyTx(v)
		switch result {
		case -1:
			t.Log("no.", k, " throw:", err)
		case 0:
			t.Log("no.", k, " back to queue:", err)
		case 1:
			t.Log("no.", k, " err:", err)
		}
	}

	/*	t.Log("-----------------------MemoAccountState---------------------------")
		for k, v := range stateValidate.MemoAccountState {
			t.Logf("address:%s, nonce:%v, bonusheight:%v, balance:%d, energybalance:%d ", k.ToString(), v.Data.Nonce, v.Data.BonusHeight, v.Data.Balance.Uint64(), v.Data.EnergyBalance.Uint64())
		}*/
	/*	t.Log("-----------------------DirtyAccountState---------------------------")
		for k, v := range stateValidate.DirtyAccountState {
			t.Logf("address:%s, nonce:%v, bonusheight:%v, balance:%d, energybalance:%d ", k.ToString(), v.Data.Nonce, v.Data.BonusHeight, v.Data.Balance.Uint64(), v.Data.EnergyBalance.Uint64())
		}*/
	//t.Log("--------------------------------------------------")
	blockInfo := stateValidate.ExecuteOplogs()
	//t.Log("-----------------------Execute---------------------------")
	/*t.Log("\nblockheader:", blockInfo.Block.Header, "\ntxCount:", blockInfo.Block.Transactions.Len())
	t.Log("\nreceiptsCount:", blockInfo.Receipts.Len())*/
	/*for _, v := range blockInfo.AccountStates {
		fmt.Printf("States1: address:%s, nonce:%d, bonusheight:%d, balance:%d, energybalance:%d \n", v.Address.ToString(), v.Data.Nonce, v.Data.BonusHeight, v.Data.Balance.Uint64(), v.Data.EnergyBalance.Uint64())
	}*/
	/*	for _, v := range blockInfo.LeagueStates {
		t.Logf("address:%s, nonce:%d, MemberRoot:%d, FrozenBox:%d, Private:%t, minBox:%d, rate:%d, creator:%s  ", v.Address.ToString(), v.Data.Nonce, v.Data.MemberRoot, v.Data.FrozenBox.Uint64(), v.Data.Private, v.MinBox, v.Rate, v.Creator.ToString())
	}*/

	/*for _, v := range blockInfo.Members {

		t.Logf("league:%s, account:%s,state:%d ", v.LeagueId.ToString(), v.Data.Account.ToString(), v.Data.Status)
	}*/
	t.Log(blockInfo.FeeSum.String())
	//t.Log("--------------------------------------------------")
	/*	fmt.Printf("block:\n txroot:%s\n receipt:%s\n account:%s\n league:%s\n",
		blockInfo.Block.Header.TxRoot.String(),
		blockInfo.Block.Header.ReceiptsRoot.String(),
		blockInfo.Block.Header.StateRoot.String(),
		blockInfo.Block.Header.LeagueRoot.String(),
	)*/
	/*for k, v := range blockInfo.Receipts {
		fmt.Printf("Receipt1: num:%d, status:%t, txhash:%s, gasused:%d \n", k, v.Status, v.TxHash.String(), v.GasUsed)
	}*/

	//blockInfo.Block.Transactions = types.Transactions{}
	_, err := validation.ValidateBlock(blockInfo.Block, ledger)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
}

func TestRemoveElementInLeagueForCreate(t *testing.T) {

	leagueKVs := []*storages.LeagueKV{
		&storages.LeagueKV{Address_1, common.Hash{}},
		&storages.LeagueKV{Address_2, common.Hash{}},
		&storages.LeagueKV{Address_3, common.Hash{}},
		&storages.LeagueKV{Address_4, common.Hash{}},
		&storages.LeagueKV{Address_5, common.Hash{}},
	}
	s := &validation.StateValidate{
		LeagueForCreate: leagueKVs,
	}
	s.RemoveElementInLeagueForCreate(Address_5)
	for _, v := range s.LeagueForCreate {
		t.Log(v.K.ToString())
	}
}
