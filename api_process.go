package main

import (
	rt "github.com/arnodel/golua/runtime"
	"github.com/arnodel/golua/lib/packagelib"
)

var processLoader = packagelib.Loader{
	Name: "process",
	Load: processLoad,
}

func processLoad(rtm *rt.Runtime) (rt.Value, func()) {
	exports := map[string]luaExport{}
	mod := rt.NewTable()
	setExports(rtm, mod, exports)

	return rt.TableValue(mod), nil
}
