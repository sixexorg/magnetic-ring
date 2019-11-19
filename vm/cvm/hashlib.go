package cvm

import (
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"github.com/shopspring/decimal"
	_ "golang.org/x/crypto/ripemd160"
	_ "golang.org/x/crypto/sha3"
	"hash"
)

func OpenHash(l *LState) int {
	mod := l.RegisterModule(HashLibName, hashFuncs)
	l.Push(mod)
	return 1
}

var hashFuncs = map[string]LGFunction{
	"sha3":      hashSha3,
	"ripemd160": hashRipemd160,
	"random":    hashRandom,
}

func hashSha3(l *LState) int {
	sha3256 := crypto.SHA3_256.New()
	return commonHash(l, sha3256)
}

func hashRipemd160(l *LState) int {
	ripemd160 := crypto.RIPEMD160.New()
	return commonHash(l, ripemd160)
}

func commonHash(l *LState, hash hash.Hash) int {
	lval := l.Get(1)
	switch lval.Type() {
	case LTString:
		hash.Write([]byte(lval.String()))
	case LTNumber:
		dec, err := decimal.NewFromString(lval.String())
		if err != nil {

		}

		data, err := dec.MarshalBinary()
		if err != nil {

		}

		hash.Write(data)
	default:
		l.RaiseError("not suppoted param type")
		return 0
	}

	ret := hash.Sum(nil)
	l.Push(LString(hex.EncodeToString(ret)))

	return 1
}

func hashRandom(l *LState) int {
	len := l.OptInt(1, 32)
	if len <= 0 {
		l.ArgError(1, "length must great than zero")
		return 0
	}
	buf := make([]byte, len)
	if _, err := rand.Read(buf); err != nil {
		l.RaiseError("unable calculate the random value! %v, because: %v", l.rawFrameFuncName(l.currentFrame), err)
		return 0
	}

	l.Push(LString(hex.EncodeToString(buf)))
	return 1
}
