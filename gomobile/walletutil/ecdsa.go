
package walletutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/tyler-smith/go-bip39"
	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/crypto"
	"github.com/sixexorg/magnetic-ring/crypto/secp256k1/bitelliptic"
	"math/big"
)

var (
	curve  = bitelliptic.S256()
	params = curve.Params()
	one    = big.NewInt(1)

)

func GenerateRandomPrvKey(seed string) (string) {
	buffer := new(bytes.Buffer)

	seedbuf,err := common.Hex2Bytes(seed)

	if err != nil {
		return "hex 2 bytes error"
	}

	buffer.Write(seedbuf)
	key, err := ecdsa.GenerateKey(bitelliptic.S256(), buffer)
	if err != nil {
		return ""
	}
	prvBytes := math.PaddedBigBytes(key.D, 32)

	return common.Bytes2Hex(prvBytes)
}

func GetPubKey(prv string) (string) {
	prvBytes,err:=common.Hex2Bytes(prv)
	if err != nil {
		return "hex 2 bytes error"
	}
	key, err := crypto.ToECDSAPrivateKey(prvBytes[:])
	if err != nil {
		return ""
	}
	pubBytes := elliptic.Marshal(bitelliptic.S256(), key.X, key.Y)
	return common.Bytes2Hex(pubBytes)
}


func Sum256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func Sign(prvkstr string, msg string) (string) {


	prvBytes,err := common.Hex2Bytes(prvkstr)
	if err != nil {
		return "hex 2 bytes error"
	}
	prvk,err := crypto.ToECDSAPrivateKey(prvBytes)
	if err != nil {
		return ""
	}
	r, s, err := ecdsa.Sign(rand.Reader, prvk, []byte(msg))
	if err != nil {
		return ""
	}

	sigbuf := serializeSignature(r, s)

	return common.Bytes2Hex(sigbuf)
}

func serializeSignature(r, s *big.Int) []byte {
	size := (params.BitSize + 7) >> 3
	res := make([]byte, size*2)

	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(res[size-len(rBytes):], rBytes)
	copy(res[size*2-len(sBytes):], sBytes)
	return res
}

func Verify(pubstr string, msg string, sig string) bool {
	pubBytes,err := common.Hex2Bytes(pubstr)
	if err != nil {
		return false
	}
	if len(sig) < 64 || len(pubBytes) == 0 {
		return false
	}

	pubk,err := crypto.UnmarshalPubkey(pubBytes)

	if err != nil {
		return false
	}
	sigbuf,err:= common.Hex2Bytes(sig)
	if err != nil {
		return false
	}
	result,err := pubk.Verify([]byte(msg),sigbuf)
	if err != nil {
		return false
	}

	return result
}

func GenerateSeed(pasw string,bitSize int) string {
	//entropy, _ := bip39.NewEntropy(256)
	mnemonic := NewMnemonic(bitSize)

	fmt.Printf("mnemonic-->%s\n",mnemonic)
	//mnemonic = "choice swamp work idle crisp donor all raven perfect dance hedgehog"

	seed := bip39.NewSeed(mnemonic, pasw)
	return common.Bytes2Hex(seed)
}

func GenerateSeedFromMnemonic(mnemonic,pasw string) string {
	//entropy, _ := bip39.NewEntropy(256)
	//mnemonic := NewMnemonic(bitSize)

	fmt.Printf("mnemonic-->%s\n",mnemonic)
	//mnemonic = "choice swamp work idle crisp donor all raven perfect dance hedgehog"

	seed := bip39.NewSeed(mnemonic, pasw)
	return common.Bytes2Hex(seed)
}

func NewMnemonic(bitsize int) string {

	entropy, _ := bip39.NewEntropy(bitsize)//bitsize suggest 256
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return err.Error()
	}
	//mnemonic = "choice swamp work idle crisp donor all raven perfect dance hedgehog"
	return mnemonic
}

func IsMnemonicValid(nem string) bool {
	gui := bip39.IsMnemonicValid(nem)
	return gui
}

