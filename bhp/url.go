package bhp

import (
	"net/http"
	"net/url"
)

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
