package bhp2

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// Takes a path relative to the current request path and produces an absolute path.
func RelPath(r *http.Request, path string) string {
	fullPath, err := url.JoinPath(r.URL.Path, path)
	if err != nil {
		panic(err)
	}
	return fullPath
}

// Takes a full path from the root of the site and produces an absolute URL.
func AbsURL(r *http.Request, path string) string {
	newurl := *r.URL
	newurl.Path = path
	return newurl.String()
}

// Takes a path relative to the current request path and produces an absolute URL.
func RelURL(r *http.Request, path string) string {
	newurl := *r.URL
	newurl.Path = RelPath(r, path)
	return newurl.String()
}

func Bust(resourceUrl string) string {
	resUrlParsed, err := url.Parse(resourceUrl)
	if err != nil {
		panic(err)
	}
	q := resUrlParsed.Query()
	q.Set("v", Hash)
	resUrlParsed.RawQuery = q.Encode()
	return resUrlParsed.String()
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
	abspath += "/index.luax" // who knows, maybe someday we could support other kinds of indexes
	srcFilename, fileInfo, err = b.ResolveFileOrDir(abspath)
	if err != nil {
		return
	}
	if fileInfo.IsDir() {
		return "", nil, fmt.Errorf("expected valid index file at %s, but got a directory", abspath)
	}
	return
}

func (b Instance) ResolveFile(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
	srcFilename, fileInfo, err = b.ResolveFileOrDir(abspath)
	if err != nil {
		return
	}
	if fileInfo.IsDir() {
		return b.ResolveDirectoryIndex(abspath)
	}
	return
}

func ChainMiddleware(middlewares ...Middleware) Middleware {
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

func LoadURLLib(l *lua.LState, r *http.Request) int {
	l.SetGlobal("relpath", l.NewClosure(WrapS_S(func(path string) string {
		return RelPath(r, path)
	})))
	l.SetGlobal("absurl", l.NewClosure(WrapS_S(func(path string) string {
		return AbsURL(r, path)
	})))
	l.SetGlobal("relurl", l.NewClosure(WrapS_S(func(path string) string {
		return RelURL(r, path)
	})))
	l.SetGlobal("bust", l.NewClosure(WrapS_S(Bust)))
	l.SetGlobal("permalink", l.NewClosure(WrapS_S(func(s string) string {
		return RelURL(r, "/")
	})))

	return 0
}
