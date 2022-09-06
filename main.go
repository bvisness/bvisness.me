package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/markdown"
)

type Bvisness struct {
	Articles []Article
}

type BaseData struct {
	Title          string
	Description    string
	OpenGraphImage string // Relative URL within site folder
}

type CommonData struct {
	Banner string
}

type Article struct {
	BaseData
	CommonData
	Date time.Time
	Slug string
	Url  string
}

var articles = []Article{
	{
		BaseData: BaseData{
			Title:          "Untangling a bizarre WASM crash in Chrome",
			Description:    "How we solved a strange issue involving the guts of Chrome and the Go compiler.",
			OpenGraphImage: "chrome-wasm-crash/ogimage.png",
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
	"permalink": func(r *http.Request) string {
		return bhp.RelURL(r, "/")
	},
}

func main() {
	bhp.Run("site", "include", funcs, Bvisness{
		Articles: articles,
	})
}
