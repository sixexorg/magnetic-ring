package types

import "github.com/sixexorg/magnetic-ring/common"

func systemTxMapInit() {
	txMap[EnergyFromMain] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "TxHash", Type: common.BType_Hash, Required: true},
		&common.Argument{Name: "Energy", Type: common.BType_BigInt, Required: true},
	}

	txMap[Join] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "TxHash", Type: common.BType_Hash, Required: true},
	}
	txMap[RaiseUT] = common.Arguments{
		&common.Argument{Name: "MetaBox", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "TxHash", Type: common.BType_Hash, Required: true},
		&common.Argument{Name: "Rate", Type: common.BType_Uint32, Required: true},
	}
}
