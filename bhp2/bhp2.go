package bhp2

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/bvisness/bvisness.me/utils"
	lua "github.com/yuin/gopher-lua"
)

//go:embed builtin/*
var builtinFS embed.FS

type Middleware func(b Instance, r *http.Request, w http.ResponseWriter, m MiddlewareData) bool
type MiddlewareData struct {
	FilePath    string
	ContentType string
}

type Instance struct {
	SrcDir      string
	FSSearchers []FSSearcher
	StaticPaths []string
	Middleware  Middleware
	Libs        map[string]GoLibLoader
}

type GoLibLoader func(l *lua.LState, b *Instance, r *http.Request) int

type Options struct {
	StaticPaths []string
	Middleware  Middleware
}

func (b Instance) Run() {
	go func() {
		// Start up private API for pprof
		addr := ":9494"
		fmt.Println("Private stuff listening on", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	addr := ":8484"
	fmt.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, b))
}

// Implements http.Handler. No mux necessary!
func (b Instance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\nERROR: %v\n\n", r)
			debug.PrintStack()
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	proxyHeaders(r)
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}

	srcFilename, fileInfo, redirectPath, err := b.ResolveFile(r.URL.Path)
	if errors.Is(err, fs.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		// TODO: 404
		return
	} else if err != nil {
		panic(err)
	}

	if redirectPath != "" {
		r.URL.Path = redirectPath
		http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
		return
	}

	file := utils.Must1(os.Open(srcFilename))
	var opening [512]byte
	file.Read(opening[:])
	file.Seek(0, io.SeekStart)
	contentType := detectContentType(fileInfo, opening[:])

	if b.Middleware != nil {
		didHandle := b.Middleware(b, r, w, MiddlewareData{
			FilePath:    srcFilename,
			ContentType: contentType,
		})
		if didHandle {
			return
		}
	}

	switch contentType {
	case "text/luax":
		l := lua.NewState()
		defer l.Close()
		b.initSearchers(l, r)

		// TODO: save bytecode of BHP for faster startup
		l.PreloadModule("bhp", LoadBHP2Lib)
		l.PreloadModule("url", func(l *lua.LState) int {
			return LoadURLLib(l, r)
		})
		utils.Must(l.DoString("require(\"bhp\")"))
		utils.Must(l.DoString("require(\"url\")"))

		fileBytes := utils.Must1(io.ReadAll(file))
		mainChunk, err := LoadLuaX(l, srcFilename, string(fileBytes))
		if err != nil {
			// TODO: Error codes and stuff for everything
			// TODO: Report syntax errors in the browser
			l.RaiseError("error loading main chunk %s: %v", srcFilename, err)
			return
		}
		l.Push(mainChunk)
		if err := l.PCall(0, lua.MultRet, nil); err != nil {
			// TODO: Error handling
			w.Header().Add("Content-Type", "text/plain")
			w.Write([]byte(err.Error()))
			panic(err)
		}
		toRender := l.CheckAny(-1)

		if t, ok := toRender.(*lua.LTable); ok {
			if action, ok := t.RawGetString("action").(lua.LString); ok && action == "redirect" {
				code := int(t.RawGetString("code").(lua.LNumber))
				location := string(t.RawGetString("url").(lua.LString))

				w.Header().Add("Location", location)
				w.WriteHeader(code)
				return
			}
		}

		if toRender == lua.LNil {
			fmt.Printf("WARNING: Page returned nil; no content will be rendered.\n")
		}

		err = l.CallByParam(lua.P{
			Fn:      l.GetGlobal("bhp").(*lua.LTable).RawGetString("render").(*lua.LFunction),
			NRet:    1,
			Protect: true,
		}, toRender)
		if err != nil {
			// TODO: Error handling
			w.Header().Add("Content-Type", "text/plain")
			w.Write([]byte(err.Error()))
			panic(err)
		}
		rendered := l.CheckString(-1)

		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(rendered))
	case "text/xml":
		// Stupid hacks 😑
		w.Write([]byte("<?xml version=\"1.0\" standalone=\"yes\" ?>\n"))
		// TODO: fix RSS
		// must(t.Execute(w, b.UserData))
	default:
		w.Header().Add("Content-Type", contentType)
		http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
	}
}

func detectContentType(fileInfo os.FileInfo, fileBytes []byte) string {
	// In the presence of some extensions, we trust you will not name your files stupidly
	switch filepath.Ext(fileInfo.Name()) {
	case ".luax":
		return "text/luax"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".xml":
		return "text/xml"
	case ".svg":
		return "image/svg+xml"
	case ".js":
		return "text/javascript"
	default:
		return http.DetectContentType(fileBytes)
	}
}

func stripContentType(contentType string) string {
	return strings.SplitN(contentType, ";", 2)[0]
}

func (b *Instance) pathIsStatic(path string) bool {
	for _, staticPath := range b.StaticPaths {
		if strings.HasPrefix(path, staticPath) {
			return true
		}
	}
	return false
}
