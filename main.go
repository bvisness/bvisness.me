package main

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/gomarkdown/markdown"
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
		md = Unindent(md)
		return template.HTML(markdown.ToHTML([]byte(md), nil, nil))
	},
}

func main() {
	bhp.Run("site", "include", funcs, Bvisness{
		Articles: articles,
	})
}

// Un-indents a string according to the whitespace on its first line of content.
func Unindent(str string) string {
	var leadingWhitespace string
	for i, r := range str {
		if !unicode.IsSpace(r) {
			leadingWhitespace = str[:i]
			break
		}
	}
	leadingWhitespace = strings.TrimLeft(leadingWhitespace, "\n\r")
	reLeadingSpace := regexp.MustCompile(`(?m)^` + leadingWhitespace)
	return reLeadingSpace.ReplaceAllString(str, "")
}
