package markdown

import (
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
)

// This implementation of syntax highlighting for gomarkdown was taken from
// https://github.com/kjk/blog/blob/e5ce8ed9aa14966391997d5cb3c382408cbcf5a1/md_to_html.go

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

func init() {
	htmlFormatter = html.New(html.WithClasses(true), html.TabWidth(4))
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
}

// based on https://github.com/alecthomas/chroma/blob/master/quick/quick.go
func htmlHighlight(w io.Writer, source, lang, defaultLang string) error {
	if lang == "" {
		lang = defaultLang
	}
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlFormatter.Format(w, highlightStyle, it)
}

func makeRenderHookCodeBlock(defaultLang string) mdhtml.RenderNodeFunc {
	return func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		codeBlock, ok := node.(*ast.CodeBlock)
		if !ok {
			return ast.GoToNext, false
		}
		lang := string(codeBlock.Info)
		if false {
			fmt.Printf("lang: '%s', code: %s\n", lang, string(codeBlock.Literal))
			io.WriteString(w, "\n<pre class=\"chroma\"><code>")
			mdhtml.EscapeHTML(w, codeBlock.Literal)
			io.WriteString(w, "</code></pre>\n")
		} else {
			htmlHighlight(w, string(codeBlock.Literal), lang, defaultLang)
		}
		return ast.GoToNext, true
	}
}
