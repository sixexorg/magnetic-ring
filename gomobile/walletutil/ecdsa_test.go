
package walletutil

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto/secp256k1/bitelliptic"
	"github.com/sixexorg/magnetic-ring/radar/orgchain"
	"testing"
)

func TestGenerateAppEcdsaKey(t *testing.T) {

	seed := GenerateSeed("abcdefg",128)


	fmt.Printf("seed=%s\n",seed)

	prvBytestr := GenerateRandomPrvKey(seed)

	pubBytes := GetPubKey(prvBytestr)

	addr := ToAddress(pubBytes)
	fmt.Printf("address-->%s\n",addr)

	fmt.Printf("pubBytes: %s\n", pubBytes)

	msg := "hello world!"

	sigbuf := Sign(prvBytestr,msg)

	fmt.Printf("src: %s\n", msg)


	fmt.Printf("sigbuf:%s\n",sigbuf)

	result := Verify(pubBytes,msg,sigbuf)

	t.Logf("verify result-->%v\n",result)


}




func TestVerify(t *testing.T) {

	pubkbytes := "049c42b72d86685a2c2069933d5e7996e52cfe7215f3d7af1b4c48ba93bd53a7f38e23f8a97d71af07c0a8733016baa8a1e3122f5e7c0da740d0d2710db0d8dd36"
	sigstr := "db784b8b6521f0916494a3aad84ef4e03de5b45400bb776712b912ca0ca6b36c3c3cf00ce9ba473c526e675c1b18346fabf6bcaa0653fc9440ce534a3ac1a06f"
	pubbuf,err := common.Hex2Bytes(pubkbytes)
	if err != nil {
		t.Errorf("error=%v\n",err)
		return
	}


	fmt.Printf("pubBytes: %v\n", pubbuf)
	fmt.Printf("pubBytes: %X\n", pubbuf)
	fmt.Printf("pubBytes len: %d\n", len(pubbuf))

	msg := "hello world!"


	//sigbuf,err := Sign(prvBytes,msg)
	//if err != nil {
	//	t.Error("err-->%v\n",err)
	//	return
	//}

	sigbuf,err := common.Hex2Bytes(sigstr)
	if err != nil {
		t.Errorf("error=%v\n",err)
		return
	}

	t.Logf("sigbuf-->%x\n",sigbuf)


	result := Verify(pubkbytes,msg,sigstr)

	t.Logf("verify result-->%v\n",result)


}

func TestNewMnemonic(t *testing.T) {
	//nem:=NewMnemonic(128)
	//fmt.Printf("nem:%s\n",nem)

	nem := "session bike timber matter scare region cloud small since turtle wasp winner"

	seed := bip39.NewSeed(nem, "")


	buffer := new(bytes.Buffer)

	//seedbuf := common.Hex2Bytes(seed)

	buffer.Write(seed)

	key, err := ecdsa.GenerateKey(bitelliptic.S256(), buffer)
	if err != nil {
		return
	}
	prvBytes := math.PaddedBigBytes(key.D, 32)

	hexstr := common.Bytes2Hex(prvBytes)

	fmt.Printf("prvkstr:%x\n",prvBytes)
	//hex := fmt.Sprintf("%x")
	k,err := crypto.HexToECDSA(hexstr)
	if err != nil {
		t.Errorf("error=%v\n",err)
		return
	}

	addr := crypto.PubkeyToAddress(k.PublicKey)
	fmt.Printf("final addr:%s\n",addr.String())

}

func TestGetaccountnonce(t *testing.T) {
	str := GetAccountBalance("http://127.0.0.1:10086","ct1J8X6KQMKs2j6mCg99uC6M463m8ug6ze4","orgid")
	fmt.Printf("str=%s\n",str)
}

func TestCreateLeague(t *testing.T) {
	str := GetOrgInfo("http://192.168.10.153:20086","ct27hJxDpRvW63TCN2MF6PyT2D1ayqjv24x")

	fmt.Printf("str=%s\n",str)
}

func TestCap(t *testing.T)  {
	fmt.Printf("begin\n")
	re := &orgchain.ResultOp{}
	fmt.Printf("☂️ ☂️ ☂️ ☂️ ☂️ ☂️ ☂️ ☂️ re--->%+v\n",re)
	select {
	case orgTx := <-re.ORGTX:
		fmt.Println("☃️ digestMainTx ORGTX--->", orgTx)

	case mcc := <-re.MCC:
		fmt.Println("☃️ digestMainTx MCC--->", mcc)
	}

}