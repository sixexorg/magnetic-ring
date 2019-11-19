package common

import (
	"crypto"

	"github.com/sixexorg/magnetic-ring/errors"

	_ "golang.org/x/crypto/ripemd160"
	_ "golang.org/x/crypto/sha3"
)

func Sha256(msg ...[]byte) []byte {
	sha3256 := crypto.SHA3_256.New()
	for _, bytes := range msg {
		sha3256.Write(bytes)
	}
	return sha3256.Sum(nil)
}

func CalcHash(buf []byte) Hash {
	result := Hash{}

	sum := Sha256(buf)

	copy(result[:],sum)
	return result
}

func Ripemd160(msg []byte) []byte {
	ripemd160 := crypto.RIPEMD160.New()
	ripemd160.Write(msg)
	return ripemd160.Sum(nil)
}

func Sha256Ripemd160(msg []byte) []byte {
	return Ripemd160(Sha256(msg))
}

func ZeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}
func ParseHashFromBytes(bytes []byte) (Hash, error) {
	if len(bytes) != HashLength {
		return Hash{}, errors.ERR_HASH_PARSE
	}
	var hash Hash
	copy(hash[:], bytes)
	return hash, nil
}
