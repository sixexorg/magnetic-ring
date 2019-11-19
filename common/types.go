package common

import (
	"math/big"
	"reflect"
)

type Arguments []*Argument

type Argument struct {
	Name     string
	Type     reflect.Type
	Required bool
	HashOut  bool
}

func (s Arguments) Len() int           { return len(s) }
func (s Arguments) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Arguments) Less(i, j int) bool { return s[i].Name < s[j].Name }

var (
	BType_TxIns       = reflect.TypeOf(TxIns{})
	BType_TxOuts      = reflect.TypeOf(TxOuts{})
	BType_Address     = reflect.TypeOf(Address{})
	BType_Symbol      = reflect.TypeOf(Symbol{})
	BType_Hash        = reflect.TypeOf(Hash{})
	BType_TxHashArray = reflect.TypeOf(HashArray{})
	BType_SigBuf      = reflect.TypeOf(SigBuf{})
	BType_BigInt      = reflect.TypeOf(big.NewInt(0))
	BType_Uint8       = reflect.TypeOf(uint8(0))
	BType_Uint16      = reflect.TypeOf(uint16(0))
	BType_Uint32      = reflect.TypeOf(uint32(0))
	BType_Uint64      = reflect.TypeOf(uint64(0))
	BType_Bool        = reflect.TypeOf(true)
	BType_Bytes       = reflect.TypeOf([]byte{})
)

type TxOut struct {
	Address Address
	Amount  *big.Int
}

//rlp can't type TxAmounts []*TxAmount
type TxOuts struct {
	Tos []*TxOut
}

type TxIn struct {
	Address Address
	Amount  *big.Int
	Nonce   uint64
}
type TxIns struct {
	Tis []*TxIn
}
