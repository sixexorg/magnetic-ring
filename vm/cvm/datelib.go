package cvm

import "time"

func OpenDateTime(l *LState) int {
	mod := l.RegisterModule(DateTimeName, dateTimeFuncs)
	l.Push(mod)
	return 1
}

var dateTimeFuncs = map[string]LGFunction{
	"now": dateTimeNow,
}

func dateTimeNow(l *LState) int {
	l.Push(LNumber(float64(time.Now().Unix())))
	return 1
}
