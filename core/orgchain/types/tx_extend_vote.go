package types

import "github.com/sixexorg/magnetic-ring/common"

func voteTxMapInit() {

	txMap[ReplyVote] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "TxHash", Type: common.BType_Hash, Required: true},
		&common.Argument{Name: "VoteReply", Type: common.BType_Uint8, Required: true},
	}

	txMap[VoteIncreaseUT] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Fee", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Nonce", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "Start", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "End", Type: common.BType_Uint64, Required: true},

		&common.Argument{Name: "MetaBox", Type: common.BType_BigInt, Required: true},
		&common.Argument{Name: "Msg", Type: common.BType_Hash, Required: false},
	}
	txMap[VoteApply] = common.Arguments{
		&common.Argument{Name: "From", Type: common.BType_Address, Required: true},
		&common.Argument{Name: "Start", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "End", Type: common.BType_Uint64, Required: true},
		&common.Argument{Name: "TxHash", Type: common.BType_Hash, Required: true},
	}
}
