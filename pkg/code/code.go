package code

import (
	"bytes"
	_ "embed"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/pkg/lru"
	"github.com/bvisness/bvisness.me/utils"
	lua "github.com/yuin/gopher-lua"
)

//go:embed code.luax
var impl string

var highlightCache = lru.New[string](1000)

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

func init() {
	htmlFormatter = html.New(html.WithClasses(true), html.TabWidth(4))
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
}

func LoadLib(l *lua.LState) int {
	mod := l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"highlight": func(l *lua.LState) int {
			var lang, src string
			if s, ok := l.Get(1).(lua.LString); ok {
				lang = string(s)
			}
			src = l.CheckString(2)

			key := lang + "/" + src

			highlighted, err := highlightCache.GetOrStore(key, func() (string, error) {
				lex := lexers.Get(lang)
				if lex == nil {
					lex = lexers.Fallback
				}
				lex = chroma.Coalesce(lex)

				it, err := lex.Tokenise(nil, src)
				if err != nil {
					return "", err
				}

				var b bytes.Buffer
				if err := htmlFormatter.Format(&b, highlightStyle, it); err != nil {
					return "", err
				}

				return b.String(), nil
			})
			if err != nil {
				return bhp.RaiseMsg(l, err, "failed to highlight code")
			}

			l.Push(lua.LString(highlighted))
			return 1
		},
	})
	l.SetGlobal("code", mod)

	loader := utils.Must1(bhp.LoadLuaX(l, "code.luax", impl))
	l.Push(loader)
	l.Call(0, lua.MultRet)

	return 0
}
