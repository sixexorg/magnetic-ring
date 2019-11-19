package types

import "github.com/sixexorg/magnetic-ring/common"

func commonTxMapInit() {

	txMap[MSG] = common.Arguments{
		&common.Argument{Name: "Msg", Type: common.BType_Hash, Required: true},
	}

	txMap[TransferBox] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Froms", Type: common.BType_TxIns, Required: true},
		&common.Argument{Name: "Tos", Type: common.BType_TxOuts, Required: true},
		&common.Argument{Name: "Msg", Type: common.BType_Hash, Required: false},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
	}

	txMap[TransferEnergy] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Froms", Type: common.BType_TxIns, Required: true},
		&common.Argument{Name: "Tos", Type: common.BType_TxOuts, Required: true},
		&common.Argument{Name: "Msg", Type: common.BType_Hash, Required: false},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
	}

	txMap[CreateLeague] = common.Arguments{
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "NodeId", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "MetaBox", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "MinBox", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "Rate", Type: common.BType_Uint32, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Private", Type: common.BType_Bool, Required: false},
		&common.Argument{Name: "Symbol", Type: common.BType_Symbol, Required: true},
	}

	txMap[JoinLeague] = common.Arguments{
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "MinBox", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "LeagueId", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Account", Type: common.BType_Address, Required: true},
	}

	txMap[RecycleMinbox] = common.Arguments{
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "MinBox", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "LeagueId", Type: common.BType_Address, Required: true},
	}
	txMap[EnergyToLeague] = common.Arguments{
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Energy", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "LeagueId", Type: common.BType_Address, Required: true},
	}
	txMap[AuthX] = common.Arguments{
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "NodePub", Type: common.BType_Bytes, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
	}

}
