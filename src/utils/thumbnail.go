package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ThumbnailCardOptions struct {
	Title      string
	Subtitle   string
	CoverURL   string
	OutputDir  string
	OutputName string
}

func CreateSpotifyStyleCard(opts ThumbnailCardOptions) (string, error) {
	if opts.OutputDir == "" {
		opts.OutputDir = os.TempDir()
	}
	if opts.OutputName == "" {
		opts.OutputName = fmt.Sprintf("spotify_card_%d.jpg", time.Now().UnixNano())
	}

	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return "", err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, 1280, 720))
	drawBackground(canvas)
	drawBars(canvas)

	if strings.TrimSpace(opts.CoverURL) != "" {
		if cover, err := downloadImage(opts.CoverURL); err == nil {
			drawCover(canvas, cover, image.Pt(90, 150), 420)
		}
	}

	drawInfoPanel(canvas)

	outPath := filepath.Join(opts.OutputDir, sanitizeFileName(opts.OutputName))
	file, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := jpeg.Encode(file, canvas, &jpeg.Options{Quality: 90}); err != nil {
		return "", err
	}

	return outPath, nil
}

func drawBackground(img *image.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		ratio := float64(y) / float64(bounds.Dy())
		r := uint8(16 + 25*ratio)
		g := uint8(18 + 110*ratio)
		b := uint8(18 + 45*ratio)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
}

func drawBars(img *image.RGBA) {
	base := image.Rect(560, 120, 1210, 640)
	draw.Draw(img, base, &image.Uniform{C: color.RGBA{R: 0, G: 0, B: 0, A: 95}}, image.Point{}, draw.Over)

	x := 590
	for i := 0; i < 28; i++ {
		h := 40 + (i*17)%280
		rect := image.Rect(x, 560-h, x+14, 560)
		col := color.RGBA{R: 30, G: uint8(145 + (i * 3 % 80)), B: 70, A: 240}
		draw.Draw(img, rect, &image.Uniform{C: col}, image.Point{}, draw.Over)
		x += 21
	}

	progress := image.Rect(590, 590, 1130, 604)
	draw.Draw(img, progress, &image.Uniform{C: color.RGBA{R: 70, G: 70, B: 70, A: 230}}, image.Point{}, draw.Over)
	progressFill := image.Rect(590, 590, 860, 604)
	draw.Draw(img, progressFill, &image.Uniform{C: color.RGBA{R: 29, G: 185, B: 84, A: 255}}, image.Point{}, draw.Over)
}

func drawInfoPanel(img *image.RGBA) {
	draw.Draw(img, image.Rect(90, 585, 510, 640), &image.Uniform{C: color.RGBA{R: 0, G: 0, B: 0, A: 130}}, image.Point{}, draw.Over)
	draw.Draw(img, image.Rect(110, 603, 128, 621), &image.Uniform{C: color.RGBA{R: 29, G: 185, B: 84, A: 255}}, image.Point{}, draw.Over)
}

func drawCover(dst *image.RGBA, src image.Image, topLeft image.Point, size int) {
	resized := resizeImage(src, size, size)
	rect := image.Rect(topLeft.X, topLeft.Y, topLeft.X+size, topLeft.Y+size)
	draw.Draw(dst, rect, resized, image.Point{}, draw.Over)
}

func resizeImage(src image.Image, targetW, targetH int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	srcBounds := src.Bounds()

	for y := 0; y < targetH; y++ {
		for x := 0; x < targetW; x++ {
			sx := srcBounds.Min.X + x*srcBounds.Dx()/targetW
			sy := srcBounds.Min.Y + y*srcBounds.Dy()/targetH
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	return dst
}

func downloadImage(url string) (image.Image, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image download failed with status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_")
	clean := replacer.Replace(name)
	if clean == "" {
		return fmt.Sprintf("thumb_%d.jpg", time.Now().UnixNano())
	}
	if !strings.HasSuffix(strings.ToLower(clean), ".jpg") && !strings.HasSuffix(strings.ToLower(clean), ".jpeg") {
		clean += ".jpg"
	}
	return clean
}
