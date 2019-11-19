package cvm


func OpenMsg(l *LState) int {
	msg := l.NewTable()
	for key, function := range msgFuncs {
		msg.RawSetString(key, l.NewFunction(function))
	}
	l.SetGlobal("msg", msg)
	return 1
}

var msgFuncs = map[string]LGFunction{
	"transfer":  msgTransfer,
	"signature": msgSignature,
	"value":     msgValue,
	"sender":    msgSender,
}

func msgTransfer(l *LState) int {
	return 1
}

func msgSignature(l *LState) int {
	return 1
}

func msgSender(l *LState) int {
	return 1
}

func msgValue(l *LState) int {
	return 1
}
