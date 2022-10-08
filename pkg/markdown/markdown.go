package markdown

import (
	"html/template"
	"regexp"
	"strings"
	"unicode"

	"github.com/bvisness/bvisness.me/pkg/lru"
	gomarkdown "github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
)

var markdownCache = lru.New[template.HTML](1000)

var TemplateFuncs = template.FuncMap{
	"markdown": func(md string) template.HTML {
		md = Unindent(md)

		if cachedHTML, ok := markdownCache.Get(md); ok {
			return cachedHTML
		} else {
			html := template.HTML(ToHTML(md))
			markdownCache.Store(md, html)
			return html
		}
	},
}

func ToHTML(md string) string {
	renderer := mdhtml.NewRenderer(mdhtml.RendererOptions{
		RenderNodeHook: makeRenderHookCodeBlock("html"),
	})
	return string(gomarkdown.ToHTML([]byte(md), nil, renderer))
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
