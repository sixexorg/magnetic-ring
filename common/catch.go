package common

import (
	"fmt"
)

func CatchPanic() {
	if err := recover(); err != nil {
		fmt.Println("❌ panic recover from ", err)
	}
}
