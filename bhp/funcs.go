package bhp

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

// Resolves the file to serve from a URL path. For a path like /foo/bar, this
// will look for the following:
//
//   - /foo/bar (the file itself)
//   - /foo/bar.luax (a .luax file with the same name)
//   - /foo/bar/index.luax (a directory index)
//
// Some paths should not be directly resolved by the browser without a
// redirect. On the other hand, when resolving a path internally (e.g. in
// middleware) you may not care. If a redirect should occur, the redirectPath
// parameter will contain the absolute URL path to redirect to.
func (b *Instance) ResolveFile(abspath string) (srcFilename string, fileInfo fs.FileInfo, redirectPath string, err error) {
	var filename string
	if abspath == "" || abspath == "/" {
		filename = ""
	} else {
		filename = strings.TrimLeft(abspath, "/")
	}

	srcFilename, fileInfo, err = b.ResolveRawFileOrDir(filename)
	if errors.Is(err, fs.ErrNotExist) {
		srcFilename, fileInfo, err = b.ResolveRawFileOrDir(filename + ".luax")
	}
	if err != nil {
		return
	}

	if fileInfo.IsDir() {
		// Redirect http://example.org/foo/bar to http://example.org/foo/bar/.
		// Only folders are subject to this behavior, and must be valid
		// folders.
		//
		// Note that this does not abort this function.
		pathEndsInSlash := len(abspath) > 0 && abspath[len(abspath)-1] == '/'
		if !pathEndsInSlash {
			redirectPath = abspath + "/"
		}

		// Resolve directory index.
		srcFilename, fileInfo, err = b.ResolveRawFileOrDir(abspath + "/index.luax")
		if err != nil {
			return
		}
		if fileInfo.IsDir() {
			err = fmt.Errorf("expected valid index file at %s, but got a directory", abspath)
			return
		}
	}
	return
}

// Get a file or directory for the given website path without resolving
// indexes, redirects, or other shenanigans. The name helps emphasize that you
// may in fact get back a directory instead of a file you can actually return.
func (b *Instance) ResolveRawFileOrDir(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
	srcFilename = filepath.Join(b.SrcDir, abspath)
	fileInfo, err = os.Stat(srcFilename)
	if err != nil {
		return "", nil, fmt.Errorf("could not resolve file for path %s: %w", abspath, err)
	}
	return
}

func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(b *Instance, r *http.Request, w http.ResponseWriter, m MiddlewareData) bool {
		for _, middleware := range middlewares {
			didHandle := middleware(b, r, w, m)
			if didHandle {
				return true
			}
		}
		return false
	}
}

func wrapUrlFunc(l *lua.LState, f func(*http.Request, string) string) *lua.LFunction {
	return l.NewFunction(func(l *lua.LState) int {
		res := f(GetRequest(l), l.CheckString(1))
		l.Push(lua.LString(res))
		return 1
	})
}

func LoadURLLib(l *lua.LState) int {
	l.SetGlobal("relpath", wrapUrlFunc(l, func(r *http.Request, path string) string {
		return RelPath(r, path)
	}))
	l.SetGlobal("absurl", wrapUrlFunc(l, func(r *http.Request, path string) string {
		return AbsURL(r, path)
	}))
	l.SetGlobal("relurl", wrapUrlFunc(l, func(r *http.Request, path string) string {
		return RelURL(r, path)
	}))
	l.SetGlobal("bust", l.NewFunction(func(l *lua.LState) int {
		res := Bust(l.CheckString(1))
		l.Push(lua.LString(res))
		return 1
	}))
	l.SetGlobal("permalink", l.NewFunction(func(l *lua.LState) int {
		res := RelURL(GetRequest(l), "/")
		l.Push(lua.LString(res))
		return 1
	}))

	return 0
}
