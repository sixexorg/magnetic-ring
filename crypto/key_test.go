package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sixexorg/magnetic-ring/common"
)

func TestGenerateKey(t *testing.T) {
	privk, err := GenerateKey()

	if err != nil {
		t.Errorf("err-->%v\n", err)
		return
	}

	puk := privk.Public()

	pubstr := common.Bytes2Hex(puk.Bytes())

	t.Logf("pubkbuf-->%v,len=%d\n", pubstr, len(puk.Bytes()))

	srcbuf := []byte("helloleavy")

	sigbuf, err := privk.Sign(srcbuf)
	if err != nil {
		t.Errorf("sigbuf-->%v\n", err)
		return
	}

	sigstr := common.Bytes2Hex(sigbuf)

	t.Logf("sigbuf-->%v\n", sigstr)

	result, err := puk.Verify(srcbuf, sigbuf)
	if err != nil {
		t.Errorf("err-->%v\n", err)
		return
	}

	t.Logf("verify result-->%v\n", result)

	hexStr := privk.Hex()
	t.Logf("privateKey to hex -->%s\n", hexStr)

	privkNew, err := HexToPrivateKey(hexStr)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}

	assert.Equal(t, privk, privkNew)
}

func TestVrf(t *testing.T) {
	privk, err := GenerateKey()

	if err != nil {
		t.Errorf("err-->%v\n", err)
		return
	}

	puk := privk.Public()

	tbf := []byte("zhangping")

	result, proof := privk.Evaluate(tbf)

	fmt.Printf("result=%x\n", result)
	fmt.Printf("prooof=%x\n", proof)

	r2, err := puk.ProofToHash(tbf, proof)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("resuly=%x\n", r2)
	fmt.Printf("compare result=%t\n", r2 == result)

}
