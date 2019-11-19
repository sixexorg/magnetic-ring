package cvm

func OpenBlockChain(l *LState) int {
	return 1
}

var bcFuncs = map[string]LGFunction {
	"coinbase": getCoinbase,
	"difficulty": getDifficulty,
}

func getCoinbase(l *LState) int {
	return 1
}

func getDifficulty(l *LState) int {
	return 1
}