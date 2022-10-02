package images

import (
	"fmt"
	"html/template"
	"image"
	"os"
	"sync"
	"time"

	_ "image/png"
)

var TemplateFuncs = template.FuncMap{
	"imgsize": func(path string) (image.Point, error) {
		return ImageSize(path)
	},
}

func ImageSize(path string) (image.Point, error) {
	img, err := getImage(path)
	if err != nil {
		return image.Point{}, err
	}
	return img.Bounds().Size(), nil
}

var imageCache sync.Map // map[path]image.Image

func getImage(path string) (image.Image, error) {
	if img, ok := imageCache.Load(path); ok {
		return img.(image.Image), nil
	}

	return storeImage(path)
}

func storeImage(path string) (image.Image, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	imageCache.Store(path, img)

	return img, nil
}

// Refresh the image cache every second. (Someday this could be less dumb.)
func init() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			imageCache.Range(func(key, value any) bool {
				storeImage(key.(string))
				return true
			})
		}
	}()
}
