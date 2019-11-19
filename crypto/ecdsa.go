// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

// use btcec https://godoc.org/github.com/btcsuite/btcd/btcec#example-package--VerifySignature

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/sixexorg/magnetic-ring/crypto/secp256k1/bitelliptic"
)

const (
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	return bitelliptic.S256()
}

// GenerateKey
func generateEcdsaKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(S256(), rand.Reader)
}

// Sign
func Sign(pk *ecdsa.PrivateKey, hash []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, pk, hash)
	if err != nil {
		return nil, err
	}

	return serializeSignature(r, s), nil
}

// Verify
func Verify(pub *ecdsa.PublicKey, hash, sig []byte) (bool, error) {
	r, s, err := deserializeSignature(sig)
	if err != nil {
		return false, err
	}
	return ecdsa.Verify(pub, hash, r, s), nil
}

// HexToECDSAPrivateKey parses a secp256k1 private key.
func HexToECDSAPrivateKey(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, err
	}
	return ToECDSAPrivateKey(b)
}

// ToECDSAPrivateKey creates a private key with the given data value.
func ToECDSAPrivateKey(d []byte) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	return priv, nil
}

// ToECDSAPublicKey creates a public key with the given data value.
func ToECDSAPublicKey(pub []byte) (*ecdsa.PublicKey, error) {
	if len(pub) == 0 {
		return nil, errors.New("ecdsa: please input public key bytes")
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

// paddedBigBytes encodes a big integer as a big-endian byte slice.
func paddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	readBits(bigint, ret)
	return ret
}

// readBits encodes the absolute value of bigint as big-endian bytes.
func readBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
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

func deserializeSignature(buf []byte) (r, s *big.Int, err error) {
	if buf == nil {
		panic("deserializeSignature: invalid argument")
	}

	length := len(buf)
	if length&1 != 0 {
		return nil, nil, errors.New("invalid length")
	}

	r = new(big.Int).SetBytes(buf[0 : length/2])
	s = new(big.Int).SetBytes(buf[length/2:])
	return r, s, nil
}

// ZeroKey zeroes the private key
func ZeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

var mask = []byte{0xff, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f}

func generateKeyFromCurve(curve *bitelliptic.BitCurve, rand io.Reader) (priv []byte, x, y *big.Int, err error) {
	N := curve.Params().N
	bitSize := N.BitLen()
	byteLen := (bitSize + 7) >> 3
	priv = make([]byte, byteLen)

	for x == nil {
		_, err = io.ReadFull(rand, priv)
		if err != nil {
			return
		}

		// We have to mask off any excess bits in the case that the size of the
		// underlying field is not a whole number of bytes.
		priv[0] &= mask[bitSize%8]
		// This is because, in tests, rand will return all zeros and we don't
		// want to get the point at infinity and loop forever.
		priv[1] ^= 0x42

		// If the scalar is out of range, sample another random number.
		if new(big.Int).SetBytes(priv).Cmp(N) >= 0 {
			continue
		}

		x, y = curve.ScalarBaseMult(priv)
	}
	return
}

func toECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = curve
	if 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(curve.N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}

	return priv, nil
}
