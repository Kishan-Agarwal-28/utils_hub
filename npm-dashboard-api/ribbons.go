package main

import (

	"time"

	svg "github.com/ajstarks/svgo"
)

// Colors
const (
	RibbonRed    = "#cb3837" // NPM Red
	RibbonBlue   = "#1976d2"
	RibbonGreen  = "#388e3c"
	RibbonYellow = "#fbc02d"
	RibbonWhite  = "#f5f5f5"
	RibbonBlack  = "#212121"
	RibbonPurple = "#7b1fa2"
	RibbonGold   = "#ffb300"
	RibbonGrey   = "#616161"
)

type Stripe struct {
	Color string
	Width int 
}

type Ribbon struct {
	Name    string
	Rank    string
	Stripes []Stripe
}

func DrawRibbon(s *svg.SVG, x, y, w, h int, r Ribbon) {
	totalUnits := 0
	for _, stripe := range r.Stripes { totalUnits += stripe.Width }

	currentX := float64(x)
	unitWidth := float64(w) / float64(totalUnits)

	for _, stripe := range r.Stripes {
		width := unitWidth * float64(stripe.Width)
		s.Rect(int(currentX), y, int(width+1), h, "fill:"+stripe.Color+";stroke:none")
		currentX += width
	}
	
	s.Rect(x, y, w, h/4, "fill:white;fill-opacity:0.2")
	s.Rect(x, y+h-(h/4), w, h/4, "fill:black;fill-opacity:0.1")
	s.Rect(x, y, w, h, "fill:none;stroke:#111;stroke-width:0.5;stroke-opacity:0.3")
}

// CalculateNpmRibbons - ADAPTED FOR NPM DATA
func CalculateNpmRibbons(data NpmDashboardData) []Ribbon {
	var ribbons []Ribbon

	// 1. DOWNLOADS MEDAL (Popularity)
	// Replaces "Stars"
	if data.TotalDownloads >= 10000000 { // 10M
		ribbons = append(ribbons, Ribbon{Name: "npm Legend", Rank: "SSS", Stripes: []Stripe{
			{RibbonGold, 4}, {RibbonRed, 2}, {RibbonGold, 4},
		}})
	} else if data.TotalDownloads >= 1000000 { // 1M
		ribbons = append(ribbons, Ribbon{Name: "1M Club", Rank: "SS", Stripes: []Stripe{
			{RibbonRed, 3}, {RibbonWhite, 1}, {RibbonRed, 3},
		}})
	} else if data.TotalDownloads >= 100000 { // 100k
		ribbons = append(ribbons, Ribbon{Name: "High Traffic", Rank: "S", Stripes: []Stripe{
			{RibbonRed, 4}, {RibbonBlack, 1}, {RibbonRed, 4},
		}})
	} else if data.TotalDownloads >= 10000 { // 10k
		ribbons = append(ribbons, Ribbon{Name: "Rising Pkg", Rank: "A", Stripes: []Stripe{
			{RibbonGrey, 3}, {RibbonRed, 1}, {RibbonGrey, 3},
		}})
	}

	// 2. PACKAGES MEDAL (Prolific)
	// Replaces "Commits"
	if data.TotalPackages >= 100 {
		ribbons = append(ribbons, Ribbon{Name: "Registry God", Rank: "SSS", Stripes: []Stripe{
			{RibbonBlack, 2}, {RibbonGold, 1}, {RibbonBlack, 2},
		}})
	} else if data.TotalPackages >= 50 {
		ribbons = append(ribbons, Ribbon{Name: "Heavy Publisher", Rank: "SS", Stripes: []Stripe{
			{RibbonBlack, 3}, {RibbonWhite, 1}, {RibbonBlack, 3},
		}})
	} else if data.TotalPackages >= 10 {
		ribbons = append(ribbons, Ribbon{Name: "Publisher", Rank: "A", Stripes: []Stripe{
			{RibbonGrey, 4}, {RibbonWhite, 1}, {RibbonGrey, 4},
		}})
	}

	// 3. SERVICE MEDAL (Time since first package)
	if !data.FirstPackageDate.IsZero() {
		years := int(time.Since(data.FirstPackageDate).Hours() / 24 / 365)
		if years >= 10 {
			ribbons = append(ribbons, Ribbon{Name: "NPM Veteran", Rank: "S", Stripes: []Stripe{
				{RibbonPurple, 3}, {RibbonGold, 1}, {RibbonPurple, 3},
			}})
		} else if years >= 5 {
			ribbons = append(ribbons, Ribbon{Name: "Senior Dev", Rank: "A", Stripes: []Stripe{
				{RibbonBlue, 3}, {RibbonWhite, 1}, {RibbonBlue, 3},
			}})
		}
	}

	return ribbons
}