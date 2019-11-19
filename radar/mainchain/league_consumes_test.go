package mainchain

import (
	"fmt"
	"testing"
)

func TestAAA(t *testing.T) {
	c := make(chan bool, 1)
	h := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-c:
				h <- false
				fmt.Println("break")
				break
			}
			fmt.Println("234")

		}
	}()
	c <- true
	fmt.Println(<-h)
	select {}
}

func TestMap(t *testing.T) {
	m := make(map[int]bool)
	m[1] = true
	if !m[2] {
		fmt.Println("123")
	}
}
