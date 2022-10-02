package bhp

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig"
)

//go:embed builtin/*
var builtinFS embed.FS

var builtin *template.Template

func init() {
	builtin = template.New("bhp")
	builtin.Funcs(sprig.FuncMap())
	builtin.ParseFS(builtinFS, "builtin/*")
}

func readFileString(fs embed.FS, name string) string {
	return string(must1(fs.ReadFile(name)))
}

type AddFuncsFunc[UserData any] func(Request[UserData]) template.FuncMap

type Instance[UserData any] struct {
	SrcDir, IncludeDir string
	Funcs              AddFuncsFunc[UserData]
	UserData           UserData
}

type Request[UserData any] struct {
	T    *template.Template
	R    *http.Request
	User UserData
}

func New[UserData any](
	srcDir, includeDir string,
	funcs AddFuncsFunc[UserData],
	userData UserData,
) Instance[UserData] {
	return Instance[UserData]{
		SrcDir:     srcDir,
		IncludeDir: includeDir,
		Funcs:      funcs,
		UserData:   userData,
	}
}

func Run[UserData any](
	srcDir, includeDir string,
	funcs AddFuncsFunc[UserData],
	userData UserData,
) {
	New(srcDir, includeDir, funcs, userData).Run()
}

func (b Instance[UserData]) Run() {
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
func (b Instance[UserData]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	t := must1(builtin.Clone())
	b.addBuiltinFuncs(t, r)

	bhpRequest := Request[UserData]{
		T:    t,
		R:    r,
		User: b.UserData,
	}
	t.Funcs(b.Funcs(bhpRequest))

	filepath.Walk(b.IncludeDir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			name := must1(filepath.Rel(b.IncludeDir, path))
			name = strings.ReplaceAll(name, "\\", "/")
			contents := must1(io.ReadAll(must1(os.Open(path))))
			must1(t.New(name).Parse(string(contents)))
		}
		return nil
	})
	var filename string
	if r.URL.Path == "" || r.URL.Path == "/" {
		filename = ""
	} else {
		filename = must1(filepath.Rel("/", r.URL.Path))
	}

	srcFilename, fileInfo, err := b.ResolveFileOrDir(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
			t.ExecuteTemplate(w, "404.html", nil)
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
				t.ExecuteTemplate(w, "404.html", nil)
				return
			} else {
				panic(err)
			}
		}
	}

	fileBytes := must1(os.ReadFile(srcFilename))
	contentType := detectContentType(fileInfo, fileBytes)
	switch stripContentType(contentType) {
	case "text/html", "text/css", "text/xml":
		must1(t.Parse(string(fileBytes)))

		if code, location := getRedirect(t, b.UserData); location != "" {
			w.Header().Add("Location", location)
			w.WriteHeader(code)
			return
		}

		w.Header().Add("Content-Type", contentType)

		if contentType == "text/xml" {
			w.Write([]byte("<?xml version=\"1.0\" standalone=\"yes\" ?>\n"))
		}
		must(t.Execute(w, b.UserData))
	default:
		must1(w.Write(fileBytes))
	}
}

func getRedirect(t *template.Template, data any) (int, string) {
	if redirect := t.Lookup("redirect"); redirect != nil {
		code := http.StatusSeeOther
		if templateCode := getStatus(t, data); templateCode != 0 {
			code = templateCode
		}

		var locationBytes bytes.Buffer
		must(t.ExecuteTemplate(&locationBytes, "redirect", data))
		location := strings.TrimSpace(locationBytes.String())
		return code, location
	}

	return 0, ""
}

func getStatus(t *template.Template, data any) int {
	if status := t.Lookup("status"); status != nil {
		var statusBytes bytes.Buffer
		must(t.ExecuteTemplate(&statusBytes, "status", data))
		statusStr := strings.TrimSpace(statusBytes.String())

		code, err := strconv.Atoi(statusStr)
		if err != nil {
			fmt.Printf("WARNING: '%s' is not a good status code!\n", statusStr)
		}
		return code
	}

	return 0
}

func (b Instance[UserData]) addBuiltinFuncs(t *template.Template, r *http.Request) {
	t.Funcs(template.FuncMap{
		"eval": func(name string, arg any) string {
			return Eval(t, name, arg)
		},
		"echo": func(str string) string {
			return str
		},
		"request": func() *http.Request {
			return r
		},
		"relpath": func(path string) string {
			return RelPath(r, path)
		},
		"absurl": func(path string) string {
			return AbsURL(r, path)
		},
		"relurl": func(path string) string {
			return RelURL(r, path)
		},
		"query": func(name string) string {
			return r.URL.Query().Get(name)
		},
		"safeHTML": func(html string) template.HTML {
			return template.HTML(html)
		},
		"safeJS": func(js string) template.JS {
			return template.JS(js)
		},
		"addstr": func(strs ...string) string {
			// how in the flying fuck does sprig not have this
			return strings.Join(strs, "")
		},
		"path2file": func(abspath string) (string, error) {
			srcFilename, _, err := b.ResolveFile(abspath)
			if err != nil {
				return "", err
			}
			return srcFilename, nil
		},
	})
}

func detectContentType(fileInfo os.FileInfo, fileBytes []byte) string {
	// In the presence of some extensions, we trust you will not name your files stupidly
	switch filepath.Ext(fileInfo.Name()) {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".xml":
		return "text/xml"
	default:
		return http.DetectContentType(fileBytes)
	}
}

func stripContentType(contentType string) string {
	return strings.SplitN(contentType, ";", 2)[0]
}

// Takes an (error) return and panics if there is an error.
// Helps avoid `if err != nil` in scripts. Use sparingly in real code.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Takes a (something, error) return and panics if there is an error.
// Helps avoid `if err != nil` in scripts. Use sparingly in real code.
func must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
