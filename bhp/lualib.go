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
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// Files with these extensions can be loaded raw from FSSearchers.
var staticFileExts = []string{".svg"}

//go:embed lua/*
var builtins embed.FS

var builtinSearcher = &EmbedFSSearcher{
	FS:     builtins,
	Prefix: "lua/",
}

func GetInstance(l *lua.LState) *Instance {
	return l.GetGlobal("bhp").(*lua.LTable).
		RawGetString("_instance").(*lua.LUserData).
		Value.(*Instance)
}

func setInstance(l *lua.LState, b *Instance) {
	ud := l.NewUserData()
	ud.Value = b
	l.GetGlobal("bhp").(*lua.LTable).RawSetString("_instance", ud)
}

func GetRequest(l *lua.LState) *http.Request {
	return l.GetGlobal("bhp").(*lua.LTable).
		RawGetString("_request").(*lua.LUserData).
		Value.(*http.Request)
}

func setRequest(l *lua.LState, r *http.Request) {
	ud := l.NewUserData()
	ud.Value = r
	l.GetGlobal("bhp").(*lua.LTable).RawSetString("_request", ud)
}

func CompileLua(source io.Reader, filePath string) (*lua.FunctionProto, error) {
	chunk, err := parse.Parse(source, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

func safeFilename(filename string) string {
	return strings.ReplaceAll(filename, "\\", "/")
}

func CompileLuaX(source, filename string) (*lua.FunctionProto, error) {
	filename = safeFilename(filename)

	transpiled, err := Transpile(source, filename)
	if err != nil {
		return nil, fmt.Errorf("error transpiling file: %w", err)
	}

	return CompileLua(strings.NewReader(transpiled), filename)
}

func LoadLuaX(l *lua.LState, filename, source string) (*lua.LFunction, error) {
	filename = safeFilename(filename)

	proto, err := CompileLuaX(source, filename)
	if err != nil {
		return nil, err
	}

	sources := l.GetGlobal("bhp").(*lua.LTable).RawGetString("_sources").(*lua.LTable)
	sources.RawSetString(filename, lua.LString(source))

	return l.NewFunctionFromProto(proto), nil
}

//
// Searchers (package.loaders)
//

type Searcher interface {
	// Should push a Lua value corresponding to package.loaders:
	// https://www.lua.org/manual/5.1/manual.html#pdf-package.loaders
	//
	// That is, it should return a loader function, a string explaining why it
	// did not find anything, or nil.
	Search(l *lua.LState) int
}

type FSSearcher struct {
	FS     fs.FS
	Prefix string // Path prefix to prepend to `require` string before lookup, e.g. `lua/`
}

func (s FSSearcher) Search(l *lua.LState) int {
	if s.searchPlainLua(l) || s.searchLuaX(l) {
		return 1
	}

	if searchStaticFile(l, s.FS) {
		return 1
	}

	l.Push(lua.LString(fmt.Sprintf("no file found for '%s'", s.Prefix+l.CheckString(1))))
	return 1
}

func (s FSSearcher) searchPlainLua(l *lua.LState) bool {
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

func (s FSSearcher) searchLuaX(l *lua.LState) bool {
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

type EmbedFSSearcher struct {
	FS     embed.FS
	Prefix string // Path prefix to prepend to `require` string before lookup, e.g. `lua/`

	protos map[string]*lua.FunctionProto
	lock   sync.RWMutex
}

func (s *EmbedFSSearcher) Search(l *lua.LState) int {
	name := l.CheckString(1)
	if proto, ok := s.getProto(name); ok {
		l.Push(l.NewFunctionFromProto(proto))
		return 1
	}

	if s.searchPlainLua(l, name) || s.searchLuaX(l, name) {
		return 1
	}

	if searchStaticFile(l, s.FS) {
		return 1
	}

	l.Push(lua.LString(fmt.Sprintf("no embedded file found for '%s'", s.Prefix+name)))
	return 1
}

func (s *EmbedFSSearcher) ensureProtos() {
	if s.protos == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		s.protos = make(map[string]*lua.FunctionProto)
	}
}

func (s *EmbedFSSearcher) getProto(name string) (*lua.FunctionProto, bool) {
	s.ensureProtos()
	s.lock.RLock()
	defer s.lock.RUnlock()
	proto, ok := s.protos[name]
	return proto, ok
}

func (s *EmbedFSSearcher) saveProto(name string, proto *lua.FunctionProto) {
	s.ensureProtos()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.protos[name] = proto
}

func (s *EmbedFSSearcher) searchPlainLua(l *lua.LState, name string) bool {
	filename := s.Prefix + name + ".lua"
	f, err := s.FS.Open(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	} else if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error opening file: %v", err)))
		return true
	}

	proto, err := CompileLua(f, filename)
	if err != nil {
		l.Push(lua.LString(fmt.Sprintf("error compiling file: %v", err)))
		return true
	}
	s.saveProto(name, proto)

	l.Push(l.NewFunctionFromProto(proto))
	return true
}

func (s *EmbedFSSearcher) searchLuaX(l *lua.LState, name string) bool {
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

	proto, err := CompileLuaX(string(source), filename)
	if err != nil {
		l.Push(lua.LString(err.Error()))
		return true
	}
	s.saveProto(name, proto)

	l.Push(l.NewFunctionFromProto(proto))
	return true
}

func searchStaticFile(l *lua.LState, files fs.FS) bool {
	name := l.CheckString(1)
	for _, ext := range staticFileExts {
		if strings.HasSuffix(name, ext) {
			f, err := files.Open(name)
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

type GoSearcher map[string]lua.LGFunction

func (s GoSearcher) Search(l *lua.LState) int {
	name := l.CheckString(1)
	if lib, ok := s[name]; ok {
		l.Push(l.NewFunction(lib))
		return 1
	} else {
		l.Push(lua.LNil)
		return 1
	}
}

func (b *Instance) initSearchers(l *lua.LState) {
	p := l.GetGlobal("package")
	oldSearchers := l.GetField(p, "loaders").(*lua.LTable)
	preloadSearcher := oldSearchers.RawGetInt(1).(*lua.LFunction)

	newSearchers := l.NewTable()
	i := 0

	// Add the original preload searcher
	i++
	l.RawSetInt(newSearchers, i, preloadSearcher)

	// Add builtin searcher
	i++
	l.RawSetInt(newSearchers, i, l.NewFunction(builtinSearcher.Search))

	// Add built-in go libs
	i++
	builtinGoSearcher := GoSearcher{
		"url": LoadURLLib,
	}
	l.RawSetInt(newSearchers, i, l.NewFunction(builtinGoSearcher.Search))

	// Add user searchers
	for _, s := range b.Searchers {
		s := s // >:(

		i++
		l.RawSetInt(newSearchers, i, l.NewFunction(s.Search))
	}

	l.SetField(p, "loaders", newSearchers)
	l.SetField(l.Get(lua.RegistryIndex), "_LOADERS", newSearchers)
}
