package account_level

import (
	"math/big"
	"sync"
)

var (
	minM    = big.NewRat(570e8, 1)
	ratU    = big.NewRat(1e8, 1)
	ratC    = big.NewRat(1, 8640)
	ratX    = big.NewRat(1, 1)
	ratN    = big.NewRat(-1, 1)
	dratMax = big.NewRat(16, 100)
	dratL2  = big.NewRat(1, 10)
	dratL1  = big.NewRat(8, 100)
	dratS   = big.NewRat(1, 100)
	ratTen  = big.NewRat(10, 1)

	rat0    = big.NewRat(0, 1)
	rat1    = big.NewRat(1, 10)
	rat2    = big.NewRat(2, 10)
	rat3    = big.NewRat(3, 10)
	rat4    = big.NewRat(4, 10)
	rat5    = big.NewRat(5, 10)
	rat6    = big.NewRat(6, 10)
	rat7    = big.NewRat(7, 10)
	rat8    = big.NewRat(8, 10)
	ratline = big.NewRat(9, 10)
)

type lvlMath struct {
	mu       sync.Mutex
	manure   []*big.Int //store the used energy in one period
	cycleLen int
	lapWidth uint64
	ydayAmt  *big.Int
	tdayAmt  *big.Int
	newest   *big.Int //the gas wait to exec
	height   uint64
}

func newLvlMath(cycleLen int, lapWidth uint64) *lvlMath {
	return &lvlMath{
		manure:   make([]*big.Int, 0, cycleLen*2+1),
		ydayAmt:  big.NewInt(0),
		tdayAmt:  big.NewInt(0),
		newest:   big.NewInt(0),
		cycleLen: cycleLen,
		lapWidth: lapWidth,
		height:   0,
	}
}

func (this *lvlMath) pull(height uint64, usedGas *big.Int) bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	freshManure := big.NewInt(0).Set(usedGas)
	this.newest.Add(this.newest, freshManure)
	this.height = height
	//fmt.Println("💊 💊 💊 ", height, this.lapWidth)
	if height%this.lapWidth == 0 {
		lapGas := new(big.Int).Set(this.newest)
		this.newest.SetInt64(0)
		this.manure = append(this.manure, lapGas)
		l := len(this.manure)
		if l <= this.cycleLen {
			this.ydayAmt.Add(this.ydayAmt, lapGas)
		} else if l <= this.cycleLen*2 {
			this.tdayAmt.Add(this.tdayAmt, lapGas)
		} else {
			this.pop()
			return true
		}
	}
	return false
}

func (this *lvlMath) pop() {
	if len(this.manure) == this.cycleLen*2+1 {
		yRmoved := this.manure[0]
		tRemoved := this.manure[this.cycleLen]
		newest := this.manure[this.cycleLen*2]
		this.manure = this.manure[1:]
		this.ydayAmt.Sub(this.ydayAmt, yRmoved).Add(this.ydayAmt, tRemoved)
		this.tdayAmt.Sub(this.tdayAmt, tRemoved).Add(this.tdayAmt, newest)
	}
}

func (this *lvlMath) r() *big.Rat {
	bt := new(big.Rat).SetInt(this.tdayAmt)
	by := new(big.Rat).SetInt(this.ydayAmt)
	rat := new(big.Rat)
	rat.Sub(bt, by)
	byInv := new(big.Rat).Set(by)
	if byInv.Cmp(rat0) != 0 {
		byInv.Inv(byInv)
	}
	rat.Mul(rat, byInv)

	sign := true
	if rat.Sign() == -1 {
		sign = false
	}
	rat.Abs(rat)
	if rat.Cmp(rat1) == -1 {
		rat.Set(rat0)

	} else if rat.Cmp(rat2) == -1 {
		rat.Set(rat1)
	} else if rat.Cmp(rat3) == -1 {
		rat.Set(rat2)
	} else if rat.Cmp(rat4) == -1 {
		rat.Set(rat3)
	} else if rat.Cmp(rat5) == -1 {
		rat.Set(rat4)
	} else if rat.Cmp(rat6) == -1 {
		rat.Set(rat5)
	} else if rat.Cmp(rat7) == -1 {
		rat.Set(rat6)
	} else if rat.Cmp(rat8) == -1 {
		rat.Set(rat7)
	} else if rat.Cmp(ratline) == -1 {
		rat.Set(rat8)
	} else {
		rat.Set(ratline)
	}
	if !sign {
		rat.Mul(rat, ratN)
	}
	return rat
}

func (this *lvlMath) m(rat *big.Rat) *big.Rat {
	by := new(big.Rat).SetInt(this.ydayAmt)
	m := new(big.Rat)
	m.Add(ratX, rat)
	m.Mul(m, by)
	m.Mul(m, ratC)
	if m.Cmp(minM) == -1 {
		m.Set(minM)
	}
	return m
}

func (this *lvlMath) d(rat *big.Rat) []*big.Rat {
	sign := rat.Sign()
	d1 := new(big.Rat).Set(ratX)
	factor := new(big.Rat)
	if rat.Cmp(rat8) == 1 {
		if sign == 1 {
			factor.Set(dratMax)
		}
	} else {
		d1.Set(ratX)
		factor.Mul(dratL2, rat)
		factor.Sub(dratL1, factor)
	}

	step := this.ratStep(rat)
	ds := make([]*big.Rat, 0, step)
	ds = append(ds, d1)
	offset := 1
	for i := lv2; i <= lv9; i++ {
		if offset >= step {
			break
		}
		dn := new(big.Rat).Add(ds[i-2], factor)
		factor.Sub(factor, dratS)
		ds = append(ds, dn)
		offset++
	}
	return ds
}

func (this *lvlMath) v(rat, m *big.Rat, dsReal []*big.Rat, distribution []*big.Int) []uint64 {
	l := len(dsReal)
	/*fmt.Println("dsreal", dsReal)
	fmt.Println("distribution", distribution)*/
	lvAoumt := make([]*big.Rat, 0, l)
	for i := 0; i < l; i++ {
		//fmt.Println("lvamount", distribution[i])
		lvAoumt = append(lvAoumt, new(big.Rat).SetInt(distribution[i]))
	}
	deno := new(big.Rat)
	for k, v := range dsReal {
		part := new(big.Rat)
		part.Mul(v, lvAoumt[k])
		//fmt.Println("deno add ", part, v, lvAoumt[k])
		deno.Add(deno, part)
	}
	//fmt.Println("deno", deno)
	if deno.Cmp(rat0) != 0 {
		deno.Inv(deno)
	}

	per := make([]uint64, 0, l)
	for _, v := range dsReal {
		part := new(big.Rat)
		part.Mul(v, m)
		part.Mul(part, deno)
		part.Mul(part, ratU)
		rfTmp := big.NewFloat(0).SetRat(part)
		integer, _ := rfTmp.Uint64()
		per = append(per, integer)
	}
	//fmt.Println("️ ", per)
	return per
}

func (this *lvlMath) ratStep(rat *big.Rat) int {
	sign := rat.Sign()
	step := 9
	if sign == 1 {
		mid := new(big.Rat).Set(rat)
		mid.Mul(ratTen, mid)
		num, _ := mid.Float64()
		step = step - int(num)
		if step == 0 {
			step = 1
		}
		return step
	} else {
		return step
	}
}

func (this *lvlMath) reward(height uint64, energyUsed *big.Int, distribution []*big.Int) []uint64 {
	bl := this.pull(height, energyUsed)
	if bl {
		r := this.r()
		m := this.m(r)
		//fmt.Println("💊 💊 💊 💊", m)
		d := this.d(r)
		return this.v(r, m, d, distribution)
	}
	//fmt.Println("📈 bl ", bl)
	return make([]uint64, 0)
}
