package validation

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
)

type opCode byte

const (
	Account_box_add      opCode = 0x01
	Account_energy_add     opCode = 0x02
	Account_box_sub      opCode = 0x03
	Account_energy_sub     opCode = 0x04
	Account_energy_consume opCode = 0x05
	Account_bonus_left   opCode = 0x06
	Account_nonce_add    opCode = 0x07
	AccountMaxLine       opCode = 0x19

	LeagueMinLine opCode = 0x20
	League_create opCode = 0x21
	//League_rename        opCode = 0x22
	League_frozen_add    opCode = 0x23
	League_frozen_sub    opCode = 0x24
	League_member_apply  opCode = 0x25
	League_member_remove opCode = 0x26
	League_member_add    opCode = 0x27
	League_nonce_add     opCode = 0x28
	League_gas_destroy   opCode = 0x29
	League_raise         opCode = 0x30
	League_gas_sub       opCode = 0x31
	League_minbox        opCode = 0x32
	LeagueMaxLine        opCode = 0x39

	Account_bonus_fee opCode = 0xa0
)

type OpLogs []*OpLog

type OpLog struct {
	method opCode
	data   interface{}
}
type OPAddressAdress struct {
	First  common.Address
	Second common.Address
}
type OPAddressUint64 struct {
	Address common.Address
	Num     uint64
}
type OPAddressBigInt struct {
	Address common.Address
	Num     *big.Int
}
type OPAddressSymbol struct {
	Address common.Address
	Symbol  common.Symbol
}

type OPCreateLeague struct {
	Creator   common.Address
	League    common.Address
	NodeId    common.Address
	FrozenBox *big.Int
	Rate      uint32
	Minbox    uint64
	Private   bool
	TxHash    common.Hash
	Symbol    common.Symbol
}

func CheckTxFee(tx *types.Transaction) bool {
	switch tx.TxType {
	case types.TransferBox, types.TransferEnergy:
		l := len(tx.TxData.Froms.Tis) + len(tx.TxData.Tos.Tos)
		limitFee := big.NewInt(0).SetUint64(types.GetFee(tx.TxType))
		limitFee.Mul(limitFee, big.NewInt(int64(l)))
		return limitFee.Cmp(tx.TxData.Fee) != -1
	case types.MSG:
		//TODO other txType
	}
	return false
}

func HandleTransaction(tx *types.Transaction) []*OpLog {
	oplogs := []*OpLog{}
	re := false
	oplogs, re = commonTxRoute(tx)
	if re {
		return oplogs
	}
	oplogs, re = systemTxRoute(tx)
	if re {
		return oplogs
	}
	oplogs, re = bonusTxRoute(tx)
	if re {
		return oplogs
	}
	return oplogs
}

func oplogPraseAddresAddress(first, second common.Address, method opCode) *OpLog {
	op := &OpLog{
		method: method,
		data: &OPAddressAdress{
			First:  first,
			Second: second,
		},
	}
	return op

}
func oplogPraseAddresUint64(address common.Address, num uint64, method opCode) *OpLog {
	op := &OpLog{
		method: method,
		data: &OPAddressUint64{
			Address: address,
			Num:     num,
		},
	}
	return op
}
func oplogPraseAddressSymbol(address common.Address, symbol common.Symbol, method opCode) *OpLog {
	data := &OPAddressSymbol{
		Address: address,
		Symbol:  symbol,
	}
	return &OpLog{method, data}
}
func oplogPraseAddressBigInt(address common.Address, num *big.Int, method opCode) *OpLog {
	data := &OPAddressBigInt{
		Address: address,
		Num:     big.NewInt(0).Set(num),
	}
	return &OpLog{method, data}
}

// for create league
func oplogPraseCreateLeague(creator, leagueId, nodeId common.Address, frozenBox *big.Int, minbox uint64, rate uint32, private bool, txHash common.Hash, symbol common.Symbol) *OpLog {
	data := &OPCreateLeague{
		Creator:   creator,
		League:    leagueId,
		FrozenBox: frozenBox,
		Rate:      rate,
		Minbox:    minbox,
		Private:   private,
		TxHash:    txHash,
		Symbol:    symbol,
	}
	return &OpLog{League_create, data}
}
