package account_level

import (
	//"encoding/gob"
	"math/big"
)

const (
	lv0 EasyLevel = iota
	lv1
	lv2
	lv3
	lv4
	lv5
	lv6
	lv7
	lv8
	lv9
)

func init() {
	//gob.Register(lv0)
}

var (
	lv1_b = big.NewInt(1e8)
	lv2_b = big.NewInt(1e11)
	lv3_b = big.NewInt(2e11)
	lv4_b = big.NewInt(4e11)
	lv5_b = big.NewInt(8e11)
	lv6_b = big.NewInt(16e11)
	lv7_b = big.NewInt(32e11)
	lv8_b = big.NewInt(64e11)
	lv9_b = big.NewInt(128e11)

	lvPrec = new(big.Rat).SetInt(lv1_b)
)

//uint8 [0,255]
//hundreds: 0->next cycle ,1->the cycel after next
//tens:next level
//ones:the level match the balance， ones == 0 case tens is the real level
type AccountLevel uint8
type EasyLevel uint8

func newAccountLevel(immediately bool, next, tagert EasyLevel) AccountLevel {
	if next > 9 || tagert > 9 {
		return 0
	}
	hundred := EasyLevel(1)
	if immediately {
		hundred = 0
	}
	lv := AccountLevel(hundred*100 + next*10 + tagert)
	return lv
}
func (al AccountLevel) Decode() (immediately bool, left, tagert, curLv EasyLevel) {
	immediately = al/100 == 0
	left = EasyLevel(al % 100 / 10)
	tagert = EasyLevel(al % 10)
	curLv = left
	if tagert > 0 {
		curLv--
	}
	return
}

type LvlItem struct {
	Lvl uint8
	Min *big.Int
}

func rankLevel(crystal *big.Int) (level EasyLevel) {
	if crystal.Cmp(lv1_b) == -1 {
		return 0
	} else if crystal.Cmp(lv2_b) == -1 {
		return 1
	} else if crystal.Cmp(lv3_b) == -1 {
		return 2
	} else if crystal.Cmp(lv4_b) == -1 {
		return 3
	} else if crystal.Cmp(lv5_b) == -1 {
		return 4
	} else if crystal.Cmp(lv6_b) == -1 {
		return 5
	} else if crystal.Cmp(lv7_b) == -1 {
		return 6
	} else if crystal.Cmp(lv8_b) == -1 {
		return 7
	} else if crystal.Cmp(lv9_b) == -1 {
		return 8
	}
	return 9
}

type LvlAmount struct {
	Lv     AccountLevel
	Amount *big.Int
}
