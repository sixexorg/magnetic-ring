package common_test

import (
	"fmt"
	"testing"

	"math/big"

	"github.com/sixexorg/magnetic-ring/common"
)

type forwardMock struct {
	height uint64
	val    int
}

func (this *forwardMock) GetHeight() uint64 {
	return this.height
}
func (this *forwardMock) GetVal() interface{} {
	return this.val
}

func TestForwarTree(t *testing.T) {
	f1 := &forwardMock{1, 1}
	f2 := &forwardMock{3, 2}
	f3 := &forwardMock{6, 3}
	f4 := &forwardMock{11, 4}
	f5 := &forwardMock{23, 5}
	f6 := &forwardMock{24, 6}
	f7 := &forwardMock{37, 7}
	f8 := &forwardMock{39, 8}
	fs := []common.FTreer{f1, f2, f3, f4, f5, f6, f7, f8}

	ft := common.NewForwardTree(5, 5, 48, fs)
	for !ft.Next() {
		fmt.Println(ft.Key(), ft.Val())
	}

	fmt.Println("----------------------------------------")
	f11 := &forwardMock{1, 1}
	f22 := &forwardMock{8, 5000}
	f33 := &forwardMock{11, 20}
	f44 := &forwardMock{21, 60000}
	fs1 := []common.FTreer{f11, f22, f33, f44}

	ft1 := common.NewForwardTree(5, 0, 24, fs1)
	for !ft1.Next() {
		fmt.Println(ft1.Key(), ft1.Val())
	}
}

func TestDeepCopy(t *testing.T) {
	sp := &common.Span{
		L:   1,
		R:   2,
		Val: big.NewInt(20),
	}
	sp1 := &common.Span{}
	err := common.DeepCopy(&sp1, sp)
	fmt.Println(sp1, err)
}
