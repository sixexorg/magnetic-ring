package cvm

const (
	// BaseLibName is here for consistency; the base functions have no namespace/library.
	BaseLibName = ""
	// LoadLibName is here for consistency; the loading system has no namespace/library.
	LoadLibName = "package"
	// TabLibName is the name of the table Library.
	TabLibName = "table"
	// IoLibName is the name of the io Library.
	// IoLibName = "io"
	// OsLibName is the name of the os Library.
	// OsLibName = "os"
	// StringLibName is the name of the string Library.
	StringLibName = "string"
	// MathLibName is the name of the math Library.
	MathLibName = "math"
	// DebugLibName is the name of the debug Library.
	DebugLibName = "debug"
	// ChannelLibName is the name of the channel Library.
	ChannelLibName = "channel"
	// CoroutineLibName is the name of the coroutine Library.
	CoroutineLibName = "coroutine"
	// HashLibName is the name of the hash function library.
	HashLibName = "hash"
	// MsgName is the name of the internal transfer information values.
	MsgName = "msg"
	// BlockChainName is the name of the internal blockchain infomation values.
	BlockChainName = "bc"
	// DateTimeName is the name of the date time library.
	DateTimeName = "date"
)

type luaLib struct {
	libName string
	libFunc LGFunction
}

var luaLibs = []luaLib{
	{LoadLibName, OpenPackage},
	{BaseLibName, OpenBase},
	{TabLibName, OpenTable},
	{StringLibName, OpenString},
	{MathLibName, OpenMath},
	{HashLibName, OpenHash},
	{DateTimeName, OpenDateTime},
	{MsgName, OpenMsg},
	{BlockChainName, OpenBlockChain},
	//{DebugLibName, OpenDebug},
}

// OpenLibs loads the built-in libraries. It is equivalent to running OpenLoad,
// then OpenBase, then iterating over the other OpenXXX functions in any order.
func (ls *LState) OpenLibs() {
	// NB: Map iteration order in Go is deliberately randomised, so must open Load/Base
	// prior to iterating.
	for _, lib := range luaLibs {
		ls.Push(ls.NewFunction(lib.libFunc))
		ls.Push(LString(lib.libName))
		ls.Call(1, 0)
	}
}
