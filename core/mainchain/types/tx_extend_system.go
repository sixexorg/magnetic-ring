package types

import "github.com/sixexorg/magnetic-ring/common"

func systemTxMapInit() {
	//system
	txMap[ConsensusLeague] = common.Arguments{
		&common.Argument{Name: "StartHeight", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "EndHeight", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "BlockRoot", Type: common.BType_Hash, Required: true},
		&common.Argument{Name: "LeagueId", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
	}

	txMap[LockLeague] = common.Arguments{
		&common.Argument{Name: "StartHeight", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "EndHeight", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "BlockRoot", Type: common.BType_Hash, Required: true},
		&common.Argument{Name: "LeagueId", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "UnlockEnergy", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Msg", Type: common.BType_Hash, Required: false},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
	}

	txMap[RaiseUT] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "LeagueId", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "MetaBox", Type: common.BType_BigInt, Required: true},
	}

	txMap[ApplyPass] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "LeagueId", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "TxHash", Type: common.BType_Hash, Required: true},
	}
}
