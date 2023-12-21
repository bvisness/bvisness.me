package images

import (
	_ "embed"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/utils"
	lua "github.com/yuin/gopher-lua"
)

//go:embed images.luax
var impl string

func vec2p(l *lua.LState, p image.Point) lua.LValue {
	constructor := l.GetGlobal("Vec2").(*lua.LTable).RawGetString("new").(*lua.LFunction)
	l.CallByParam(lua.P{
		Fn:   constructor,
		NRet: 1,
	}, l.GetGlobal("Vec2"), lua.LNumber(p.X), lua.LNumber(p.Y))
	v := l.ToTable(-1)
	l.Pop(1)
	return v
}

func LoadLib(l *lua.LState, b *bhp.Instance, r *http.Request) int {
	utils.Must(l.DoString("require(\"vec\")"))
	mod := l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"variants": func(l *lua.LState) int {
			abspath := l.ToString(1)
			scale := int(l.ToNumber(2))

			if scale < 1 {
				scale = 1
			}

			filepath, _, _, err := b.ResolveFile(abspath)
			if err != nil {
				return bhp.Raise(l, err)
			}

			processed, err := imageCache.GetOrStore(cacheKey(filepath, scale), func() (ProcessedImage, error) {
				return ProcessImage(filepath, scale, ImageOptions{})
			})
			if err != nil {
				return bhp.Raise(l, err)
			}

			variants := l.NewTable()
			for _, variant := range processed.Variants {
				url := fmt.Sprintf(
					"%s?%s",
					bhp.AbsURL(r, abspath),
					url.Values{
						"orig":  {strconv.Itoa(processed.Source.Scale)},
						"scale": {strconv.Itoa(variant.Scale)},
						"fmt":   {variant.ContentType},
					}.Encode(),
				)

				v := l.NewTable()
				v.RawSetString("url", lua.LString(url))
				v.RawSetString("contentType", lua.LString(variant.ContentType))
				v.RawSetString("scale", lua.LNumber(variant.Scale))
				v.RawSetString("actualSize", vec2p(l, variant.ActualSize))
				v.RawSetString("intendedSize", vec2p(l, variant.IntendedSize))
				variants.Append(v)
			}

			l.Push(variants)
			return 1
		},
	})
	l.SetGlobal("images", mod)

	loader := utils.Must1(bhp.LoadLuaX(l, "images.luax", impl))
	l.Push(loader)
	l.Call(0, lua.MultRet)

	return 0
}
