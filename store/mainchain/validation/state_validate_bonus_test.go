package validation

import (
	"testing"

	"math/big"

	"fmt"

	"reflect"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/store/mainchain/account_level"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

func TestForwardTreeAss(t *testing.T) {

	f1 := &states.AccountState{Height: 1, Data: &states.Account{Balance: big.NewInt(1)}}
	f2 := &states.AccountState{Height: 3, Data: &states.Account{Balance: big.NewInt(2)}}
	f3 := &states.AccountState{Height: 6, Data: &states.Account{Balance: big.NewInt(3)}}
	f4 := &states.AccountState{Height: 11, Data: &states.Account{Balance: big.NewInt(4)}}
	f5 := &states.AccountState{Height: 23, Data: &states.Account{Balance: big.NewInt(5)}}
	f6 := &states.AccountState{Height: 24, Data: &states.Account{Balance: big.NewInt(6)}}
	f7 := &states.AccountState{Height: 37, Data: &states.Account{Balance: big.NewInt(7)}}
	f8 := &states.AccountState{Height: 39, Data: &states.Account{Balance: big.NewInt(8)}}
	fs := []common.FTreer{f1, f2, f3, f4, f5, f6, f7, f8}

	ft := common.NewForwardTree(1, 1, 6, fs[0:1])
	for !ft.Next() {
		fmt.Println(ft.Key(), ft.Val().(*big.Int), reflect.TypeOf(ft.Val()))
	}
}
func TestForwardTreeLvl(t *testing.T) {
	f1 := &account_level.HeightLevel{Height: 1, Lv: account_level.EasyLevel(1)}
	f2 := &account_level.HeightLevel{Height: 3, Lv: account_level.EasyLevel(2)}
	f3 := &account_level.HeightLevel{Height: 6, Lv: account_level.EasyLevel(3)}
	f4 := &account_level.HeightLevel{Height: 11, Lv: account_level.EasyLevel(4)}
	f5 := &account_level.HeightLevel{Height: 23, Lv: account_level.EasyLevel(5)}
	f6 := &account_level.HeightLevel{Height: 24, Lv: account_level.EasyLevel(6)}
	f7 := &account_level.HeightLevel{Height: 37, Lv: account_level.EasyLevel(7)}
	f8 := &account_level.HeightLevel{Height: 39, Lv: account_level.EasyLevel(8)}
	fs := []common.FTreer{f1, f2, f3, f4, f5, f6, f7, f8}

	ft := common.NewForwardTree(5, 0, 48, fs)
	for !ft.Next() {
		fmt.Println(ft.Val(), reflect.TypeOf(ft.Val()))
	}
}

func TestAggregateBonus(t *testing.T) {
	var width, start, end uint64 = 5, 1, 80

	/*	var ed [16]struct {
		aHeight uint64
		lHeight uint64
		balance *big.Int
		level   account_level.EasyLevel
	}*/

	assFTree := make([]common.FTreer, 0)
	assFTree = append(assFTree, &states.AccountState{Height: 1, Data: &states.Account{Balance: big.NewInt(1)}})
	assFTree = append(assFTree, &states.AccountState{Height: 8, Data: &states.Account{Balance: big.NewInt(5000)}})
	assFTree = append(assFTree, &states.AccountState{Height: 11, Data: &states.Account{Balance: big.NewInt(20)}})
	assFTree = append(assFTree, &states.AccountState{Height: 20, Data: &states.Account{Balance: big.NewInt(60000)}})
	assFTree = append(assFTree, &states.AccountState{Height: 50, Data: &states.Account{Balance: big.NewInt(3000)}})
	assFTree = append(assFTree, &states.AccountState{Height: 80, Data: &states.Account{Balance: big.NewInt(2222)}})

	//aft := common.NewForwardTree(width, start, end, assFTree)

	/*	for i := 0; !aft.Next(); i++ {
		ed[i].aHeight = aft.Key()
		ed[i].balance = aft.Val().(*big.Int)
	}*/

	lvlFTree := make([]common.FTreer, 0)
	lvlFTree = append(lvlFTree, &account_level.HeightLevel{Height: 1, Lv: account_level.EasyLevel(2)})
	lvlFTree = append(lvlFTree, &account_level.HeightLevel{Height: 13, Lv: account_level.EasyLevel(4)})
	lvlFTree = append(lvlFTree, &account_level.HeightLevel{Height: 17, Lv: account_level.EasyLevel(5)})
	lvlFTree = append(lvlFTree, &account_level.HeightLevel{Height: 58, Lv: account_level.EasyLevel(6)})
	lvlFTree = append(lvlFTree, &account_level.HeightLevel{Height: 77, Lv: account_level.EasyLevel(2)})

	/*lft := common.NewForwardTree(width, start, end, lvlFTree)
	for i := 0; !lft.Next(); i++ {
		ed[i].lHeight = lft.Key()
		ed[i].level = lft.Val().(account_level.EasyLevel)
	}*/

	bonus := make(map[uint64][]uint64)
	bonus[5] = []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	bonus[10] = []uint64{11, 22, 33, 44, 55, 66, 77, 88, 99}
	bonus[15] = []uint64{111, 222, 333, 444, 555, 666, 777, 888, 999}
	bonus[20] = []uint64{1111, 2222, 3333, 4444, 5555, 6666, 7777, 8888, 9999}
	bonus[25] = []uint64{11111, 22222, 33333, 44444, 55555, 66666, 77777, 88888, 99999}
	bonus[30] = []uint64{111111, 222222, 333333, 444444, 555555, 666666, 777777, 888888, 999999}
	bonus[35] = []uint64{1111111, 2222222, 3333333, 4444444, 5555555, 6666666, 7777777, 8888888, 9999999}
	bonus[40] = []uint64{111111, 222222, 333333, 444444, 555555, 666666, 777777, 888888, 999999}
	bonus[45] = []uint64{11111, 22222, 33333, 44444, 55555, 66666, 77777, 88888, 99999}
	bonus[50] = []uint64{1111, 2222, 3333, 4444, 5555, 6666, 7777, 8888, 9999}
	bonus[55] = []uint64{111, 222, 333, 444, 555, 666, 777, 888, 999}
	bonus[60] = []uint64{11, 22, 33, 44, 55, 66, 77, 88, 99}
	bonus[65] = []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	bonus[70] = []uint64{11, 22, 33, 44, 55, 66, 77, 88, 99}
	bonus[75] = []uint64{111, 222, 333, 444, 555, 666, 777, 888, 999}
	bonus[80] = []uint64{1111, 2222, 3333, 4444, 5555, 6666, 7777, 8888, 9999}

	/*for k, v := range ed {
		idx := int(v.level) - 1
		amount := big.NewInt(0)
		rat := uint64(0)
		if idx > 0 {
			rat = bonus[start][idx]
			amount = big.NewInt(0).SetUint64(rat)
			amount.Mul(amount, v.balance)
		}
		fmt.Printf("No.%d aheight:%d lheight:%d balance:%d level:%d rate:%d amount:%d\n", k+1, v.aHeight, v.lHeight, v.balance.Uint64(), v.level, rat, amount.Uint64())
		start += width
	}*/

	amount, left := aggregateBonus(assFTree, lvlFTree, bonus, 0, end, width)
	fmt.Println(amount, left)
	fmt.Println(start, width, end)
}

func TestDeepCopy(t *testing.T) {
	sp := &common.Span{
		L:   1,
		R:   2,
		Val: 20,
	}
	sp1 := &common.Span{}
	common.DeepCopy(&sp1, sp)
	fmt.Println(sp1)
}
