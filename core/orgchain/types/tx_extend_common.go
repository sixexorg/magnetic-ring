package types

import "github.com/sixexorg/magnetic-ring/common"

func commonTxMapInit() {
	txMap[MSG] = common.Arguments{
		&common.Argument{Name: "Msg", Type: common.BType_Hash, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
	}

	txMap[TransferUT] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Froms", Type: common.BType_TxIns, Required: true},
		&common.Argument{Name: "Tos", Type: common.BType_TxOuts, Required: true},
		&common.Argument{Name: "Msg", Type: common.BType_Hash, Required: false},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
	}

	txMap[EnergyToMain] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Energy", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
	}
}
