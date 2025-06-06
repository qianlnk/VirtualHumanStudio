package storages

import (
	"bufio"
	"bytes"
	"image"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"strconv"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

// ResizeImage image resize
func ResizeImage(data []byte, width int, height int) (resBs []byte, err error) {
	if width == 0 && height == 0 {
		return data, nil
	}

	w1, h1 := getDimensions(data)
	if width == int(w1) && height == int(h1) {
		return data, nil
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		r := bytes.NewReader(data)
		img, err = webp.Decode(r)
		if err != nil {
			return nil, err
		}

		format = "webp"
	}

	if format != "webp" {
		imgfmt, err := imaging.FormatFromExtension(format)
		if err != nil {
			return nil, err
		}

		if imgfmt == imaging.GIF {
			return gifResize(data, width, height)
		}
	}

	dstImg := imaging.Resize(img, width, height, imaging.Lanczos)
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err = imaging.Encode(w, dstImg, imaging.JPEG, imaging.PNGCompressionLevel(png.BestCompression))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func gifResize(data []byte, width int, height int) ([]byte, error) {
	gifImg, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	dstImg := &gif.GIF{
		Image:     make([]*image.Paletted, len(gifImg.Image)),
		LoopCount: gifImg.LoopCount,
		Delay:     make([]int, len(gifImg.Delay)),
		Disposal:  make([]byte, len(gifImg.Disposal)),
	}

	var mu sync.Mutex
	token := make(chan struct{}, 10)
	var wg sync.WaitGroup

	xRate := 1.0
	yRate := 1.0
	for i, img := range gifImg.Image {
		token <- struct{}{}
		wg.Add(1)

		if i == 0 {
			if width != 0 && height == 0 {
				xRate = float64(img.Bounds().Dx()) / float64(width)
				yRate = xRate
			} else if height != 0 && width == 0 {
				yRate = float64(img.Bounds().Dy()) / float64(height)
				xRate = yRate
			} else if width != 0 && height != 0 {
				xRate = float64(img.Bounds().Dx()) / float64(width)
				yRate = float64(img.Bounds().Dy()) / float64(height)
			}
		}

		go func(i int, img *image.Paletted) {

			//resize时多一个像素点，防止因int去小数丢像素
			w := int(float64(img.Bounds().Dx())/float64(xRate)) + 1
			h := int(float64(img.Bounds().Dy())/float64(yRate)) + 1
			x0 := int(float64(img.Rect.Min.X) / float64(xRate))
			y0 := int(float64(img.Rect.Min.Y) / float64(yRate))
			x1 := int(float64(img.Rect.Max.X) / float64(xRate))
			y1 := int(float64(img.Rect.Max.Y) / float64(yRate))

			if i == 0 {
				if width != 0 {
					x1 = width
				}

				if height != 0 {
					y1 = height
				}
			}

			dst := imaging.Resize(img, w, h, imaging.Lanczos)
			p := image.NewPaletted(image.Rect(x0, y0, x1, y1), gifImg.Image[0].Palette)
			draw.Draw(p, p.Bounds(), dst, image.ZP, draw.Src)

			mu.Lock()
			dstImg.Image[i] = p
			dstImg.Delay[i] = gifImg.Delay[i]
			dstImg.Disposal[i] = gifImg.Disposal[i]
			mu.Unlock()

			wg.Done()
			<-token
		}(i, img)
	}

	wg.Wait()

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err = gif.EncodeAll(w, dstImg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func getDimensions(data []byte) (int, int) {
	format := getFormat(data)
	switch format {
	case "jpg":
		return getJpgDimensions(data)
	case "png":
		return getPngDimensions(data)
	case "bmp":
		return getBmpDimensions(data)
	case "gif":
		return getGifDimensions(data)
	default:
		return 0, 0
	}
}

func getFormat(data []byte) string {
	file := bytes.NewReader(data)
	head := make([]byte, 4)

	n, err := file.ReadAt(head, 0)
	if err != nil {
		return ""
	}

	if n < 4 {
		return ""
	}

	if head[0] == 0x89 && head[1] == 0x50 && head[2] == 0x4E && head[3] == 0x47 {
		return "png"
	}
	if head[0] == 0xFF && head[1] == 0xD8 {
		return "jpg"
	}
	if head[0] == 0x47 && head[1] == 0x49 && head[2] == 0x46 && head[3] == 0x38 {
		return "gif"
	}
	if head[0] == 0x42 && head[1] == 0x4D {
		return "bmp"
	}
	return ""
}

func getGifDimensions(data []byte) (width int, height int) {
	file := bytes.NewReader(data)

	bytes := make([]byte, 4)
	file.ReadAt(bytes, 6)
	width = int(bytes[0]) + int(bytes[1])*256
	height = int(bytes[2]) + int(bytes[3])*256
	return
}

func getBmpDimensions(data []byte) (width int, height int) {
	file := bytes.NewReader(data)

	bytes := make([]byte, 8)
	file.ReadAt(bytes, 18)
	width = int(bytes[3])<<24 | int(bytes[2])<<16 | int(bytes[1])<<8 | int(bytes[0])
	height = int(bytes[7])<<24 | int(bytes[6])<<16 | int(bytes[5])<<8 | int(bytes[4])
	return
}

func getPngDimensions(data []byte) (width int, height int) {
	file := bytes.NewReader(data)
	bytes := make([]byte, 8)
	file.ReadAt(bytes, 16)
	width = int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	height = int(bytes[4])<<24 | int(bytes[5])<<16 | int(bytes[6])<<8 | int(bytes[7])
	return
}

func getJpgDimensions(data []byte) (width int, height int) {
	file := bytes.NewReader(data)

	fileSize := file.Size()

	position := int64(4)
	bytes := make([]byte, 4)
	file.ReadAt(bytes[:2], position)
	length := int(bytes[0]<<8) + int(bytes[1])
	for position < fileSize {
		position += int64(length)
		file.ReadAt(bytes, position)
		length = int(bytes[2])<<8 + int(bytes[3])
		if (bytes[1] == 0xC0 || bytes[1] == 0xC2) && bytes[0] == 0xFF && length > 7 {
			file.ReadAt(bytes, position+5)
			width = int(bytes[2])<<8 + int(bytes[3])
			height = int(bytes[0])<<8 + int(bytes[1])
			return
		}
		position += 2
	}
	return 0, 0
}

func mustGetSize(size string) (int, int) {
	spl := strings.Split(size, "x")
	width, _ := strconv.Atoi(spl[0])
	height, _ := strconv.Atoi(spl[1])

	return width, height
}
