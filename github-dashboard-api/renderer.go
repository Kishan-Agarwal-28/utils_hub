package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	svg "github.com/ajstarks/svgo"
)

// --- Configuration ---
const (
	Width      = 1150
	Height     = 550 
	ColorBg    = "#151515"
	ColorText  = "#c5c8c6"
	ColorGreen = "#b5bd68"
	ColorBlue  = "#81a2be"
	ColorPurple= "#b294bb"
	ColorYellow= "#f0c674"
	ColorOrange= "#de935f"
	ColorRed   = "#cc6666"
	ColorDim   = "#373b41"
	FontFamily = "'Courier New', Courier, monospace"
)

// --- Types ---
type DashboardData struct {
	Username       string
	AvatarURL      string
	Followers      string
	Following      string
	RawFollowers   int
	RawCommits     int
	RawStars       int
	AccountCreated time.Time
	RawForks       int
	RawRepos       int
	RawPRs         int
	RawPRsMerged   int
	RawIssues      int
	RawContributed int
	TotalCommits   []int
	Languages      []LanguageItem
	StarHistory    []int
	ForkHistory    []int
	TopRepos       []RepoItem
	RecentActivity []ActivityItem
	TimeOfDay      []int

	AccountJoinedAt  interface{} 
}
type LanguageItem struct { Name string; Percent int }
type RepoItem struct { Name, Stars, Forks string }
type ActivityItem struct { Action, Repo, Type string }

// --- Renderer ---
type Renderer struct { canvas *svg.SVG }
func NewRenderer(w io.Writer) *Renderer { return &Renderer{canvas: svg.New(w)} }

func (r *Renderer) Render(data DashboardData) {
	canvas := r.canvas
	canvas.Start(Width, Height)
	r.defineDefs()

	canvas.Rect(0, 0, Width, Height, "fill:"+ColorBg)
	
	// Header
	canvas.Text(10, 20, fmt.Sprintf("github-stat-top v1.0 - User: %s", data.Username), 
		fmt.Sprintf("font-family:%s;font-size:14px;fill:%s", FontFamily, ColorText))

	// Layout - Left Side
	r.drawCommitsPanel(10, 35, 520, 140, data.TotalCommits)
	r.drawStatsPanel(10, 185, 520, 140, data.StarHistory, data.ForkHistory)
	r.drawActivityGraphPanel(10, 335, 400, 105, data.TimeOfDay)
	r.drawRecentActivityPanel(420, 335, 370, 105, data.RecentActivity)

	// Layout - Right Side
	r.drawLanguagesPanel(540, 35, 250, 140, data.Languages)
	r.drawTopReposPanel(540, 185, 250, 140, data.TopRepos)

	// Profile Panel (Top Right)
	// Height 405 -> Ends at y=440
	r.drawProfilePanel(810, 35, 330, 405, data)

	// NEW: Honors/Ribbon Panel (Bottom Right)
	// Starts at y=450 (10px gap), Height 90 -> Ends at y=540
	r.drawRibbonPanel(810, 450, 330, 90, data)
	r.drawSystemPanel(10, 450, 790, 90, data)

	canvas.End()
}

// --- Definitions ---
func (r *Renderer) defineDefs() {
	r.canvas.Def()
	
	makeGrad := func(id, color string) {
		r.canvas.LinearGradient(id, 0, 0, 0, 100, []svg.Offcolor{
			{Offset: 0, Color: color, Opacity: 1.0},
			{Offset: 100, Color: color, Opacity: 0.1},
		})
	}
	makeGrad("grad-green", ColorGreen)
	makeGrad("grad-yellow", ColorYellow)
	makeGrad("grad-orange", ColorOrange)
	makeGrad("grad-red", ColorRed)

	r.canvas.Pattern("pat-dots", 0, 0, 12, 16, "user")
	r.canvas.Rect(0, 0, 12, 16, "fill:black")
	r.canvas.Circle(2, 2, 1, "fill:white"); r.canvas.Circle(6, 2, 1, "fill:white")
	r.canvas.Circle(2, 6, 1, "fill:white"); r.canvas.Circle(6, 6, 1, "fill:white")
	r.canvas.Circle(2, 10, 1, "fill:white"); r.canvas.Circle(6, 10, 1, "fill:white")
	r.canvas.PatternEnd()

	r.canvas.Mask("mask-dotted", 0, 0, Width, Height)
	r.canvas.Rect(0, 0, Width, Height, "fill:url(#pat-dots)")
	r.canvas.MaskEnd()
	r.canvas.DefEnd()
}

// --- Panel Drawers ---

func (r *Renderer) drawCommitsPanel(x, y, w, h int, data []int) {
	r.drawRetroContainer(x, y, w, h, "CPU", "COMMITS OVER TIME", ColorGreen)
	r.drawLegend(x+20, y+20)
	r.drawAutoScaledChart(x+10, y+30, w-20, h-40, data, []int{5, 10, 20})
}

func (r *Renderer) drawLegend(x, y int) {
	colors := []string{ColorGreen, ColorYellow, ColorOrange, ColorRed}
	labels := []string{"Low", "Med", "High", "Crit"}
	style := fmt.Sprintf("font-family:%s;font-size:9px;fill:%s", FontFamily, ColorText)
	currentX := x
	for i, color := range colors {
		r.canvas.Rect(currentX, y, 8, 8, "fill:"+color)
		r.canvas.Text(currentX+12, y+7, labels[i], style)
		currentX += 45
	}
}

func (r *Renderer) drawStatsPanel(x, y, w, h int, stars, forks []int) {
	r.drawRetroContainer(x, y, w, h, "Memory", "ACCOUNT VELOCITY (Deltas)", ColorPurple)
	chartW := (w - 40) / 2
	r.canvas.Text(x+20, y+35, "STARS/DAY", "font-family:"+FontFamily+";font-size:12px;fill:"+ColorText)
	r.drawAutoScaledChart(x+10, y+45, chartW, h-55, stars, []int{2, 5, 10})

	r.canvas.Text(x+20+chartW, y+35, "FORKS/DAY", "font-family:"+FontFamily+";font-size:12px;fill:"+ColorText)
	r.drawAutoScaledChart(x+20+chartW, y+45, chartW, h-55, forks, []int{2, 5, 10})
}

func (r *Renderer) drawActivityGraphPanel(x, y, w, h int, data []int) {
	r.drawRetroContainer(x, y, w, h, "Bottom", "COMMIT ACTIVITY (Time of Day)", ColorGreen)
	r.drawAutoScaledChart(x+10, y+30, w-20, h-40, data, []int{5, 10, 15})
	
	steps := 8; stepW := float64(w-20) / float64(steps)
	for i := 0; i <= steps; i++ {
		hour := i * 3; if hour > 23 { continue }
		posX := x + 10 + int(float64(i)*stepW)
		r.canvas.Text(posX, y+h-5, fmt.Sprintf("%02d:00", hour), fmt.Sprintf("font-family:%s;font-size:10px;fill:%s;text-anchor:middle", FontFamily, ColorText))
	}
}

func (r *Renderer) drawProfilePanel(x, y, w, h int, data DashboardData) {
	r.drawRetroContainer(x, y, w, h, "Profile", "IDENTITY", ColorGreen)

	// 1. Avatar
	avatarSize := 280
	density := 2
	virtualSize := avatarSize * density
	imgX := x + (w-avatarSize)/2
	imgY := y + 40

	// Super-sample group
	r.canvas.Gtransform(fmt.Sprintf("translate(%d, %d) scale(%f)", imgX, imgY, 1.0/float64(density)))
	
	config := DitherConfig{
		Url:            data.AvatarURL,
		GridSize:       1,
		Contrast:       1.2,
		Brightness:     0.05,
		PrimaryColor:   "#f5f5f5", // Light Grey/White for face
		SecondaryColor: "#11011D", // Dark Purple for bg
	}

	DrawDitheredAvatar(r.canvas, 0, 0, virtualSize, virtualSize, config)
	r.canvas.Gend()

	// 2. Text Info
	textY := imgY + avatarSize + 35
	r.canvas.Text(x+w/2, textY, "@"+data.Username, 
		fmt.Sprintf("font-family:%s;font-size:20px;fill:%s;text-anchor:middle;font-weight:bold", FontFamily, ColorText))
	
	r.canvas.Text(x+w/2, textY+25, "Full-Stack Developer", 
		fmt.Sprintf("font-family:%s;font-size:12px;fill:%s;text-anchor:middle;opacity:0.8", FontFamily, ColorText))

	statsText := fmt.Sprintf("%s Followers · %s Following", data.Followers, data.Following)
	r.canvas.Text(x+w/2, textY+45, statsText, 
		fmt.Sprintf("font-family:%s;font-size:12px;fill:%s;text-anchor:middle;font-weight:bold;opacity:0.9", FontFamily, ColorGreen))
}

func (r *Renderer) drawSystemPanel(x, y, w, h int, data DashboardData) {
	r.drawRetroContainer(x, y, w, h, "System", "LIFETIME STATISTICS", ColorOrange)
	
	// Helper to format numbers like "46.8k"
	fmtNum := func(n int) string {
		if n >= 1000 { return fmt.Sprintf("%.1fk", float64(n)/1000.0) }
		return fmt.Sprintf("%d", n)
	}

	// Layout: 4 Columns
	// Col 1: Repos, Stars
	// Col 2: Forks, Commits
	// Col 3: PRs, Merged
	// Col 4: Issues, Contributed
	
	labelStyle := fmt.Sprintf("font-family:%s;font-size:11px;fill:%s", FontFamily, ColorText)
	valueStyle := fmt.Sprintf("font-family:%s;font-size:12px;fill:%s;font-weight:bold", FontFamily, ColorGreen)

	colW := w / 4
	startY := y + 35
	lineH := 25

	// Column 1
	r.canvas.Text(x+20, startY, "Total Repos:", labelStyle)
	r.canvas.Text(x+110, startY, fmt.Sprintf("%d", data.RawRepos), valueStyle)
	
	r.canvas.Text(x+20, startY+lineH, "Total Stars:", labelStyle)
	r.canvas.Text(x+110, startY+lineH, fmtNum(data.RawStars), valueStyle)

	// Column 2
	cx := x + colW
	r.canvas.Text(cx, startY, "Total Forks:", labelStyle)
	r.canvas.Text(cx+90, startY, fmtNum(data.RawForks), valueStyle)
	
	r.canvas.Text(cx, startY+lineH, "Total Commits:", labelStyle)
	r.canvas.Text(cx+90, startY+lineH, fmtNum(data.RawCommits), valueStyle)

	// Column 3
	cx = x + colW*2
	r.canvas.Text(cx, startY, "Total PRs:", labelStyle)
	r.canvas.Text(cx+80, startY, fmt.Sprintf("%d", data.RawPRs), valueStyle)
	
	r.canvas.Text(cx, startY+lineH, "PRs Merged:", labelStyle)
	r.canvas.Text(cx+80, startY+lineH, fmt.Sprintf("%d", data.RawPRsMerged), valueStyle)

	// Column 4
	cx = x + colW*3
	r.canvas.Text(cx, startY, "Total Issues:", labelStyle)
	r.canvas.Text(cx+90, startY, fmt.Sprintf("%d", data.RawIssues), valueStyle)
	
	r.canvas.Text(cx, startY+lineH, "Contributed:", labelStyle)
	r.canvas.Text(cx+90, startY+lineH, fmtNum(data.RawContributed), valueStyle)
}

// --- NEW RIBBON PANEL ---
func (r *Renderer) drawRibbonPanel(x, y, w, h int, data DashboardData) {
	// Container Title: "Honors" or "Decorations"
	r.drawRetroContainer(x, y, w, h, "Honors", "SERVICE RIBBONS", ColorYellow)

	ribbons := CalculateRibbons(data)
	
	// Ribbon Config
	ribbonW := 60
	ribbonH := 18
	gap := 5
	cols := 3
	
	// Center the rack horizontally
	rackW := cols*ribbonW + (cols-1)*gap
	startX := x + (w-rackW)/2
	
	// Center the rack vertically in the panel (accounting for header offset ~25px)
	// Available height for ribbons = h - 30
	// 2 rows = 18*2 + 5 = 41px height.
	// StartY should be around y + 35
	startY := y + 35 
	
	for i, rib := range ribbons {
		// Limit to 6 ribbons to fit in the small panel (2 rows of 3)
		if i >= 6 { break }
		
		row := i / cols
		col := i % cols
		
		rx := startX + col*(ribbonW+gap)
		ry := startY + row*(ribbonH+gap)
		
		DrawRibbon(r.canvas, rx, ry, ribbonW, ribbonH, rib)
	}
	
	// If no ribbons, show placeholder?
	if len(ribbons) == 0 {
		r.canvas.Text(x+w/2, y+h/2+5, "No Ribbons Earned Yet", 
			fmt.Sprintf("font-family:%s;font-size:10px;fill:%s;text-anchor:middle;opacity:0.5", FontFamily, ColorText))
	}
}

// --- Standard Helpers ---

func (r *Renderer) drawRetroContainer(x, y, w, h int, left, center, color string) {
	r.canvas.Roundrect(x, y, w, h, 5, 5, "fill:none;stroke:"+ColorDim+";stroke-width:1")
	if left != "" {
		lw := len(left) * 8; r.canvas.Rect(x+10, y-5, lw, 10, "fill:"+ColorBg)
		r.canvas.Text(x+10, y+4, left, fmt.Sprintf("font-family:%s;font-size:12px;fill:%s", FontFamily, ColorText))
	}
	if center != "" {
		lw := len(center) * 8; sx := x + (w/2) - (lw/2); r.canvas.Rect(sx, y-5, lw, 10, "fill:"+ColorBg)
		r.canvas.Text(sx, y+4, center, fmt.Sprintf("font-family:%s;font-size:12px;fill:%s", FontFamily, ColorText))
	}
}

func (r *Renderer) drawLanguagesPanel(x, y, w, h int, langs []LanguageItem) {
	r.drawRetroContainer(x, y, w, h, "", "LANGUAGES", ColorBlue)
	startY := y + 40
	for i, l := range langs {
		if i >= 5 { break }
		color := ColorText; if i == 0 { color = ColorGreen } else if i == 1 { color = ColorBlue }
		r.canvas.Text(x+20, startY, l.Name+":", fmt.Sprintf("font-family:%s;font-size:14px;fill:%s", FontFamily, color))
		r.canvas.Text(x+w-20, startY, fmt.Sprintf("%d%%", l.Percent), fmt.Sprintf("font-family:%s;font-size:14px;fill:%s;text-anchor:end", FontFamily, ColorText))
		startY += 20
	}
}

func (r *Renderer) drawTopReposPanel(x, y, w, h int, repos []RepoItem) {
	r.drawRetroContainer(x, y, w, h, "Disks", "TOP REPOSITORIES", ColorYellow)
	startY := y + 40
	for i, repo := range repos {
		if i >= 3 { break }
		r.canvas.Text(x+15, startY, "★", "font-size:12px;fill:"+ColorGreen)
		r.canvas.Text(x+30, startY, repo.Name+":", fmt.Sprintf("font-family:%s;font-size:13px;fill:%s", FontFamily, ColorText))
		r.canvas.Text(x+30, startY+15, fmt.Sprintf("%s stars, %s forks", repo.Stars, repo.Forks), fmt.Sprintf("font-family:%s;font-size:11px;fill:%s", FontFamily, ColorGreen))
		startY += 40
	}
}

func (r *Renderer) drawRecentActivityPanel(x, y, w, h int, activity []ActivityItem) {
	r.drawRetroContainer(x, y, w, h, "Processes", "RECENT ACTIVITY (Last 24h)", ColorPurple)
	startY := y + 40
	for i, act := range activity {
		if i >= 4 { break }
		r.drawIcon(x+15, startY-8, act.Type)
		text := fmt.Sprintf("%s %s", act.Action, act.Repo)
		if act.Type == "none" { text = act.Action }
		if len(text) > 42 { text = text[:39] + "..." }
		r.canvas.Text(x+35, startY, text, fmt.Sprintf("font-family:%s;font-size:12px;fill:%s", FontFamily, ColorText))
		startY += 20
	}
}

func (r *Renderer) drawIcon(x, y int, t string) {
	t = strings.ToLower(t)
	strk := func(c string) string { return "fill:none;stroke:" + c + ";stroke-width:1.5" }
	fill := func(c string) string { return "stroke:none;fill:" + c }
	switch {
	case strings.Contains(t, "push"):
		r.canvas.Circle(x+3, y+10, 2, strk(ColorGreen)); r.canvas.Circle(x+3, y+3, 2, strk(ColorGreen)); r.canvas.Line(x+3, y+5, x+3, y+8, strk(ColorGreen))
	case strings.Contains(t, "pr"):
		r.canvas.Circle(x+3, y+10, 2, strk(ColorBlue)); r.canvas.Circle(x+3, y+3, 2, strk(ColorBlue)); r.canvas.Line(x+3, y+5, x+3, y+8, strk(ColorBlue)); r.canvas.Path(fmt.Sprintf("M%d,%d C%d,%d %d,%d %d,%d", x+3, y+8, x+3, y+5, x+8, y+5, x+8, y+3), strk(ColorBlue))
	case strings.Contains(t, "issue"):
		r.canvas.Circle(x+6, y+6, 5, strk(ColorRed)); r.canvas.Line(x+6, y+4, x+6, y+7, strk(ColorRed)); r.canvas.Circle(x+6, y+9, 1, fill(ColorRed))
	case strings.Contains(t, "none"):
		r.canvas.Circle(x+6, y+6, 2, fill(ColorDim))
	default:
		r.canvas.Circle(x+6, y+6, 3, fill(ColorDim))
	}
}

func (r *Renderer) drawAutoScaledChart(x, y, w, h int, data []int, thresholds []int) {
	if len(data) < 2 { return }

	dataMax := 0
	for _, v := range data { if v > dataMax { dataMax = v } }
	
	gradUrl := "url(#grad-green)"
	strokeColor := ColorGreen

	if dataMax > thresholds[2] {
		gradUrl = "url(#grad-red)"
		strokeColor = ColorRed
	} else if dataMax > thresholds[1] {
		gradUrl = "url(#grad-orange)"
		strokeColor = ColorOrange
	} else if dataMax > thresholds[0] {
		gradUrl = "url(#grad-yellow)"
		strokeColor = ColorYellow
	}

	scaleMax := dataMax
	if scaleMax == 0 { scaleMax = 1 } 

	peakText := fmt.Sprintf("Peak: %d", dataMax)
	r.canvas.Text(x+w-5, y+10, peakText, fmt.Sprintf("font-family:%s;font-size:10px;fill:%s;text-anchor:end;opacity:0.7", FontFamily, ColorText))

	xStep := float64(w) / float64(len(data)-1)
	var xPts, yPts []int
	for i, val := range data {
		effVal := val
		px := x + int(float64(i)*xStep)
		py := y + h - int((float64(effVal)/float64(scaleMax))*float64(h))
		xPts = append(xPts, px); yPts = append(yPts, py)
	}
	
	fX := append([]int{}, xPts...); fY := append([]int{}, yPts...)
	fX = append(fX, x+w, x); fY = append(fY, y+h, y+h)
	r.canvas.Polygon(fX, fY, fmt.Sprintf("fill:%s;mask:url(#mask-dotted);stroke:none", gradUrl))
	r.canvas.Polyline(xPts, yPts, "fill:none;stroke:"+strokeColor+";stroke-width:1.5")
}