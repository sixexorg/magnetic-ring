package account_level

import (
	"testing"

	"math/big"

	"fmt"

	"os"

	"github.com/sixexorg/magnetic-ring/common/sink"
	"github.com/sixexorg/magnetic-ring/mock"
	"github.com/sixexorg/magnetic-ring/store/mainchain/states"
)

func TestLevelManager(t *testing.T) {
	dbDir := "./test/"
	defer os.RemoveAll(dbDir)
	lm, err := NewLevelManager(1, 1, dbDir)
	if err != nil {
		t.Fail()
		t.Error(err)
		return
	}
	asmap := getAccountStates()
	destroy := big.NewInt(50000)
	for i := uint64(1); i <= uint64(len(asmap)); i++ {
		lm.ReceiveAccountStates(asmap[i][2:3], destroy, i)
		destroy.Mul(destroy, big.NewInt(10))
		destroy.Div(destroy, big.NewInt(9))
		fmt.Printf("----------------------------------loop %d------------------------------\n", i)
		for k, v := range lm.nextAccLvl {
			im, l, r, cur := v.Lv.Decode()
			fmt.Printf("account:%s amount:%d im:%v l:%d r:%d cur:%d\n", k.ToString(), v.Amount, im, l, r, cur)
		}
		for j := EasyLevel(1); j <= lv9; j++ {
			fmt.Printf("lv %d count:%d amount:%d \n", j, lm.distribution[j].Count, lm.distribution[j].Amount.Uint64())
		}
		fmt.Println(lm.getLvAmountDistribution())
		bs, err := lm.GetNextHeaderProperty(i + 1)
		if err != nil {
			t.Log(err)
		}
		t.Logf("b%d %v", i+1, bs)
	}
	bs, err := lm.GetNextHeaderProperty(5)
	if err != nil {
		t.Log(err)
	}
	t.Log("bs1:", bs)
	bs, err = lm.GetNextHeaderProperty(11)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("bs2:", bs)

	b := big.NewInt(0).SetUint64(bs[3])
	b.Mul(b, big.NewInt(4.5e11))
	b.Div(b, big.NewInt(1e8))
	fmt.Println(b)
}

func getAccountStates() map[uint64]states.AccountStates {

	asmap := make(map[uint64]states.AccountStates, 10)

	ass1 := states.AccountStates{}
	ass1 = append(ass1,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(1.5e8),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(3.5e12),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(3.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(6.5e11),
			},
		},
	)
	asmap[1] = ass1

	ass2 := states.AccountStates{}
	ass2 = append(ass2,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(0.5e8),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.5e10),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(6.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(128.5e11),
			},
		},
	)
	asmap[2] = ass2

	ass3 := states.AccountStates{}
	ass3 = append(ass3,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(30e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(2.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(8.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(128.5e11),
			},
		},
	)
	asmap[3] = ass3

	ass4 := states.AccountStates{}
	ass4 = append(ass4,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(1000),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.e10),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(9.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(129e11),
			},
		},
	)

	asmap[4] = ass4
	ass5 := states.AccountStates{}
	ass5 = append(ass5,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(500),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.9e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(4.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(130e11),
			},
		},
	)
	asmap[5] = ass5
	ass6 := states.AccountStates{}
	ass6 = append(ass6,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(500),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.9e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(4.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(130e11),
			},
		},
	)
	asmap[6] = ass6
	ass7 := states.AccountStates{}
	ass7 = append(ass7,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(500),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.9e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(4.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(130e11),
			},
		},
	)
	asmap[7] = ass7
	ass8 := states.AccountStates{}
	ass8 = append(ass8,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(500),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.9e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(4.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(130e11),
			},
		},
	)
	asmap[8] = ass8
	ass9 := states.AccountStates{}
	ass9 = append(ass9,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(500),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.9e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(4.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(130e11),
			},
		},
	)
	asmap[9] = ass9
	ass10 := states.AccountStates{}
	ass10 = append(ass10,
		&states.AccountState{
			Address: mock.Address_1,
			Data: &states.Account{
				Balance: big.NewInt(500),
			},
		},
		&states.AccountState{
			Address: mock.Address_2,
			Data: &states.Account{
				Balance: big.NewInt(1.9e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_3,
			Data: &states.Account{
				Balance: big.NewInt(4.5e11),
			},
		},
		&states.AccountState{
			Address: mock.Address_4,
			Data: &states.Account{
				Balance: big.NewInt(130e11),
			},
		},
	)
	asmap[10] = ass10
	return asmap
}

func TestDecode(t *testing.T) {
	sk := sink.NewZeroCopySink(nil)
	sk.WriteUint8(uint8(10))
	h := uint8(sk.Bytes()[0])
	fmt.Println(h)
	a := []uint64{1}
	fmt.Println(len(a), cap(a))
}
