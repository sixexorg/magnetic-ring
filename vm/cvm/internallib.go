package cvm

import (
	"fmt"
)

// baseEmit sends contract internal events
// Function signature emit(...) Indefinite length parameter. Up to 10 parameters, more than 10 parameters are ignored
func baseEmit(l *LState) int {
	return 1
}

// baseRevert returns the local currency of the contract
// function signature revert()
func baseRevert(l *LState) int {
	return 1
}

// baseStore storage field
// function signature store(property)
// property:
// LString
// LNumber
// LTable
// LBool
func baseStore(L *LState) int {
	top := L.GetTop()
	if top != 2 {
		fmt.Println("number of paras must be 2")
	}
	key := L.Get(1).String()
	value := L.Get(2)
	bytes := LVSerialize(value)
	lv := LVDeserialize(bytes)
	fmt.Println(key, lv)
	return 1
}

// baseLoadStorage reads the saved field
// function signature loadStorage(name, defaultValue)
// name: field name, string type
// defaultValue: default value, if you can't read the saved data, replace it with the default value.
func baseLoadStorage(L *LState) int {
	top := L.GetTop()
	if top != 1 {
		fmt.Println("number of paras must be 2")
	}
	key := L.Get(1)

	fmt.Println(key)
	return 1
}
