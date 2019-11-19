package sink_test

import (
	"math/big"
	"testing"

	"bytes"

	"github.com/stretchr/testify/assert"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/common/sink"
)

func TestBigIntToComplex(t *testing.T) {
	a1, _ := big.NewInt(0).SetString("213000000001230000000000000022222222222", 10)
	cp, err := sink.BigIntToComplex(a1)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	a2, err := cp.ComplexToBigInt()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, a1, a2)
}

func TestTxAmountsToComplex(t *testing.T) {

	Address_1, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdy")
	Address_2, _ := common.ToAddress("ct1qK96vAkK6E8S7JgYUY3YY28Qhj6cmfdz")
	tos := &common.TxOuts{}
	tos.Tos = append(tos.Tos,
		&common.TxOut{
			Address: Address_1,
			Amount:  big.NewInt(200),
		},
		&common.TxOut{
			Address: Address_2,
			Amount:  big.NewInt(300),
		},
	)

	cp, err := sink.TxOutsToComplex(tos)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	txa, err := cp.ComplexToTxOuts()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, txa, tos)

}

func TestBytesToComplex(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	buff.WriteString("hello world")
	bytes := buff.Bytes()
	cp, err := sink.BytesToComplex(bytes)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	txa, err := cp.ComplexToBytes()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, txa, bytes)
}
func TestHashArrayToComplex(t *testing.T) {
	ha := common.HashArray{common.Hash{1, 2, 3}, common.Hash{3, 4, 5}, common.Hash{5, 6, 7}}
	cp, err := sink.HashArrayToComplex(ha)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	ha2, err := cp.ComplexToHashArray()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, ha, ha2)
}

func TestBytesArrToComplex(t *testing.T) {
	bs := common.SigBuf{{1, 2, 3}, {4, 5, 6}}
	cp, err := sink.SigBufToComplex(bs)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	bs2, err := cp.ComplexToSigBuf()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	assert.Equal(t, bs, bs2)
}

func TestUintptr(t *testing.T) {
	a, _ := big.NewInt(0).SetString("98091100000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231000000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231000000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231000000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231000000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231000000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231000000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231000000000000000000002222312312312310000000000000000000022223123123123100000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000002000000000000000000000222231231231231230000000000000000000000000000000000000000000000000000000000", 10)
	t.Log(a.BitLen())
	t.Log(len(a.Bytes()))

}
func TestUint16Check(t *testing.T) {
	uint16Max := ^uint16(0)
	t.Log(int(uint16Max))
	t.Log(^uint16(0))

}
