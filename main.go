package main

import (
	"fmt"
	"html/template"
	"time"

	"github.com/bvisness/bvisness.me/bhp"
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
		return template.HTML(fmt.Sprintf("<pre>%s</pre>", md))
	},
}

func main() {
	bhp.Run("site", "include", funcs, Bvisness{
		Articles: articles,
	})
}
