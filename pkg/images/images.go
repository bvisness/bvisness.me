package images

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"

	"image/jpeg"
	"image/png"

	"github.com/bvisness/bvisness.me/pkg/lru"
	"github.com/nfnt/resize"
)

var imageCache = lru.New[ProcessedImage](1000)

type ProcessedImage struct {
	Source   Variant
	Variants []Variant
}

type Variant struct {
	Data         []byte      // The raw image data
	ContentType  string      // The image's MIME type (e.g. "image/png")
	Scale        int         // e.g. 2 for a 2x image
	ActualSize   image.Point // The actual size of this image.
	IntendedSize image.Point // The size at which this image is intended to be rendered. (actual / scale)
}

var AllFormats = []string{"image/jpeg", "image/png", "image/webp"}

var stdimgType2MimeType = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"webp": "image/webp",
}

type encoder func(w io.Writer, img image.Image) error

var mimeType2Encoder = map[string]encoder{
	"image/png": png.Encode,
	"image/jpeg": func(w io.Writer, img image.Image) error {
		return jpeg.Encode(w, img, nil)
	},
}

func ProcessImage(filepath string, originalScale int, opts ImageOptions) (ProcessedImage, error) {
	var res ProcessedImage

	reader, err := os.Open(filepath)
	if err != nil {
		return ProcessedImage{}, fmt.Errorf("failed to open image: %w", err)
	}
	defer reader.Close()

	imgData, err := io.ReadAll(reader)
	if err != nil {
		return ProcessedImage{}, fmt.Errorf("failed to read image data: %w", err)
	}

	img, imgType, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return ProcessedImage{}, fmt.Errorf("failed to decode image: %w", err)
	}
	mimeType := stdimgType2MimeType[imgType]

	res.Source = Variant{
		Data:         imgData,
		ContentType:  mimeType,
		Scale:        originalScale,
		ActualSize:   img.Bounds().Size(),
		IntendedSize: intendedSize(img.Bounds().Size(), originalScale),
	}

	formats := opts.Formats
	if len(formats) == 0 {
		formats = []string{mimeType} // use only the current encoding
	}

	for scale := originalScale; scale >= 1; scale-- {
		resized := img
		if scale != originalScale {
			width := img.Bounds().Size().X
			newWidth := width * scale / originalScale
			resized = resize.Resize(uint(newWidth), 0, img, resize.Bicubic)
		}

		for _, format := range formats {
			var outData []byte
			if format == mimeType && scale == originalScale {
				// Use the original image data with no further processing.
				outData = imgData
			} else {
				// Encode the resized data to the new output format.
				var outBuf bytes.Buffer
				encoder := mimeType2Encoder[format]
				err := encoder(&outBuf, resized)
				if err != nil {
					return ProcessedImage{}, fmt.Errorf("failed to encode resized image: %w", err)
				}
				outData = outBuf.Bytes()
			}

			res.Variants = append(res.Variants, Variant{
				Data:         outData,
				ContentType:  format,
				Scale:        scale,
				ActualSize:   resized.Bounds().Size(),
				IntendedSize: intendedSize(resized.Bounds().Size(), scale),
			})
		}
	}

	return res, nil
}

func intendedSize(actualSize image.Point, scale int) image.Point {
	return image.Point{
		X: actualSize.X / scale,
		Y: actualSize.Y / scale,
	}
}

type ImageOptions struct {
	Formats []string // If not provided, only the image's current format will be used.
}

// Refresh the image cache on a timer. (Someday this could be less dumb.)
// func init() {
// 	go func() {
// 		for {
// 			time.Sleep(10 * time.Second)
// 			imageCache.Range(func(key, value any) bool {
// 				storeImage(key.(string))
// 				return true
// 			})
// 		}
// 	}()
// }
