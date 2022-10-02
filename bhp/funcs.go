package bhp

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func Eval(t *template.Template, templateName string, data any) string {
	var buf bytes.Buffer
	must(t.ExecuteTemplate(&buf, templateName, data))
	return buf.String()
}

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

func (b Instance[any]) ResolveFileOrDir(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
	srcFilename = filepath.Join(b.SrcDir, abspath)
	fileInfo, err = os.Stat(srcFilename)
	if err != nil {
		return "", nil, fmt.Errorf("could not resolve file: %w", err)
	}
	return
}

func (b Instance[any]) ResolveDirectoryIndex(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
	abspath += "/index.html" // who knows, maybe someday we could support other kinds of indexes
	srcFilename, fileInfo, err = b.ResolveFileOrDir(abspath)
	if fileInfo.IsDir() {
		return "", nil, fmt.Errorf("expected valid index file at %s, but got a directory", abspath)
	}
	return
}

func (b Instance[any]) ResolveFile(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
	srcFilename, fileInfo, err = b.ResolveFileOrDir(abspath)
	if fileInfo.IsDir() {
		return b.ResolveDirectoryIndex(abspath)
	}
	return
}

func MergeFuncMaps(funcMaps ...template.FuncMap) template.FuncMap {
	result := make(template.FuncMap)
	for _, funcs := range funcMaps {
		for name, f := range funcs {
			result[name] = f
		}
	}
	return result
}
