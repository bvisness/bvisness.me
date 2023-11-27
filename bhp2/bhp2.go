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
}

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

	// OLD: walk includes
	// TODO: maybe add some directory to the require path or something?

	var filename string
	if r.URL.Path == "" || r.URL.Path == "/" {
		filename = ""
	} else {
		filename = utils.Must1(filepath.Rel("/", r.URL.Path))
	}

	srcFilename, fileInfo, err := b.ResolveFileOrDir(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
			// TODO: 404
			return
		} else {
			panic(err)
		}
	}

	// Resolve folders (either redirecting or finding an index)
	if fileInfo.IsDir() {
		// Redirect e.g. http://example.org/foo/bar to http://example.org/foo/bar/
		// Only folders are subject to this behavior, and must be valid folders.
		pathEndsInSlash := len(r.URL.Path) > 0 && r.URL.Path[len(r.URL.Path)-1] == '/'
		if !pathEndsInSlash {
			r.URL.Path += "/"
			http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
			return
		}

		srcFilename, fileInfo, err = b.ResolveDirectoryIndex(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				// TODO: 404
				return
			} else {
				panic(err)
			}
		}
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
		fileBytes := utils.Must1(io.ReadAll(file))
		transpiled := utils.Must1(Transpile(string(fileBytes), filename))

		l := lua.NewState()
		defer l.Close()
		b.initSearchers(l)

		// TODO: save bytecode of BHP for faster startup
		l.PreloadModule("bhp", LoadBHP2)
		utils.Must(l.DoString("require(\"bhp\")"))

		saveSource(l, SafeName(filename), string(fileBytes))
		mainChunk, err := l.Load(strings.NewReader(transpiled), filename)
		if err != nil {
			l.RaiseError("error loading %s: %v", filename, err)
			return
		}
		l.Push(mainChunk)
		l.Call(0, lua.MultRet)

		w.Header().Add("Content-Type", "text/html")
		w.Write([]byte(getRendered(l)))

		// TODO: write something?

		// if code, location := getRedirect(t, b.UserData); location != "" {
		// 	w.Header().Add("Location", location)
		// 	w.WriteHeader(code)
		// 	return
		// }

		// must(t.Execute(w, b.UserData))
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

func (b Instance) ResolveFileOrDir(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
	srcFilename = filepath.Join(b.SrcDir, abspath)
	fileInfo, err = os.Stat(srcFilename)
	if err != nil {
		return "", nil, fmt.Errorf("could not resolve file: %w", err)
	}
	return
}

func (b Instance) ResolveDirectoryIndex(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
	abspath += "/index.luax"
	srcFilename, fileInfo, err = b.ResolveFileOrDir(abspath)
	if err != nil {
		return
	}
	if fileInfo.IsDir() {
		return "", nil, fmt.Errorf("expected valid index file at %s, but got a directory", abspath)
	}
	return
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

func ChainMiddleware[UserData any](middlewares ...Middleware) Middleware {
	return func(b Instance, r *http.Request, w http.ResponseWriter, m MiddlewareData) bool {
		for _, middleware := range middlewares {
			didHandle := middleware(b, r, w, m)
			if didHandle {
				return true
			}
		}
		return false
	}
}
