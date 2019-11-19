package comm

const BroadCast  = "ALL"

type ExcuteType uint8

const (
	MinExcuteType ExcuteType = iota
	StartNewHeight
	EarthProcess
	MaxExcuteType
)

type BftActionType uint8

const (
	SealBlock BftActionType = iota
	FastForward // for syncer catch up
	ReBroadcast
)

const (
	V_One   = 1
	V_Two   = 2
	V_Three = 3
)
