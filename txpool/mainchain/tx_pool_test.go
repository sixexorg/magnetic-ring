package mainchain

import (
	"bytes"
	"encoding/gob"
	"math/big"
	"testing"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
	"github.com/sixexorg/magnetic-ring/store/mainchain/storages"
	"github.com/sixexorg/magnetic-ring/store/mainchain/validation"
)

var (
	balance1, balance2                                = big.NewInt(1000), big.NewInt(2000)
	energy1, energy2                                      = big.NewInt(5000), big.NewInt(4000)
	nonce1, n1, n2, n3, n4, n5                 uint64 = 50, 51, 52, 53, 54, 55
	nonce2, m1, m2, m3, m4, m5                 uint64 = 100, 101, 102, 103, 104, 105
	fee1, fee2, fee3, fee4, fee5                      = big.NewInt(20), big.NewInt(10), big.NewInt(2), big.NewInt(20), big.NewInt(5)
	tx1a1, tx1a2, touta1_1, touta1_2, touta1_3        = big.NewInt(20), big.NewInt(108), big.NewInt(30), big.NewInt(50), big.NewInt(28)
	tx2a1, tx2a2, touta2_1, touta2_2, touta2_3        = big.NewInt(80), big.NewInt(300), big.NewInt(50), big.NewInt(50), big.NewInt(280)
	tx3a1, tx3a2, touta3_1, touta3_2, touta3_3        = big.NewInt(800), big.NewInt(500), big.NewInt(500), big.NewInt(300), big.NewInt(500)
	tx4a1, tx4a2, touta4_1, touta4_2, touta4_3        = big.NewInt(200), big.NewInt(20), big.NewInt(20), big.NewInt(100), big.NewInt(100)
	tx5a1, tx5a2, touta5_1, touta5_2, touta5_3        = big.NewInt(3000), big.NewInt(40), big.NewInt(1000), big.NewInt(1000), big.NewInt(1040)

	Address_1, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfda")
	Address_2, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdb")
	Address_3, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdc")
	Address_4, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdd")
	Address_5, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfde")

	height = uint64(50)

	tx1 *types.Transaction
	tx2 *types.Transaction
	tx3 *types.Transaction
	tx4 *types.Transaction
	tx5 *types.Transaction
	txs []*types.Transaction
)

func TestRun(t *testing.T) {

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
		//TxType:  types.TransferBox,		//转box
		TxType: types.TransferEnergy, //转energy
		TxData: txd1,
	}
	deepCopy(&tx2, tx1)
	tx2.TxData.Froms.Tis[0].Nonce = n2
	tx2.TxData.Froms.Tis[0].Amount = tx2a1
	tx2.TxData.Froms.Tis[1].Nonce = m2
	tx2.TxData.Froms.Tis[1].Amount = tx2a2
	tx2.TxData.Tos.Tos[0].Amount = touta2_1
	tx2.TxData.Tos.Tos[1].Amount = touta2_2
	tx2.TxData.Tos.Tos[2].Amount = touta2_3
	tx2.TxData.Fee = fee2
	deepCopy(&tx3, tx1)
	tx3.TxData.Froms.Tis[0].Nonce = n3
	tx3.TxData.Froms.Tis[0].Amount = tx3a1
	tx3.TxData.Froms.Tis[1].Nonce = m3
	tx3.TxData.Froms.Tis[1].Amount = tx3a2
	tx3.TxData.Tos.Tos[0].Amount = touta3_1
	tx3.TxData.Tos.Tos[1].Amount = touta3_2
	tx3.TxData.Tos.Tos[2].Amount = touta3_3
	tx3.TxData.Fee = fee3
	deepCopy(&tx4, tx1)
	tx4.TxData.Froms.Tis[0].Nonce = n4
	tx4.TxData.Froms.Tis[0].Amount = tx4a1
	tx4.TxData.Froms.Tis[1].Nonce = m4
	tx4.TxData.Froms.Tis[1].Amount = tx4a2
	tx4.TxData.Tos.Tos[0].Amount = touta4_1
	tx4.TxData.Tos.Tos[1].Amount = touta4_2
	tx4.TxData.Tos.Tos[2].Amount = touta4_3
	tx4.TxData.Fee = fee4
	deepCopy(&tx5, tx1)
	tx5.TxData.Froms.Tis[0].Nonce = n5
	tx5.TxData.Froms.Tis[0].Amount = tx5a1
	tx5.TxData.Froms.Tis[1].Nonce = m5
	tx5.TxData.Froms.Tis[1].Amount = tx5a2
	tx5.TxData.Tos.Tos[0].Amount = touta5_1
	tx5.TxData.Tos.Tos[1].Amount = touta5_2
	tx5.TxData.Tos.Tos[2].Amount = touta5_3
	tx5.TxData.Fee = fee5

	txs = append(txs, tx1, tx2, tx3, tx4, tx5)

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

	acctstore, err := storages.NewAccoutStore("./storedir", true)

	if err != nil {
		t.Error(err)
		return
	}

	pool := NewTxPool(acctstore)
	setPool(pool.stateValidator)
	pool.Start()

	t.Logf("pool:%+v\n", pool)

	tx1.Raw = []byte("raw tx1")
	tx2.Raw = []byte("raw tx2")
	tx3.Raw = []byte("raw tx3")
	tx4.Raw = []byte("raw tx4")
	tx5.Raw = []byte("raw tx5")

	pool.AddTx(tx1)
	//pool.AddTx(tx2)
	//pool.AddTx(tx3)
	//pool.AddTx(tx4)
	//pool.AddTx(tx5)

	blk := pool.GenerateBlock(2, true)

	t.Logf("===============================================block info bg===============================================")

	for i, v := range blk.Transactions {

		t.Logf("index=%d,\thash=%s,type=%d,address=%s,fee=%d\n", i, v.Hash(), v.TxType, v.TxData.From, v.TxData.Fee.Uint64())
	}

	t.Logf("===============================================block info ed===============================================")

	t.Logf("blk-->%+v\n", blk)

	t.Logf("pool.queue:%+v\n", pool.queue)

	t.Log("-----------------------MemoAccountState---------------------------")
	for k, v := range pool.stateValidator.MemoAccountState {
		t.Logf("address:%d, nonce:%v, bonusheight:%v, balance:%d, energybalance:%d ", k, v.Data.Nonce, v.Data.BonusHeight, v.Data.Balance.Uint64(), v.Data.EnergyBalance.Uint64())
	}

	t.Logf("===============================================")
	blockInfo := pool.Execute()

	for k, v := range blockInfo.AccountStates {
		t.Logf("address:%d, nonce:%d, bonusheight:%d, balance:%d, energybalance:%d ", k, v.Data.Nonce, v.Data.BonusHeight, v.Data.Balance.Uint64(), v.Data.EnergyBalance.Uint64())
	}

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt, os.Kill,syscall.SIGTERM)
	//<-c

}

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func setPool(stateValidate *validation.StateValidate) {
	stateValidate.ParentBonusHeight = 1000
	stateValidate.MemoAccountState[Address_1] = &states.AccountState{
		Address: Address_1,
		Height:  height,
		Data: &states.Account{
			Nonce:       nonce1,
			Balance:     balance1,
			EnergyBalance: energy1,
			BonusHeight: 50,
		},
	}
	stateValidate.MemoAccountState[Address_2] = &states.AccountState{
		Address: Address_2,
		Height:  height,
		Data: &states.Account{
			Nonce:       nonce2,
			Balance:     balance2,
			EnergyBalance: energy2,
			BonusHeight: 40,
		},
	}
	stateValidate.MemoAccountState[Address_3] = &states.AccountState{
		Address: Address_3,
		Height:  height,
		Data: &states.Account{
			Nonce:       0,
			Balance:     big.NewInt(10000),
			EnergyBalance: big.NewInt(20000),
			BonusHeight: 0,
		},
	}
	stateValidate.MemoAccountState[Address_4] = &states.AccountState{
		Address: Address_4,
		Height:  height,
		Data: &states.Account{
			Nonce:       0,
			Balance:     big.NewInt(0),
			EnergyBalance: big.NewInt(0),
			BonusHeight: 0,
		},
	}
	stateValidate.MemoAccountState[Address_5] = &states.AccountState{
		Address: Address_5,
		Height:  height,
		Data: &states.Account{
			Nonce:       0,
			Balance:     big.NewInt(0),
			EnergyBalance: big.NewInt(0),
			BonusHeight: 0,
		},
	}

	deepCopy(&stateValidate.DirtyAccountState, stateValidate.MemoAccountState)
}
