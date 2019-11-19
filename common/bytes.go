package common

import (
	"encoding/base32"
	"encoding/hex"
)

const Alphabet = `123456789abcdefghjkmnpqrstuvwxyz`

var encoding = base32.NewEncoding(Alphabet)

func ToBase32(hash []byte) string {
	return encoding.EncodeToString(hash)
}

func FromBase32(std32 string) ([]byte, error) {
	return encoding.DecodeString(std32)
}

// Bytes2Hex returns the hexadecimal encoding of d.
func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) ([]byte, error) {
	h, err := hex.DecodeString(str)
	return h, err
}
