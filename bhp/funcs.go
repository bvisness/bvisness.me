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

func (b Instance[any]) ResolveFile(abspath string) (srcFilename string, fileInfo fs.FileInfo, err error) {
findfile:
	srcFilename = filepath.Join(b.SrcDir, abspath)
	fileInfo, err = os.Stat(srcFilename)
	if err != nil {
		return "", nil, fmt.Errorf("could not resolve file: %w", err)
	}
	if fileInfo.IsDir() {
		abspath += "/index.html"
		goto findfile // would be hilarious if you made a directory called index.html
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