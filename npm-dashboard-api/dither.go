package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"strconv"

	svg "github.com/ajstarks/svgo"
)

var bayerMatrix4x4 = [4][4]float64{
	{0, 8, 2, 10}, {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5},
}

type DitherConfig struct {
	Url            string
	GridSize       int
	Contrast       float64
	Brightness     float64
	PrimaryColor   string
	SecondaryColor string
}

func DrawDitheredAvatar(s *svg.SVG, startX, startY, displayWidth, displayHeight int, cfg DitherConfig) {
	resp, err := http.Get(cfg.Url)
	if err != nil { return }
	defer resp.Body.Close()

	srcImg, _, err := image.Decode(resp.Body)
	if err != nil { return }

	// 1. Draw Background
	s.Rect(startX, startY, displayWidth, displayHeight, "fill:"+cfg.SecondaryColor)

	// 2. Setup
	rPrim, gPrim, bPrim := parseHexColor(cfg.PrimaryColor)
	fillStr := fmt.Sprintf("fill:rgb(%d,%d,%d)", rPrim, gPrim, bPrim)
	
	srcW, srcH := float64(srcImg.Bounds().Dx()), float64(srcImg.Bounds().Dy())
	
	s.Gtransform(fmt.Sprintf("translate(%d, %d)", startX, startY))

	// 3. Loop Rows
	for y := 0; y < displayHeight; y += cfg.GridSize {
		
		// RLE: Track the current run of "active" pixels
		runStart := -1
		
		for x := 0; x < displayWidth; x += cfg.GridSize {
			
			// --- Pixel Sampling Logic ---
			srcX := int(math.Floor((float64(x) / float64(displayWidth)) * srcW))
			srcY := int(math.Floor((float64(y) / float64(displayHeight)) * srcH))
			
			r, g, b, _ := srcImg.At(srcImg.Bounds().Min.X+srcX, srcImg.Bounds().Min.Y+srcY).RGBA()
			
			// Color Adjust
			r8 := clamp((float64(r>>8)-128)*cfg.Contrast+128+cfg.Brightness*255, 0, 255)
			g8 := clamp((float64(g>>8)-128)*cfg.Contrast+128+cfg.Brightness*255, 0, 255)
			b8 := clamp((float64(b>>8)-128)*cfg.Contrast+128+cfg.Brightness*255, 0, 255)
			
			lum := (0.299*r8 + 0.587*g8 + 0.114*b8) / 255.0
			
			matrixX := (x / cfg.GridSize) % 4
			matrixY := (y / cfg.GridSize) % 4
			threshold := bayerMatrix4x4[matrixY][matrixX] / 16.0

			isActive := lum >= threshold

			// --- RLE Logic ---
			if isActive {
				if runStart == -1 {
					runStart = x // Start new run
				}
				// If already in a run, just continue loop
			} else {
				if runStart != -1 {
					// End of run, draw the accumulated rect
					width := x - runStart
					s.Rect(runStart, y, width, cfg.GridSize, fillStr)
					runStart = -1 // Reset
				}
			}
		}
		// End of row: if a run was active, draw it
		if runStart != -1 {
			width := displayWidth - runStart
			s.Rect(runStart, y, width, cfg.GridSize, fillStr)
		}
	}
	s.Gend()
}

// ... (Helpers: clamp, parseHexColor same as before) ...
func clamp(val, min, max float64) float64 {
	if val < min { return min }
	if val > max { return max }
	return val
}

func parseHexColor(s string) (int, int, int) {
	if len(s) > 0 && s[0] == '#' { s = s[1:] }
	if len(s) == 3 { s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]}) }
	if len(s) != 6 { return 0, 0, 0 }
	rgb, err := strconv.ParseUint(s, 16, 32)
	if err != nil { return 0, 0, 0 }
	return int(rgb >> 16), int((rgb >> 8) & 0xFF), int(rgb & 0xFF)
}