package main

import (
	"log"

	"github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// L is the Lua vm's state.
var L *lua.LState

func init() {
	L = lua.NewState()
}

// Load runs a lua script.
func Load(file string) {
	if err := L.DoFile(file); err != nil {
		log.Printf("file '%s' could not be loaded. reason: %v\n", file, err)
	}
}

// Call will call a Lua method in a loaded plugin.
func Call(function string, args ...interface{}) (lua.LValue, error) {
	var luaArgs []lua.LValue
	for _, v := range args {
		luaArgs = append(luaArgs, luar.New(L, v))
	}
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal(function),
		NRet:    1,
		Protect: true,
	}, luaArgs...); err != nil {
		panic(err)
	}
	ret := L.Get(-1) // returned value
	L.Pop(1)         // remove received value
	return ret, nil
}
