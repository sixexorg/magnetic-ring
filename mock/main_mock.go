package mock

import (
	"math/big"

	"os"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

var (
	Mock_Tx_CreateLeague *types.Transaction

	Mock_Tx        *types.Transaction
	Mock_TxData    *types.TxData
	Mock_Address_1 common.Address
	Mock_Address_2 common.Address

	MockHash common.Hash
)

func CatchPanic(dbdir string) {
	if err := recover(); err != nil {
		goto REMOVE
	}
REMOVE:
	os.RemoveAll(dbdir)
}

func init() {

	nodeId, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmaaa")

	Mock_Tx_CreateLeague = &types.Transaction{
		Version: 0x01,
		TxType:  types.CreateLeague,
		TxData: &types.TxData{
			From:    Address_1,
			Nonce:   1,
			Fee:     big.NewInt(20),
			Rate:    20,
			MinBox:  10,
			MetaBox: big.NewInt(200),
			NodeId:  nodeId,
			Private: false,
		},
	}

	//address, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdx")
	Mock_Address_1, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdy")
	Mock_Address_2, _ = common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdz")
	MockHash = common.Hash{1, 'a', 'd', 3}
	froms := &common.TxIns{}
	froms.Tis = append(froms.Tis,
		&common.TxIn{
			Address: Mock_Address_1,
			Nonce:   100,
			Amount:  big.NewInt(200),
		},
		&common.TxIn{
			Address: Mock_Address_2,
			Nonce:   200,
			Amount:  big.NewInt(300),
		},
	)
	tos := &common.TxOuts{}
	tos.Tos = append(tos.Tos,
		&common.TxOut{
			Address: Mock_Address_1,
			Amount:  big.NewInt(200),
		},
		&common.TxOut{
			Address: Mock_Address_2,
			Amount:  big.NewInt(300),
		},
	)
	Mock_TxData = &types.TxData{
		Froms: froms,
		Tos:   tos,
		Fee:   big.NewInt(100),
	}

	Mock_Tx = &types.Transaction{
		Version: 0x01,
		TxType:  types.TransferBox,
		TxData:  Mock_TxData,
	}

}
