package validation

import (
	"fmt"
	"math/big"
	"testing"

	"encoding/gob"

	"github.com/sixexorg/magnetic-ring/common"
	"github.com/sixexorg/magnetic-ring/core/mainchain/types"
	"github.com/sixexorg/magnetic-ring/store/orgchain/states"
)

func TestAggregateBonus(t *testing.T) {
	var width, start, end uint64 = 5, 5, 80

	gob.Register(&big.Int{})

	var ed [16]struct {
		aHeight uint64
		balance *big.Int
	}

	assFTree := make([]common.FTreer, 0)
	assFTree = append(assFTree, &states.AccountState{Height: 1, Data: &states.Account{Balance: big.NewInt(1)}})
	assFTree = append(assFTree, &states.AccountState{Height: 8, Data: &states.Account{Balance: big.NewInt(5000)}})
	assFTree = append(assFTree, &states.AccountState{Height: 11, Data: &states.Account{Balance: big.NewInt(20)}})
	assFTree = append(assFTree, &states.AccountState{Height: 20, Data: &states.Account{Balance: big.NewInt(60000)}})
	assFTree = append(assFTree, &states.AccountState{Height: 50, Data: &states.Account{Balance: big.NewInt(3000)}})
	assFTree = append(assFTree, &states.AccountState{Height: 80, Data: &states.Account{Balance: big.NewInt(2222)}})

	aft := common.NewForwardTree(width, start, end, assFTree)

	for i := 0; !aft.Next(); i++ {
		ed[i].aHeight = aft.Key()
		ed[i].balance = aft.Val().(*big.Int)
	}
	bonus := make(map[uint64]uint64)
	bonus[5] = 1
	bonus[10] = 11
	bonus[15] = 111
	bonus[20] = 1111
	bonus[25] = 11111
	bonus[30] = 111111
	bonus[35] = 1111111
	bonus[40] = 111111
	bonus[45] = 11111
	bonus[50] = 1111
	bonus[55] = 111
	bonus[60] = 11
	bonus[65] = 1
	bonus[70] = 11
	bonus[75] = 111
	bonus[80] = 1111

	sum := big.NewInt(0)
	for k, v := range ed {
		b := big.NewInt(0).SetUint64(bonus[start])
		b.Mul(b, v.balance)
		sum.Add(sum, b)
		start += width
		fmt.Printf("No.%d offset:%d height:%d balance:%d rate:%d amount:%d\n", k+1, start, v.aHeight, v.balance.Uint64(), bonus[start], sum.Uint64())
	}

	amount, left := aggregateBonus(assFTree, bonus, 5, 80, 5)
	fmt.Println(amount, left)
}

func TestAppend(t *testing.T) {
	txs := make([]*types.Transaction, 0, 2)
	var tx *types.Transaction
	txs = append(txs, tx)
	fmt.Println(len(txs), tx)
	fmt.Println(txs)
}
