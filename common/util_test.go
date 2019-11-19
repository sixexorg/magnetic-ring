package common

import (
	"fmt"
	"math"
	"testing"
)

func TestRandomCSPRNG(t *testing.T) {
	for i := 2; i < 9; i++ {
		f64 := math.Pow(2, float64(i))
		buf := RandomCSPRNG(int(f64))

		fmt.Printf("len=%d,buf=%x\n", int(f64), buf)

	}

}

func TestUint16(t *testing.T) {
	pkbf := "04a21186f94eb8b037c2e304a86a333952c898ee7eca0de0184b4728e78d5cf7a998b032034ec5074a23ab8c86c1aa071fab3d9c55eb2ad79a8377b24774981356"

	buf, _ := Hex2Bytes(pkbf)

	u16 := BytesToUint16(buf)

	fmt.Printf("u16=%v\n", u16)

	newbuf := Uint16ToBytes(u16)

	fmt.Printf("qqq-->%x\n", newbuf)

}
