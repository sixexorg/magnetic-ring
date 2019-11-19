package storelaw

import (
	"fmt"
	"math/big"
	"runtime"
	"testing"
)

func TestVoteState_GetResult(t *testing.T) {
	vs := &VoteState{
		Agree:      big.NewInt(660),
		Against:    big.NewInt(220),
		Abstention: big.NewInt(100),
	}
	vs1 := &VoteState{
		Agree:      big.NewInt(671),
		Against:    big.NewInt(100),
		Abstention: big.NewInt(100),
	}

	total := big.NewInt(1000)
	proportion := int64(67)

	fmt.Println(vs.GetResult(total, proportion))
	fmt.Println(vs1.GetResult(total, proportion))
}

func TestMap(t *testing.T) {
	m := make(map[int]bool)
	for i := 0; i < 1e8; i++ {
		m[i] = true
		delete(m, i)
		if i%100 == 0 {
			printMemStats()
		}
	}
}
func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("HeapAlloc = %v HeapIdel= %v HeapSys = %v  HeapReleased = %v\n", m.HeapAlloc/1024, m.HeapIdle/1024, m.HeapSys/1024, m.HeapReleased/1024)
}
