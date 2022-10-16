package images

import (
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"github.com/bvisness/bvisness.me/bhp"
)

func TemplateFuncs[T any](b bhp.Instance[T], r bhp.Request[T]) template.FuncMap {
	return template.FuncMap{
		// "imgsize": func(path string) (image.Point, error) {
		// 	return ImageSize(path)
		// },

		// Takes the source image, rescales and re-encodes it into several useful
		// resolutions and formats, and returns a value suitable for a `srcset`
		// attribute.
		//
		// Example:
		// input: {{ srcset "/desmos/images/foo.png" 2 }}
		// output: images/foo.png?orig=2&scale=2 2x, images/foo.png?orig=2&scale=1 1x
		"srcset": func(abspath string, scale int) (string, error) {
			var candidates []string
			for candidateScale := scale; candidateScale >= 1; candidateScale-- {
				candidates = append(candidates, fmt.Sprintf(
					"%s?%s %dx",
					bhp.AbsURL(r.R, abspath),
					url.Values{
						"orig":  {strconv.Itoa(scale)},
						"scale": {strconv.Itoa(candidateScale)},
					}.Encode(),
					candidateScale,
				))
			}

			return strings.Join(candidates, ", "), nil
		},
	}
}
