package images

import (
	"fmt"
	"html/template"
	"image"
	"os"

	_ "image/png"
)

var TemplateFuncs = template.FuncMap{
	"imgsize": func(path string) (image.Point, error) {
		return ImageSize(path)
	},
}

func ImageSize(path string) (image.Point, error) {
	reader, err := os.Open(path)
	if err != nil {
		return image.Point{}, fmt.Errorf("failed to get image size: %w", err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return image.Point{}, fmt.Errorf("failed to decode image: %w", err)
	}
	return img.Bounds().Size(), nil
}
