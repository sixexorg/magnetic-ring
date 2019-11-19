package common

import (
	"fmt"
	"testing"
)

func TestCheckLeagueName(t *testing.T) {
	name := "ABD1E"
	fmt.Println(CheckLSymbol(name))

	fmt.Println(String2Symbol(name))
}
