package code

import (
	"bytes"
	_ "embed"
	"net/http"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/bvisness/bvisness.me/bhp2"
	"github.com/bvisness/bvisness.me/utils"
	lua "github.com/yuin/gopher-lua"
)

//go:embed code.luax
var impl string

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

func init() {
	htmlFormatter = html.New(html.WithClasses(true), html.TabWidth(4))
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
}

func LoadLib(l *lua.LState, b *bhp2.Instance, r *http.Request) int {
	mod := l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"highlight": func(l *lua.LState) int {
			lang := l.ToString(1)
			src := l.ToString(2)

			lex := lexers.Get(lang)
			if lex == nil {
				lex = lexers.Fallback
			}
			lex = chroma.Coalesce(lex)

			it, err := lex.Tokenise(nil, src)
			if err != nil {
				return bhp2.RaiseMsg(l, err, "failed to highlight code")
			}

			var b bytes.Buffer
			if err := htmlFormatter.Format(&b, highlightStyle, it); err != nil {
				return bhp2.RaiseMsg(l, err, "failed to highlight code")
			}

			l.Push(lua.LString(b.String()))
			return 1
		},
	})
	l.SetGlobal("code", mod)

	loader := utils.Must1(bhp2.LoadLuaX(l, "code.luax", impl))
	l.Push(loader)
	l.Call(0, lua.MultRet)

	return 0
}
