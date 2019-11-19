package common

import (
	"bytes"
	"fmt"
	"regexp"
)

const (
	namePattern = "^[A-Z1-9]{5}$"
	line        = "_"
	prefix      = "BX"
	SymbolLen   = 5
)

type Symbol [SymbolLen]byte

func String2Symbol(symbol string) (Symbol, bool) {
	if !CheckLSymbol(symbol) {
		return Symbol{}, false
	}
	buff := new(bytes.Buffer)
	buff.WriteString(symbol)
	return Bytes2Symbol(buff.Bytes()), true
}

func CheckLSymbol(name string) bool {
	bl, err := regexp.MatchString(namePattern, name)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return bl
}

func (this Symbol) String() string {
	buff := new(bytes.Buffer)
	buff.WriteString(prefix)
	buff.WriteString(line)
	buff.Write(this[:])
	return buff.String()
}

func Bytes2Symbol(data []byte) Symbol {
	var ls Symbol
	copy(ls[:], data[:])
	return ls
}
