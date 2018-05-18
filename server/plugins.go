package main

import (
	"github.com/yuin/gopher-lua"
)

var L *lua.LState

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
func Load(file, method string) lua.LValue {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoFile(file); err != nil {
		panic(err)
	}
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal(method),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(3)); err != nil {
		panic(err)
	}
	ret := L.Get(-1) // returned value
	L.Pop(1)         // remove received value
	return ret
}

// Call will call a Lua method in a loaded plugin.
func Call(function string) (lua.LValue, error) {
	var luaFunc lua.LValue
	luaFunc = L.GetGlobal(function)
	err := L.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    1,
		Protect: true,
	}, nil)
	ret := L.Get(-1) // returned value
	if ret.String() != "nil" {
		L.Pop(1) // remove received value
	}
	return ret, err
}
