package bhp2

import (
	"bytes"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

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

type FSSearcher struct {
	FS     fs.FS
	Prefix string // Path prefix to prepend to `require` string before lookup, e.g. `lua/`
}

var builtinFSSearcher = FSSearcher{
	FS:     builtins,
	Prefix: "lua/",
}

func loadPlainLua(l *lua.LState, s FSSearcher) bool {
	name := l.CheckString(1)
	filename := s.Prefix + name + ".lua"
	f, err := s.FS.Open(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error opening file: %v", err)))
		return true
	}

	b, err := io.ReadAll(f)
	if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error reading file: %v", err)))
		return true
	}

	loader, err := l.Load(bytes.NewBuffer(b), name)
	if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error in file: %v", err)))
		return true
	}

	l.Push(loader)
	return true
}

func loadLuax(l *lua.LState, s FSSearcher) bool {
	name := l.CheckString(1)
	filename := s.Prefix + name + ".luax"
	f, err := s.FS.Open(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error opening file: %v", err)))
		return true
	}

	b, err := io.ReadAll(f)
	if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error reading file: %v", err)))
		return true
	}

	// Save file source in bhp._sources
	bhpSources := l.GetGlobal("bhp").(*lua.LTable).RawGetString("_sources").(*lua.LTable)
	bhpSources.RawSetString(filename, lua.LString(b))

	transpiled, err := Transpile(string(b), filename)
	if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error transpiling file: %v", err)))
		return true
	}

	loader, err := l.Load(strings.NewReader(transpiled), name)
	if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error in file: %v", err)))
		return true
	}

	l.Push(loader)
	return true
}

func (b *Instance) initSearchers(l *lua.LState) {
	p := l.GetGlobal("package")
	oldLoaders := l.GetField(p, "loaders").(*lua.LTable)
	preloadSearcher := oldLoaders.RawGetInt(1)

	newLoaders := l.NewTable()
	l.RawSetInt(newLoaders, 1, preloadSearcher)

	fsSearchers := append(b.FSSearchers, builtinFSSearcher)
	for i := 1; i <= len(fsSearchers); i++ {
		s := fsSearchers[i-1]
		l.RawSetInt(newLoaders, i+1, l.NewFunction(func(l *lua.LState) int {
			if loadPlainLua(l, s) || loadLuax(l, s) {
				return 1
			}

			l.Push(lua.LNil)
			return 1
		}))
	}

	l.SetField(p, "loaders", newLoaders)
	l.SetField(l.Get(lua.RegistryIndex), "_LOADERS", newLoaders)
}

func getRendered(l *lua.LState) string {
	bhp := l.GetGlobal("bhp").(*lua.LTable)
	return string(bhp.RawGetString("_rendered").(lua.LString))
}

func saveSource(l *lua.LState, filename, source string) {
	sources := l.GetGlobal("bhp").(*lua.LTable).RawGetString("_sources").(*lua.LTable)
	sources.RawSetString(filename, lua.LString(source))
}
