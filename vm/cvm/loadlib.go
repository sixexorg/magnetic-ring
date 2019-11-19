package cvm

import (
	"fmt"
)

/* load lib {{{ */

var loLoaders = []LGFunction{loLoaderPreload}

func OpenPackage(L *LState) int {
	packagemod := L.RegisterModule(LoadLibName, loFuncs)

	L.SetField(packagemod, "preload", L.NewTable())

	loaders := L.CreateTable(len(loLoaders), 0)
	for i, loader := range loLoaders {
		L.RawSetInt(loaders, i+1, L.NewFunction(loader))
	}
	L.SetField(packagemod, "loaders", loaders)
	L.SetField(L.Get(RegistryIndex), "_LOADERS", loaders)

	loaded := L.NewTable()
	L.SetField(packagemod, "loaded", loaded)
	L.SetField(L.Get(RegistryIndex), "_LOADED", loaded)

	L.SetField(packagemod, "path", emptyLString)
	L.SetField(packagemod, "cpath", emptyLString)

	L.Push(packagemod)
	return 1
}

var loFuncs = map[string]LGFunction{}

func loLoaderPreload(L *LState) int {
	name := L.CheckString(1)
	preload := L.GetField(L.GetField(L.Get(EnvironIndex), "package"), "preload")
	if _, ok := preload.(*LTable); !ok {
		L.RaiseError("package.preload must be a table")
	}
	lv := L.GetField(preload, name)
	if lv == LNil {
		L.Push(LString(fmt.Sprintf("no field package.preload['%s']", name)))
		return 1
	}
	L.Push(lv)
	return 1
}

/* }}} */

//
