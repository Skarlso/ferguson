package main

import (
	"github.com/yuin/gopher-lua"
)

// L is the Lua vm's state.
var L *lua.LState

func init() {
	L = lua.NewState()
}

// // LoadFile loads a lua file
// func LoadFile(module string, file string, data string) error {
// 	pluginDef := "local P = {};" + module + " = P;setmetatable(" + module + ", {__index = _G});setfenv(1, P);"

// 	if fn, err := L.Load(strings.NewReader(pluginDef+data), file); err != nil {
// 		return err
// 	} else {
// 		L.Push(fn)
// 		return L.PCall(0, lua.MultRet, nil)
// 	}
// }

// func LoadPlugins() {
// 	LoadFile("test")
// }

// Load runs a lua script.
func Load(file string) {
	if err := L.DoFile(file); err != nil {
		// TODO: This needs to be replaced by an error so we ignore plugins that could not be loaded.
		panic(err)
	}
}

// Call will call a Lua method in a loaded plugin.
func Call(function string) (lua.LValue, error) {
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal(function),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(3)); err != nil {
		panic(err)
	}
	ret := L.Get(-1) // returned value
	L.Pop(1)         // remove received value
	return ret, nil
}
