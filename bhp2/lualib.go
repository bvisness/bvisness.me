package bhp2

import (
	"bytes"
	"embed"
	_ "embed"

	"github.com/bvisness/bvisness.me/utils"
	lua "github.com/yuin/gopher-lua"
)

//go:embed lua/*
var builtins embed.FS

func LoadBHP2(l *lua.LState) int {
	bhpSource := utils.Must1(builtins.ReadFile("lua/bhp.lua"))
	bhp := utils.Must1(l.Load(bytes.NewBuffer(bhpSource), "bhp"))
	l.Push(bhp)
	l.Call(0, 1)
	return 1
}

func changeSearchers(l *lua.LState) {
	p := l.GetGlobal("package")
	oldLoaders := l.GetField(p, "loaders").(*lua.LTable)
	preloadSearcher := oldLoaders.RawGetInt(1)
	embeddedSearcher := l.NewFunction(func(l *lua.LState) int {
		name := l.CheckString(1)
		b, err := builtins.ReadFile("lua/" + name + ".lua")
		if err != nil {
			l.RaiseError("error opening builtin file: %v", err)
			return 0
		}
		loader, err := l.Load(bytes.NewBuffer(b), name)
		if err != nil {
			l.RaiseError("error in builtin file: %v", err)
			return 0
		}
		l.Push(loader)
		return 1
	})

	newLoaders := l.NewTable()
	l.RawSetInt(newLoaders, 1, preloadSearcher)
	l.RawSetInt(newLoaders, 2, embeddedSearcher)

	l.SetField(p, "loaders", newLoaders)
	l.SetField(l.Get(lua.RegistryIndex), "_LOADERS", newLoaders)
}

func getRendered(l *lua.LState) string {
	bhp := l.GetGlobal("bhp").(*lua.LTable)
	return string(bhp.RawGetString("_rendered").(lua.LString))
}

func setSource(l *lua.LState, source string) {
	bhp := l.GetGlobal("bhp").(*lua.LTable)
	bhp.RawSetString("_source", lua.LString(source))
}
