package bhp

import (
	"bytes"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/bvisness/bvisness.me/utils"
	lua "github.com/yuin/gopher-lua"
)

// Files with these extensions can be loaded raw from FSSearchers.
var staticFileExts = []string{".svg"}

//go:embed lua/*
var builtins embed.FS

func LoadbhpLib(l *lua.LState) int {
	bhpSource := utils.Must1(builtins.ReadFile("lua/bhp.lua"))
	bhp := utils.Must1(l.Load(bytes.NewBuffer(bhpSource), "bhp"))
	l.Push(bhp)
	l.Call(0, 1)
	return 1
}

func LoadLuaX(l *lua.LState, filename, source string) (*lua.LFunction, error) {
	filename = strings.ReplaceAll(filename, "\\", "/")

	transpiled, err := Transpile(source, filename)
	if err != nil {
		return nil, fmt.Errorf("error transpiling file: %w", err)
	}

	loader, err := l.Load(strings.NewReader(transpiled), filename)
	if err != nil {
		return nil, fmt.Errorf("error in file: %w", err)
	}

	sources := l.GetGlobal("bhp").(*lua.LTable).RawGetString("_sources").(*lua.LTable)
	sources.RawSetString(filename, lua.LString(source))

	return loader, nil
}

//
// Searchers (package.loaders)
//

type FSSearcher struct {
	FS     fs.FS
	Prefix string // Path prefix to prepend to `require` string before lookup, e.g. `lua/`
}

var builtinFSSearcher = FSSearcher{
	FS:     builtins,
	Prefix: "lua/",
}

func searchPlainLua(l *lua.LState, s FSSearcher) bool {
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

func searchLuaX(l *lua.LState, s FSSearcher) bool {
	name := l.CheckString(1)
	filename := s.Prefix + name + ".luax"
	f, err := s.FS.Open(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error opening file: %v", err)))
		return true
	}

	source, err := io.ReadAll(f)
	if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error reading file: %v", err)))
		return true
	}

	loader, err := LoadLuaX(l, filename, string(source))
	if err != nil {
		l.Push(lua.LString(err.Error()))
		return true
	}

	l.Push(loader)
	return true
}

func searchStaticFile(l *lua.LState, s FSSearcher) bool {
	name := l.CheckString(1)
	for _, ext := range staticFileExts {
		if strings.HasSuffix(name, ext) {
			f, err := s.FS.Open(name)
			if errors.Is(err, fs.ErrNotExist) {
				return false
			} else if err != nil {
				l.Push(lua.LString(fmt.Sprintf("error reading file: %v", err)))
				return true
			}

			var loader lua.LGFunction = func(l *lua.LState) int {
				contents, err := io.ReadAll(f)
				if err != nil {
					RaiseMsg(l, err, "failed to read static file")
				}

				l.Push(lua.LString(contents))
				return 1
			}
			l.Push(l.NewClosure(loader))
			return true
		}
	}

	return false
}

func (b *Instance) initSearchers(l *lua.LState, r *http.Request) {
	p := l.GetGlobal("package")
	oldLoaders := l.GetField(p, "loaders").(*lua.LTable)
	preloadSearcher := oldLoaders.RawGetInt(1).(*lua.LFunction)

	newLoaders := l.NewTable()
	i := 0

	// Add the original preload searcher
	i++
	l.RawSetInt(newLoaders, i, preloadSearcher)

	// Add Go library searcher
	i++
	l.RawSetInt(newLoaders, i, l.NewFunction(func(l *lua.LState) int {
		name := l.CheckString(1)
		if lib, ok := b.Libs[name]; ok {
			l.Push(l.NewFunction(func(l *lua.LState) int {
				return lib(l, b, r)
			}))
			return 1
		} else {
			l.Push(lua.LNil)
			return 1
		}
	}))

	// Add searcher functions for FSSearchers
	fsSearchers := append(b.FSSearchers, builtinFSSearcher)
	for _, s := range fsSearchers {
		s := s // >:(

		i++
		l.RawSetInt(newLoaders, i, l.NewFunction(func(l *lua.LState) int {
			if searchPlainLua(l, s) || searchLuaX(l, s) {
				return 1
			}

			if searchStaticFile(l, s) {
				return 1
			}

			l.Push(lua.LNil)
			return 1
		}))
	}

	l.SetField(p, "loaders", newLoaders)
	l.SetField(l.Get(lua.RegistryIndex), "_LOADERS", newLoaders)
}
