package types

import "github.com/sixexorg/magnetic-ring/common"

func bonusTxMapInit() {
	txMap[GetBonus] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
	}
}
