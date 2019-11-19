package sink

import (
	"math/big"

	"bytes"

	"reflect"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/errors"
	"github.com/sixexorg/magnetic-ring/rlp"
)

const (
	TxIns     byte = 0x01
	TxOuts    byte = 0x02
	BigInt    byte = 0x03
	HashArray byte = 0x04
	Bytes     byte = 0x05
	Addresses byte = 0x06
	SigBuff   byte = 0x07
	SigPack   byte = 0x08
	Tmplt     byte = 0x09

	maxU16 = int(^uint16(0))
)

type ComplexType struct {
	Size  [2]byte
	MType byte
	Data  []byte
}

func ZeroCopySourceRelease(source *ZeroCopySource, typ reflect.Type) (out interface{}, irregular, eof bool) {
	eof = true
	switch typ {
	case common.BType_Bool:
		out, irregular, eof = source.NextBool()
	case common.BType_Uint8:
		out, eof = source.NextUint8()
	case common.BType_Uint16:
		out, eof = source.NextUint16()
	case common.BType_Uint32:
		out, eof = source.NextUint32()
	case common.BType_Uint64:
		out, eof = source.NextUint64()
	case common.BType_BigInt:
		bg := &ComplexType{}
		bg, eof = source.NextComplex()
		out, _ = bg.ComplexToBigInt()
	case common.BType_Hash:
		out, eof = source.NextHash()
	case common.BType_Address:
		out, eof = source.NextAddress()
	case common.BType_Symbol:
		out, eof = source.NextSymbol()
	case common.BType_TxIns:
		bg := &ComplexType{}
		bg, eof = source.NextComplex()
		out, _ = bg.ComplexToTxIns()
	case common.BType_TxOuts:
		bg := &ComplexType{}
		bg, eof = source.NextComplex()
		out, _ = bg.ComplexToTxOuts()
	case common.BType_TxHashArray:
		bg := &ComplexType{}
		bg, eof = source.NextComplex()
		out, _ = bg.ComplexToHashArray()
	case common.BType_SigBuf:
		bg := &ComplexType{}
		bg, eof = source.NextComplex()
		out, _ = bg.ComplexToSigBuf()
	case common.BType_Bytes:
		bg := &ComplexType{}
		bg, eof = source.NextComplex()
		out, _ = bg.ComplexToBytes()
	}
	return
}

func ZeroCopySinkAppend(sk *ZeroCopySink, value reflect.Value) error {
	switch value.Interface().(type) {
	case bool:
		sk.WriteBool(value.Interface().(bool))
	case uint8:
		sk.WriteUint8(value.Interface().(uint8))
	case uint16:
		sk.WriteUint16(value.Interface().(uint16))
	case uint32:
		sk.WriteUint32(value.Interface().(uint32))
	case uint64:
		sk.WriteUint64(value.Interface().(uint64))
	case *big.Int:
		cpx, err := BigIntToComplex(value.Interface().(*big.Int))
		if err != nil {
			return err
		}
		sk.WriteComplex(cpx)
	case common.Hash:
		sk.WriteHash(value.Interface().(common.Hash))
	case common.Address:
		sk.WriteAddress(value.Interface().(common.Address))
	case common.Symbol:
		sk.WriteSymbol(value.Interface().(common.Symbol))
	case *common.TxIns:
		cpx, err := TxInsToComplex(value.Interface().(*common.TxIns))
		if err != nil {
			return err
		}
		sk.WriteComplex(cpx)
	case *common.TxOuts:
		cpx, err := TxOutsToComplex(value.Interface().(*common.TxOuts))
		if err != nil {
			return err
		}
		sk.WriteComplex(cpx)
	case *common.HashArray:
		cpx, err := HashArrayToComplex(value.Interface().(common.HashArray))
		if err != nil {
			return err
		}
		sk.WriteComplex(cpx)
	case *common.SigBuf:
		cpx, err := SigBufToComplex(value.Interface().(common.SigBuf))
		if err != nil {
			return err
		}
		sk.WriteComplex(cpx)
	case []byte:
		arr := value.Interface().([]byte)
		cpx, err := BytesToComplex(arr)
		if err != nil {
			return err
		}
		sk.WriteComplex(cpx)
	}
	return nil
}

func DataCheck(data []byte) (size [2]byte, err error) {
	l := len(data)
	if l > int(^uint16(0)) {
		err = errors.ERR_TX_DATA_OVERFLOW

	} else {
		u16 := common.Uint16ToBytes(uint16(l))
		size = [2]byte{u16[0], u16[1]}
	}
	return
}
func BigIntToComplex(int *big.Int) (*ComplexType, error) {
	if int == nil {
		int = big.NewInt(0)
	}
	size, err := DataCheck(int.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: BigInt,
		Data:  int.Bytes(),
	}
	return ct, nil
}

func TxInsToComplex(tas *common.TxIns) (*ComplexType, error) {
	buff := new(bytes.Buffer)
	err := rlp.Encode(buff, tas)
	if err != nil {
		return nil, err
	}
	size, err := DataCheck(buff.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: TxIns,
		Data:  buff.Bytes(),
	}
	return ct, nil
}

func TxOutsToComplex(tas *common.TxOuts) (*ComplexType, error) {
	buff := new(bytes.Buffer)
	err := rlp.Encode(buff, tas)
	if err != nil {
		return nil, err
	}
	size, err := DataCheck(buff.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: TxOuts,
		Data:  buff.Bytes(),
	}
	return ct, nil
}

func SigPackToComplex(tas *common.SigPack) (*ComplexType, error) {
	if tas == nil {
		tas = &common.SigPack{
			make([]*common.SigMap, 0),
		}
	}
	buff := new(bytes.Buffer)
	err := rlp.Encode(buff, tas)
	if err != nil {
		return nil, err
	}
	size, err := DataCheck(buff.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: SigPack,
		Data:  buff.Bytes(),
	}
	return ct, nil
}

func (cpx *ComplexType) ComplexToSigpack() (*common.SigPack, error) {
	if cpx.MType != SigPack {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	var fs common.SigPack
	err := rlp.Decode(bytes.NewReader(cpx.Data), &fs)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func (cpx *ComplexType) ComplexToTxOuts() (*common.TxOuts, error) {
	if cpx.MType != TxOuts {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	var fs common.TxOuts
	err := rlp.Decode(bytes.NewReader(cpx.Data), &fs)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func (cpx *ComplexType) ComplexToTxIns() (*common.TxIns, error) {
	if cpx.MType != TxIns {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	var fs common.TxIns
	err := rlp.Decode(bytes.NewReader(cpx.Data), &fs)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func (cpx *ComplexType) ComplexToBigInt() (*big.Int, error) {
	if cpx.MType != BigInt {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	return new(big.Int).SetBytes(cpx.Data), nil
}

func HashArrayToComplex(tas common.HashArray) (*ComplexType, error) {
	buff := new(bytes.Buffer)
	err := rlp.Encode(buff, tas)
	if err != nil {
		return nil, err
	}
	size, err := DataCheck(buff.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: HashArray,
		Data:  buff.Bytes(),
	}
	return ct, nil
}

func (cpx *ComplexType) ComplexToHashArray() (common.HashArray, error) {
	if cpx.MType != HashArray {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	var fs common.HashArray
	err := rlp.Decode(bytes.NewReader(cpx.Data), &fs)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func BytesToComplex(bytes []byte) (*ComplexType, error) {
	size, err := DataCheck(bytes)
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: Bytes,
		Data:  bytes,
	}
	return ct, nil
}
func (cpx *ComplexType) ComplexToBytes() ([]byte, error) {
	if cpx.MType != Bytes {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	return cpx.Data, nil
}
func (cpx *ComplexType) ComplexToTxAddresses() ([]common.Address, error) {
	if cpx.MType != Addresses {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	var fs []common.Address
	err := rlp.Decode(bytes.NewReader(cpx.Data), &fs)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func AddressesToComplex(addresses []common.Address) (*ComplexType, error) {
	buff := new(bytes.Buffer)
	err := rlp.Encode(buff, addresses)
	if err != nil {
		return nil, err
	}
	size, err := DataCheck(buff.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: Addresses,
		Data:  buff.Bytes(),
	}
	return ct, nil
}

func SigBufToComplex(tas common.SigBuf) (*ComplexType, error) {
	buff := new(bytes.Buffer)
	err := rlp.Encode(buff, tas)
	if err != nil {
		return nil, err
	}
	size, err := DataCheck(buff.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: SigBuff,
		Data:  buff.Bytes(),
	}
	return ct, nil
}

func (cpx *ComplexType) ComplexToSigBuf() (common.SigBuf, error) {
	if cpx.MType != SigBuff {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	var fs common.SigBuf
	err := rlp.Decode(bytes.NewReader(cpx.Data), &fs)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func SigTmpltToComplex(tas *common.SigTmplt) (*ComplexType, error) {
	if tas == nil {
		tas = &common.SigTmplt{
			make([]*common.Maus, 0),
		}
	}
	buff := new(bytes.Buffer)
	err := rlp.Encode(buff, tas)
	if err != nil {
		return nil, err
	}
	size, err := DataCheck(buff.Bytes())
	if err != nil {
		return nil, err
	}
	ct := &ComplexType{
		Size:  size,
		MType: Tmplt,
		Data:  buff.Bytes(),
	}
	return ct, nil
}

func (cpx *ComplexType) ComplexToSigTmplt() (*common.SigTmplt, error) {
	if cpx.MType != Tmplt {
		return nil, errors.ERR_SINK_TYPE_DIFF
	}
	var fs common.SigTmplt
	err := rlp.Decode(bytes.NewReader(cpx.Data), &fs)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}
