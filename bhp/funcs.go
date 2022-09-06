package bhp

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"
)

func Eval(t *template.Template, templateName string, data any) string {
	var buf bytes.Buffer
	must(t.ExecuteTemplate(&buf, templateName, data))
	return buf.String()
}

func AbsURL(r *http.Request, path string) string {
	newurl := *r.URL
	newurl.Path = path
	return newurl.String()
}

func RelURL(r *http.Request, relurl string) string {
	res, err := url.JoinPath(r.URL.Path, relurl)
	if err != nil {
		panic(err)
	}
	newurl := *r.URL
	newurl.Path = res
	return newurl.String()
}
