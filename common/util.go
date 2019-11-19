package common

import (
	"crypto/rand"
	"github.com/shopspring/decimal"
	"io"
	"math/big"
)

func RandomCSPRNG(n int) []byte {
	buff := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, buff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return buff
}

func CalcPer(a,b *big.Int) (float64,bool) {
	//a,b := big.NewInt(1),big.NewInt(3)
	ad,bd := decimal.NewFromBigInt(a,0),decimal.NewFromBigInt(b,0)
	cd := ad.DivRound(bd,4)
	return cd.Float64()
}