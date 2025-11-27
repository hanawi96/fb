package imageprocessor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// AddCaptionToImage adds a caption overlay to an image
func AddCaptionToImage(imageURL, caption string) ([]byte, error) {
	if caption == "" {
		// No caption, return original image
		return downloadImage(imageURL)
	}

	// Download image
	imgData, err := downloadImage(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create new image with caption overlay
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	// Add caption overlay at bottom
	addCaptionOverlay(newImg, caption)

	// Encode back to bytes
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, newImg, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(buf, newImg)
	default:
		err = jpeg.Encode(buf, newImg, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func addCaptionOverlay(img *image.RGBA, caption string) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate overlay height (10% of image height, min 60px, max 150px)
	overlayHeight := height / 10
	if overlayHeight < 60 {
		overlayHeight = 60
	}
	if overlayHeight > 150 {
		overlayHeight = 150
	}

	// Draw semi-transparent black background at bottom
	overlayRect := image.Rect(0, height-overlayHeight, width, height)
	overlayColor := color.RGBA{0, 0, 0, 180} // Black with 70% opacity
	draw.Draw(img, overlayRect, &image.Uniform{overlayColor}, image.Point{}, draw.Over)

	// Wrap text to fit width
	lines := wrapText(caption, width-40) // 20px padding on each side

	// Draw text (white, centered)
	textColor := color.RGBA{255, 255, 255, 255}
	face := basicfont.Face7x13

	// Calculate starting Y position to center text vertically
	lineHeight := 15
	totalTextHeight := len(lines) * lineHeight
	startY := height - overlayHeight + (overlayHeight-totalTextHeight)/2 + 13

	point := fixed.Point26_6{
		X: fixed.I(20), // 20px left padding
		Y: fixed.I(startY),
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  point,
	}

	for _, line := range lines {
		drawer.Dot.X = fixed.I(20)
		drawer.DrawString(line)
		drawer.Dot.Y += fixed.I(lineHeight)
	}
}

func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := ""
	charWidth := 7 // basicfont.Face7x13 character width

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine)*charWidth > maxWidth-40 {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Word too long, add it anyway
				lines = append(lines, word)
			}
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Limit to 3 lines
	if len(lines) > 3 {
		lines = lines[:3]
		lines[2] = lines[2][:len(lines[2])-3] + "..."
	}

	return lines
}
