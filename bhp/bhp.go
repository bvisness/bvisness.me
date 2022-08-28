package bhp

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
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

func Run(srcDir, includeDir string, data interface{}) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("\nERROR: %v\n\n", r)
				debug.PrintStack()
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		t := must1(builtin.Clone())
		must1(t.ParseFS(os.DirFS(includeDir), "*"))

		var filename string
		if r.URL.Path == "" || r.URL.Path == "/" {
			filename = ""
		} else {
			filename = must1(filepath.Rel("/", r.URL.Path))
		}
		if filepath.Ext(filename) == "" {
			filename += "/index.html"
		}

		srcFilename := filepath.Join(srcDir, filename)
		if _, err := os.Stat(srcFilename); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				w.WriteHeader(http.StatusNotFound)
				t.ExecuteTemplate(w, "404.html", nil)
				return
			} else {
				panic(err)
			}
		}
		must1(t.Parse(string(must1(os.ReadFile(srcFilename)))))

		if code, location := getRedirect(t, data); location != "" {
			w.Header().Add("Location", location)
			w.WriteHeader(code)
			return
		}

		must0(t.Execute(w, data))
	})

	addr := ":8484"
	fmt.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getRedirect(t *template.Template, data interface{}) (int, string) {
	if redirect := t.Lookup("redirect"); redirect != nil {
		code := http.StatusSeeOther
		if templateCode := getStatus(t, data); templateCode != 0 {
			code = templateCode
		}

		var locationBytes bytes.Buffer
		must0(t.ExecuteTemplate(&locationBytes, "redirect", data))
		location := strings.TrimSpace(locationBytes.String())
		return code, location
	}

	return 0, ""
}

func getStatus(t *template.Template, data interface{}) int {
	if status := t.Lookup("status"); status != nil {
		var statusBytes bytes.Buffer
		must0(t.ExecuteTemplate(&statusBytes, "status", data))
		statusStr := strings.TrimSpace(statusBytes.String())

		code, err := strconv.Atoi(statusStr)
		if err != nil {
			fmt.Printf("WARNING: '%s' is not a good status code!\n", statusStr)
		}
		return code
	}

	return 0
}

// Takes an (error) return and panics if there is an error.
// Helps avoid `if err != nil` in scripts. Use sparingly in real code.
func must0(err error) {
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
