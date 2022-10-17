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
		// Takes the source image, rescales and re-encodes it into several useful
		// resolutions and formats, and returns a value suitable for a `srcset`
		// attribute.
		//
		// Example:
		// input: {{ srcset "/desmos/images/foo.png" 2 }}
		// output: images/foo.png?orig=2&scale=2 2x, images/foo.png?orig=2&scale=1 1x
		"srcset": func(abspath string, scale int) (string, error) {
			if scale < 1 {
				scale = 1
			}

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

		// WARNING! Chrome is big stupid!
		"imageset": func(abspath string, scale int) (template.CSS, error) {
			if scale < 1 {
				scale = 1
			}

			filepath, _, err := b.ResolveFile(abspath)
			if err != nil {
				return "", err
			}

			processed, err := imageCache.GetOrStore(cacheKey(filepath, scale), func() (ProcessedImage, error) {
				return ProcessImage(filepath, scale, ImageOptions{})
			})
			if err != nil {
				return "", err
			}

			var options []string
			for _, variant := range processed.Variants {
				url := fmt.Sprintf(
					"%s?%s",
					bhp.AbsURL(r.R, abspath),
					url.Values{
						"orig":  {strconv.Itoa(processed.Source.Scale)},
						"scale": {strconv.Itoa(variant.Scale)},
						"fmt":   {variant.ContentType},
					}.Encode(),
				)
				options = append(options, fmt.Sprintf(
					`url("%s") %dx type("%s")`,
					url, variant.Scale, variant.ContentType,
				))
			}

			return template.CSS(strings.Join(options, ", ")), nil
		},

		"picturesource": func(abspath string, scale int) (template.HTML, error) {
			if scale < 1 {
				scale = 1
			}

			filepath, _, err := b.ResolveFile(abspath)
			if err != nil {
				return "", err
			}

			processed, err := imageCache.GetOrStore(cacheKey(filepath, scale), func() (ProcessedImage, error) {
				return ProcessImage(filepath, scale, ImageOptions{})
			})
			if err != nil {
				return "", err
			}

			optionsByType := make(map[string][]string)
			for _, variant := range processed.Variants {
				url := fmt.Sprintf(
					"%s?%s",
					bhp.AbsURL(r.R, abspath),
					url.Values{
						"orig":  {strconv.Itoa(processed.Source.Scale)},
						"scale": {strconv.Itoa(variant.Scale)},
						"fmt":   {variant.ContentType},
					}.Encode(),
				)
				options := optionsByType[variant.ContentType]
				optionsByType[variant.ContentType] = append(options, fmt.Sprintf(
					`%s %dx`,
					url, variant.Scale,
				))
			}

			var b strings.Builder
			for contentType, options := range optionsByType {
				b.WriteString(fmt.Sprintf(
					`<source srcset="%s" type="%s">`,
					strings.Join(options, ", "), contentType,
				))
			}
			b.WriteString(fmt.Sprintf(`<img src="%s">`, bhp.AbsURL(r.R, abspath)))

			return template.HTML(b.String()), nil
		},
	}
}
