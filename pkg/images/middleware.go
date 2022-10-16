package images

import (
	"net/http"
	"strconv"

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
	format := r.R.URL.Query().Get("fmt")

	processed, err := imageCache.GetOrStore(cacheKey(m.FilePath, origScale), func() (ProcessedImage, error) {
		return ProcessImage(m.FilePath, origScale, ImageOptions{})
	})
	if err != nil {
		panic(err)
	}

	for _, variant := range processed.Variants {
		scaleOk := variant.Scale == scale
		var formatOk bool
		if format == "" {
			// Accept only the original content type
			formatOk = variant.ContentType == processed.Source.ContentType
		} else {
			formatOk = variant.ContentType == format
		}
		if scaleOk && formatOk {
			w.Header().Set("Content-Type", variant.ContentType)
			w.Write(variant.Data)
			return true
		}
	}

	return false
}
