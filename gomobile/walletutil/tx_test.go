package walletutil

import (
	"fmt"
	"github.com/sixexorg/magnetic-ring/crypto"
	"testing"
)


func TestTransferBox(t *testing.T) {
	tbtx := TransferBox("04cd6a2853d090e8a0031c3c4ebfaaa3fbefc2672424561f75ce6e83d33240eaa90292c556fc804d22366b87bb2ee91f254bf0f38e1c83fd3fabaffad5d8a560c8","ct1K8X6KQMKs2j6mCg99uC6M463m8ug6ze5","this is msg",200,200,2)
	fmt.Printf("tbtx=%s\n",tbtx)
}


func TestSignMainTransaction(t *testing.T) {
	signedtx := SignMainTransaction("0110020008c1c04e0009f84cf84af848f84601f843b84104cd6a2853d090e8a0031c3c4ebfaaa3fbefc2672424561f75ce6e83d33240eaa90292c556fc804d22366b87bb2ee91f254bf0f38e1c83fd3fabaffad5d8a560c8010003c889fa595a72c06259ade84696598ca299f4f2fda3011c0001dbdad99589fa595a72c06259ade84696598ca299f4f2fda30181c80200000000000000000000000000000000000000000000000000000000000000001b0002dad9d89591fa595a72c06259ade84696598ca299f4f2fda40181c8","ct1J8X6KQMKs2j6mCg99uC6M463m8ug6ze4","70fe34923ddd9762b9ed9e0258721b406a29b46628680c75aeb18b6ededeab8a")
	fmt.Printf(signedtx)
}

func TestPk(t *testing.T) {
	pk,err := crypto.HexToPrivateKey("70fe34923ddd9762b9ed9e0258721b406a29b46628680c75aeb18b6ededeab8a")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%s\n",pk.Public().Hex())
}