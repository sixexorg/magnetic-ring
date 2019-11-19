package account_level

import (
	"fmt"
	"math/big"
	"testing"
)

func TestAA(t *testing.T) {
	l, _ := big.NewInt(0).SetString("100000000000000000000000000", 10)
	r := big.NewInt(30)

	/*lf := big.NewFloat(0).SetInt(l)
	rf := big.NewFloat(0).SetInt(r)*/

	//rat := big.NewRat(0, 1).SetFrac(l, r)

	yl := new(big.Rat).SetInt(l)
	tl := new(big.Rat).SetInt(r)
	fmt.Println(yl, tl)
	fmt.Println(tl.Sub(tl, yl))
	rat := big.NewRat(18, 10)

	m := new(big.Rat)
	rTmp := m.Mul(yl, rat).Mul(m, big.NewRat(1, 8640))

	rfTmp := big.NewFloat(0).SetRat(rTmp)
	rfTmp.Uint64()
	fmt.Println("float string:", rfTmp.Prec())

	b, oc := rfTmp.Int(big.NewInt(0))
	fmt.Println(rTmp)
	fmt.Println(b, oc)
	fmt.Println(30000 * 10000000000 * 8640 * 2)

}
func TestPer(t *testing.T) {
	dis := []*big.Int{
		big.NewInt(1000000),
		big.NewInt(2000000),
		big.NewInt(3000000),
		big.NewInt(2000000),
		big.NewInt(1000000),
		big.NewInt(2000000),
		big.NewInt(3000000),
		big.NewInt(2000000),
		big.NewInt(1000000),
	}
	math := newLvlMath(1, 1)
	bl := math.pull(1, big.NewInt(25000))
	if bl {
		r := math.r()
		m := math.m(r)
		d := math.d(r)
		amouts := math.v(r, m, d, dis)
		fmt.Println("No.1:", r, m, d, amouts)
	}

	bl = math.pull(2, big.NewInt(30000))
	if bl {
		r := math.r()
		m := math.m(r)
		d := math.d(r)
		amouts := math.v(r, m, d, dis)
		fmt.Println("No.2:", r, m, d, amouts)
	}
	bl = math.pull(3, big.NewInt(40000))
	if bl {
		r := math.r()
		m := math.m(r)
		d := math.d(r)
		amouts := math.v(r, m, d, dis)
		fmt.Println("No.3:", r, m, d, amouts)
	}
	bl = math.pull(4, big.NewInt(50000))
	if bl {
		r := math.r()
		m := math.m(r)
		d := math.d(r)
		amouts := math.v(r, m, d, dis)
		fmt.Println("No.4:", r, m, d, amouts)
	}
	bl = math.pull(5, big.NewInt(60100))
	if bl {
		r := math.r()
		m := math.m(r)
		d := math.d(r)
		amouts := math.v(r, m, d, dis)
		fmt.Println("No.5:", r, m, d, amouts)
	}
	bl = math.pull(6, big.NewInt(50000))
	if bl {
		r := math.r()
		m := math.m(r)
		d := math.d(r)
		amouts := math.v(r, m, d, dis)
		fmt.Println("No.6:", r, m, d, amouts)
	}
}
