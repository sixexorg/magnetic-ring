package common

import (
	"testing"
)

func TestHash_String(t *testing.T) {
	byteArr := Sha256([]byte("who are you!"))
	hash, err := ParseHashFromBytes(byteArr)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	t.Log(Bytes2Hex(byteArr))
	t.Log(hash.String())

	t.Logf("NormalAddress=%d\n",NormalAddress)
	t.Logf("MultiAddress=%d\n",MultiAddress)
	t.Logf("OrganizationAddress=%d\n",OrganizationAddress)
	t.Logf("ContractAddress=%d\n",ContractAddress)
}

