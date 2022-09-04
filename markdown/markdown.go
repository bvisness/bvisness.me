package markdown

import (
	"regexp"
	"strings"
	"unicode"

	gomarkdown "github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
)

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
