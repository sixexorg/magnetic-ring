package validation

import (
	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/orgchain/types"
)

type opCode byte

const (
	Account_ut_add       opCode = 0x01
	Account_energy_add     opCode = 0x02
	Account_ut_sub       opCode = 0x03
	Account_energy_sub     opCode = 0x04
	Account_energy_consume opCode = 0x05
	Account_bonus_left   opCode = 0x06
	Account_nonce_add    opCode = 0x07
	AccountMaxLine       opCode = 0x19

	Vote_against      opCode = 0x20
	Vote_agree        opCode = 0x21
	Vote_abstention   opCode = 0x22
	VoteMaxLine       opCode = 0x23
	Account_bonus_fee opCode = 0xa0
	League_Raise_ut   opCode = 0xb0
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
type OPAddressUin32 struct {
	Address common.Address
	Num     *big.Int
}
type OPHashAddress struct {
	Hash    common.Hash
	Address common.Address
}

type OPCreateLeague struct {
	Creator   common.Address
	League    common.Address
	NodeId    common.Address
	FrozenBox *big.Int
	Rate      uint32
	Minbox    uint64
	Private   bool
}

func CheckTxFee(tx *types.Transaction) bool {
	switch tx.TxType {
	case types.TransferUT:
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
	oplogs, re = voteTxRoute(tx)
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
func oplogPraseAddresUint64(address common.Address, nonce uint64, method opCode) *OpLog {
	op := &OpLog{
		method: method,
		data: &OPAddressUint64{
			Address: address,
			Num:     nonce,
		},
	}
	return op
}

func oplogPraseAddressBigInt(address common.Address, num *big.Int, method opCode) *OpLog {
	data := &OPAddressBigInt{
		Address: address,
		Num:     num,
	}
	return &OpLog{method, data}
}

func oplogPraseHashAddress(hash common.Hash, address common.Address, method opCode) *OpLog {
	data := &OPHashAddress{
		Hash:    hash,
		Address: address,
	}
	return &OpLog{method, data}
}
