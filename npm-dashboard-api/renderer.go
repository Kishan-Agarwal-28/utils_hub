package main

import (
	"fmt"
	"io"

	svg "github.com/ajstarks/svgo"
)

// --- NPM RED THEME ---
const (
	Width = 1150; Height = 550
	ColorBg = "#151515"; ColorText = "#e0e0e0"
	ColorNpmRed = "#cb3837"; ColorWhite  = "#ffffff"; ColorGrey = "#333333"; ColorDim = "#282828"; ColorYellow = "#fbc02d"
	FontFamily = "'Courier New', Courier, monospace"
)

type Renderer struct { canvas *svg.SVG }
func NewRenderer(w io.Writer) *Renderer { return &Renderer{canvas: svg.New(w)} }

func (r *Renderer) Render(data NpmDashboardData) {
	canvas := r.canvas
	canvas.Start(Width, Height)
	r.defineDefs()

	canvas.Rect(0, 0, Width, Height, "fill:"+ColorBg)

	// Header
	canvas.Text(10, 20, fmt.Sprintf("npm-stat-pro v1.0 - Maintainer: %s", data.Username), 
		fmt.Sprintf("font-family:%s;font-size:14px;fill:%s", FontFamily, ColorText))

	// 1. Download Graph (Main)
	r.drawDownloadsPanel(10, 35, 520, 140, data.DownloadHistory)
	
	// 2. Top Packages (Top Right)
	r.drawTopPackagesPanel(540, 35, 250, 140, data.TopPackages)
	
	// 3. Stats Grid (Bottom Left) - Now visually richer
	r.drawNpmStatsPanel(10, 185, 520, 140, data)
	
	// 4. Activity List
	r.drawActivityPanel(10, 335, 400, 105, data.RecentReleases)
	
	// 5. System/Summary
	r.drawSystemPanel(420, 335, 370, 105, data)

	// 6. Profile & Ribbons
	r.drawProfilePanel(810, 35, 330, 405, data)
	r.drawRibbonPanel(810, 450, 330, 90, data)

	canvas.End()
}

func (r *Renderer) defineDefs() {
	r.canvas.Def()
	r.canvas.LinearGradient("grad-npm", 0, 0, 0, 100, []svg.Offcolor{
		{Offset: 0, Color: ColorNpmRed, Opacity: 1.0},
		{Offset: 100, Color: ColorNpmRed, Opacity: 0.1},
	})
	r.canvas.Pattern("pat-dots", 0, 0, 12, 16, "user")
	r.canvas.Rect(0, 0, 12, 16, "fill:#1a1a1a")
	r.canvas.Circle(2, 2, 1, "fill:#444")
	r.canvas.PatternEnd()
	r.canvas.Mask("mask-dotted", 0, 0, Width, Height)
	r.canvas.Rect(0, 0, Width, Height, "fill:url(#pat-dots)")
	r.canvas.MaskEnd()
	r.canvas.DefEnd()
}

func (r *Renderer) drawDownloadsPanel(x, y, w, h int, data []int) {
	r.drawRetroContainer(x, y, w, h, "Network", "TRAFFIC (Top Pkg History)", ColorNpmRed)
	r.drawAutoScaledChart(x+10, y+30, w-20, h-40, data, ColorNpmRed, "url(#grad-npm)")
}

func (r *Renderer) drawTopPackagesPanel(x, y, w, h int, pkgs []PackageItem) {
	r.drawRetroContainer(x, y, w, h, "Registry", "TOP PACKAGES (Year)", ColorWhite)
	startY := y + 40
	for i, p := range pkgs {
		if i >= 3 { break }
		r.canvas.Text(x+15, startY, "📦", "font-size:12px;fill:"+ColorNpmRed)
		r.canvas.Text(x+35, startY, p.Name, "font-family:"+FontFamily+";font-size:13px;fill:"+ColorText)
		
		// Add Download Count next to version for density
		meta := fmt.Sprintf("%s · %s/yr", p.Version, p.Downloads)
		r.canvas.Text(x+35, startY+15, meta, "font-family:"+FontFamily+";font-size:11px;fill:"+ColorGrey)
		startY += 40
	}
}

// UPDATED: Now includes a "Heatmap" style visual for Publish Activity
func (r *Renderer) drawNpmStatsPanel(x, y, w, h int, data NpmDashboardData) {
	r.drawRetroContainer(x, y, w, h, "Metrics", "PUBLISH ACTIVITY (12 Mo)", ColorNpmRed)
	
	// Draw Stats Numbers (Top Half)
	colW := w/2; startY := y+35
	r.canvas.Text(x+20, startY, "TOTAL PACKAGES", "font-family:"+FontFamily+";font-size:10px;fill:"+ColorGrey)
	r.canvas.Text(x+20, startY+15, fmt.Sprintf("%d", data.TotalPackages), "font-family:"+FontFamily+";font-size:20px;fill:"+ColorWhite)

	r.canvas.Text(x+colW+20, startY, "TOTAL DOWNLOADS", "font-family:"+FontFamily+";font-size:10px;fill:"+ColorGrey)
	r.canvas.Text(x+colW+20, startY+15, formatNumber(data.TotalDownloads), "font-family:"+FontFamily+";font-size:20px;fill:"+ColorNpmRed)

	// Draw "Heatmap" Bars (Bottom Half) - Fills empty space
	barW := (w - 40) / 12
	barBaseY := y + h - 15
	maxPub := 0; for _, v := range data.PublishActivity { if v > maxPub { maxPub = v } }
	if maxPub == 0 { maxPub = 1 }

	r.canvas.Text(x+20, startY+40, "MONTHLY RELEASE CADENCE:", "font-family:"+FontFamily+";font-size:9px;fill:"+ColorGrey)
	
	for i, v := range data.PublishActivity {
		bh := int((float64(v) / float64(maxPub)) * 30)
		if bh < 2 { bh = 2 } // min height
		
		bx := x + 20 + (i * barW)
		// Color logic: High activity = Red, Low = Grey
		color := ColorGrey
		if v > 0 { color = ColorNpmRed }
		
		r.canvas.Rect(bx, barBaseY-bh, barW-4, bh, "fill:"+color)
	}
}

func (r *Renderer) drawActivityPanel(x, y, w, h int, items []ActivityItem) {
	r.drawRetroContainer(x, y, w, h, "Log", "RECENT PUBLISHES", ColorWhite)
	startY := y + 40
	for i, item := range items {
		if i >= 4 { break } // Show up to 4 now
		r.canvas.Circle(x+15, startY-4, 3, "fill:"+ColorNpmRed)
		
		text := item.Repo
		if len(text) > 42 { text = text[:39] + "..." }
		r.canvas.Text(x+30, startY, text, "font-family:"+FontFamily+";font-size:12px;fill:"+ColorText)
		startY += 20
	}
}

func (r *Renderer) drawSystemPanel(x, y, w, h int, data NpmDashboardData) {
	r.drawRetroContainer(x, y, w, h, "System", "STATUS", ColorGrey)
	r.canvas.Text(x+20, y+40, "REGISTRY: ONLINE", "font-family:"+FontFamily+";font-size:12px;fill:"+ColorText)
	r.canvas.Text(x+20, y+60, "USER: ACTIVE", "font-family:"+FontFamily+";font-size:12px;fill:"+ColorText)
	r.canvas.Text(x+20, y+80, fmt.Sprintf("SCORE: %s", "A+"), "font-family:"+FontFamily+";font-size:12px;fill:"+ColorNpmRed)
}

func (r *Renderer) drawProfilePanel(x, y, w, h int, data NpmDashboardData) {
	r.drawRetroContainer(x, y, w, h, "Identity", "MAINTAINER", ColorNpmRed)
	
	// Dithered Avatar
	avatarSize := 180; density := 2; virtualSize := avatarSize * density
	imgX := x + (w-avatarSize)/2; imgY := y + 30
	
	r.canvas.Gtransform(fmt.Sprintf("translate(%d, %d) scale(%f)", imgX, imgY, 1.0/float64(density)))
	config := DitherConfig{
		Url: data.AvatarURL, GridSize: 1, Contrast: 1.2, Brightness: 0.05,
		PrimaryColor: "#f5f5f5", SecondaryColor: "#11011D",
	}
	DrawDitheredAvatar(r.canvas, 0, 0, virtualSize, virtualSize, config)
	r.canvas.Gend()

	textY := imgY + avatarSize + 25
	r.canvas.Text(x+w/2, textY, data.Username, "font-family:"+FontFamily+";font-size:20px;fill:"+ColorWhite+";text-anchor:middle;font-weight:bold")
	r.canvas.Text(x+w/2, textY+25, "Package Maintainer", "font-family:"+FontFamily+";font-size:11px;fill:"+ColorText+";text-anchor:middle;opacity:0.8")
}

func (r *Renderer) drawRibbonPanel(x, y, w, h int, data NpmDashboardData) {
	r.drawRetroContainer(x, y, w, h, "Honors", "ACHIEVEMENTS", ColorYellow)
	ribbons := CalculateNpmRibbons(data)
	ribW := 60; ribH := 18; gap := 5; cols := 3
	rackW := cols*ribW + (cols-1)*gap
	sx := x + (w-rackW)/2; sy := y + 35
	
	for i, rib := range ribbons {
		if i >= 6 { break }
		row := i/cols; col := i%cols
		rx := sx + col*(ribW+gap); ry := sy + row*(ribH+gap)
		DrawRibbon(r.canvas, rx, ry, ribW, ribH, rib)
	}
}

// ... Helpers (drawRetroContainer, drawAutoScaledChart) ...
// (Reuse existing helpers from previous code)
func (r *Renderer) drawRetroContainer(x, y, w, h int, left, center, color string) {
	r.canvas.Roundrect(x, y, w, h, 5, 5, "fill:none;stroke:"+ColorDim+";stroke-width:1")
	if left != "" {
		lw := len(left) * 9; r.canvas.Rect(x+10, y-5, lw, 10, "fill:"+ColorBg)
		r.canvas.Text(x+10, y+4, left, fmt.Sprintf("font-family:%s;font-size:12px;fill:%s", FontFamily, ColorText))
	}
	if center != "" {
		lw := len(center) * 9; sx := x + (w/2) - (lw/2); r.canvas.Rect(sx, y-5, lw, 10, "fill:"+ColorBg)
		r.canvas.Text(sx, y+4, center, fmt.Sprintf("font-family:%s;font-size:12px;fill:%s", FontFamily, ColorText))
	}
}

func (r *Renderer) drawAutoScaledChart(x, y, w, h int, data []int, strokeColor, fillDef string) {
	if len(data) < 2 { return }
	dataMax := 0; for _, v := range data { if v > dataMax { dataMax = v } }
	if dataMax == 0 { dataMax = 1 }
	r.canvas.Text(x+w-5, y+10, fmt.Sprintf("Peak: %d", dataMax), "font-family:"+FontFamily+";font-size:10px;fill:"+ColorGrey+";text-anchor:end")

	xStep := float64(w) / float64(len(data)-1)
	var xPts, yPts []int
	for i, val := range data {
		px := x + int(float64(i)*xStep)
		py := y + h - int((float64(val)/float64(dataMax))*float64(h))
		xPts = append(xPts, px); yPts = append(yPts, py)
	}
	fX := append([]int{}, xPts...); fY := append([]int{}, yPts...)
	fX = append(fX, x+w, x); fY = append(fY, y+h, y+h)
	
	r.canvas.Polygon(fX, fY, fmt.Sprintf("fill:%s;mask:url(#mask-dotted);stroke:none", fillDef))
	r.canvas.Polyline(xPts, yPts, "fill:none;stroke:"+strokeColor+";stroke-width:2")
}

