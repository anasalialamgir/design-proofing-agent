package main

import (
	"fmt"
	"image"
	"os"

	"github.com/fogleman/gg"
)

func BuildProof(artworkPath, outputPath string, orderID string) error {
	file, err := os.Open(artworkPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	w := float64(img.Bounds().Dx()) + 120.0
	h := float64(img.Bounds().Dy()) + 200.0

	dc := gg.NewContext(int(w), int(h))
	dc.SetRGB(0.95, 0.95, 0.95)
	dc.Clear()

	// Header
	dc.SetRGB(0.1, 0.1, 0.1)
	dc.DrawRectangle(20, 20, w-40, 60)
	dc.Fill()

	dc.SetRGB(1, 1, 1)
	dc.DrawString(fmt.Sprintf("SAMPLE PROOF - ORDER #%s", orderID), 40, 55)

	// Image and Dotted Cut Lines
	dc.DrawImage(img, 60, 100)
	dc.SetRGBA(0.9, 0.1, 0.1, 0.8)
	dc.SetLineWidth(2.0)
	dc.DrawRectangle(69, 109, float64(img.Bounds().Dx())-18, float64(img.Bounds().Dy())-18)
	dc.Stroke()

	// Watermark
	dc.SetRGBA(0.8, 0.2, 0.2, 0.2)
	dc.DrawStringAnchored("UNAPPROVED PROOF", w/2, h/2, 0.5, 0.5)

	return dc.SavePNG(outputPath)
}
