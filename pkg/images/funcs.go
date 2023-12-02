package images

import (
	_ "embed"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bvisness/bvisness.me/bhp"
	"github.com/bvisness/bvisness.me/bhp2"
	"github.com/bvisness/bvisness.me/utils"
	lua "github.com/yuin/gopher-lua"
)

//go:embed images.luax
var impl string

// Takes the source image, rescales and re-encodes it into several useful
// resolutions and formats, and returns a value suitable for a `srcset`
// attribute.
//
// Example:
// input: {{ srcset "/desmos/images/foo.png" 2 }}
// output: images/foo.png?orig=2&scale=2 2x, images/foo.png?orig=2&scale=1 1x
func SrcSet(r *http.Request, abspath string, scale int) string {
	if scale < 1 {
		scale = 1
	}

	var candidates []string
	for candidateScale := scale; candidateScale >= 1; candidateScale-- {
		candidates = append(candidates, fmt.Sprintf(
			"%s?%s %dx",
			bhp2.AbsURL(r, abspath),
			url.Values{
				"orig":  {strconv.Itoa(scale)},
				"scale": {strconv.Itoa(candidateScale)},
			}.Encode(),
			candidateScale,
		))
	}

	return strings.Join(candidates, ", ")
}

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

func LoadLib(l *lua.LState, b *bhp2.Instance, r *http.Request) int {
	utils.Must(l.DoString("require(\"vec\")"))
	mod := l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"variants": func(l *lua.LState) int {
			abspath := l.ToString(1)
			scale := int(l.ToNumber(2))

			if scale < 1 {
				scale = 1
			}

			filepath, _, err := b.ResolveFile(abspath)
			if err != nil {
				return bhp2.Raise(l, err)
			}

			processed, err := imageCache.GetOrStore(cacheKey(filepath, scale), func() (ProcessedImage, error) {
				return ProcessImage(filepath, scale, ImageOptions{})
			})
			if err != nil {
				return bhp2.Raise(l, err)
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

	loader := utils.Must1(bhp2.LoadLuaX(l, "images.luax", impl))
	l.Push(loader)
	l.Call(0, lua.MultRet)

	return 0
}
