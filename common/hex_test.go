package common_test

import (
	"testing"

	"github.com/sixexorg/magnetic-ring/common"
)

func TestUint16ToBytes(t *testing.T) {
	t.Log(common.Uint16ToBytes(554))
}

func TestBytesToUint16(t *testing.T) {

	bytearr := []byte{0, 4}
	t.Log(common.BytesToUint16(bytearr))
}
