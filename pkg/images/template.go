package images

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bvisness/bvisness.me/bhp"
)

func Middleware[T any](b bhp.Instance[T], r bhp.Request[T], w http.ResponseWriter, m bhp.MiddlewareData[T]) bool {
	isImage := false
	for _, format := range AllFormats {
		if m.ContentType == format {
			isImage = true
		}
	}

	if !isImage {
		return false
	}

	origScaleStr := r.R.URL.Query().Get("orig")
	scaleStr := r.R.URL.Query().Get("scale")
	if origScaleStr == "" || scaleStr == "" {
		return false
	}
	origScale, err1 := strconv.Atoi(origScaleStr)
	scale, err2 := strconv.Atoi(scaleStr)
	if err1 != nil || err2 != nil {
		return false
	}

	// TODO: alternate encodings

	key := fmt.Sprintf("%s:orig(%d)", m.FilePath, origScale)
	processed, err := imageCache.GetOrStore(key, func() (ProcessedImage, error) {
		return ProcessImage(m.FilePath, origScale, ImageOptions{})
	})
	if err != nil {
		panic(err)
	}

	for _, variant := range processed.Variants {
		if variant.Scale == scale {
			w.Header().Set("Content-Type", variant.ContentType)
			w.Write(variant.Data)
			return true
		}
	}

	return false
}

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
