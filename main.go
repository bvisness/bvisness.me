package main

import (
	"fmt"
	"html/template"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/markdown"
)

type Bvisness struct {
	Articles []Article
}

type HeaderData struct {
	Title string
}

type Article struct {
	HeaderData
	Date    time.Time
	Slug    string
	Excerpt string
	Url     string
}

var articles = []Article{
	{
		HeaderData: HeaderData{
			Title: "Untangling a bizarre WASM crash in Chrome",
		},
		Slug: "chrome-wasm-crash",
		Date: time.Date(2021, 7, 9, 0, 0, 0, 0, time.UTC),
	},
}

var hash string = fmt.Sprintf("%d", time.Now().Unix())

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("failed to read build info")
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			hash = setting.Value
		}
	}
}

var funcs = template.FuncMap{
	"article": func(slug string) Article {
		for _, article := range articles {
			if article.Slug == slug {
				return article
			}
		}
		panic(fmt.Errorf("No article found with slug %s", slug))
	},
	"markdown": func(md string) template.HTML {
		md = markdown.Unindent(md)
		return template.HTML(markdown.ToHTML(md))
	},
	"bust": func(resourceUrl string) string {
		resUrlParsed, err := url.Parse(resourceUrl)
		if err != nil {
			panic(err)
		}
		q := resUrlParsed.Query()
		q.Set("v", hash)
		resUrlParsed.RawQuery = q.Encode()
		return resUrlParsed.String()
	},
}

func main() {
	bhp.Run("site", "include", funcs, Bvisness{
		Articles: articles,
	})
}
