package main

import (
	// "fmt"
	"time"

	svg "github.com/ajstarks/svgo"
)

// --- Color Constants (Military / Service Colors) ---
const (
	RibbonRed    = "#d32f2f"
	RibbonBlue   = "#1976d2"
	RibbonGreen  = "#388e3c"
	RibbonYellow = "#fbc02d"
	RibbonWhite  = "#f5f5f5"
	RibbonBlack  = "#212121"
	RibbonPurple = "#7b1fa2"
	RibbonGold   = "#ffb300"
	RibbonOrange = "#f57c00"
	RibbonCyan   = "#00acc1"
	RibbonBrown  = "#5d4037"
	RibbonGrey   = "#616161"
)

type Stripe struct {
	Color string
	Width int // Relative width (e.g. 1, 2, 5)
}

type Ribbon struct {
	Name    string
	Rank    string // SSS, SS, S, AAA...
	Stripes []Stripe
}

// DrawRibbon renders a single ribbon at x, y
func DrawRibbon(s *svg.SVG, x, y, w, h int, r Ribbon) {
	// 1. Calculate total units for width distribution
	totalUnits := 0
	for _, stripe := range r.Stripes {
		totalUnits += stripe.Width
	}

	// 2. Draw stripes
	currentX := float64(x)
	unitWidth := float64(w) / float64(totalUnits)

	for _, stripe := range r.Stripes {
		width := unitWidth * float64(stripe.Width)
		// Draw rect
		s.Rect(int(currentX), y, int(width+1), h, "fill:"+stripe.Color+";stroke:none")
		currentX += width
	}
	
	// 3. Effects: Shine & Shadow for that "enamel/fabric" look
	s.Rect(x, y, w, h/4, "fill:white;fill-opacity:0.2") // Shine
	s.Rect(x, y+h-(h/4), w, h/4, "fill:black;fill-opacity:0.1") // Shadow
	
	// 4. Border
	s.Rect(x, y, w, h, "fill:none;stroke:#111;stroke-width:0.5;stroke-opacity:0.3")
}

// CalculateRibbons determines which ribbons a user has earned
// It sorts them by "Priority" (implicitly by append order)
func CalculateRibbons(data DashboardData) []Ribbon {
	var ribbons []Ribbon

	// -------------------------------------------------------------------------
	// 1. SERVICE MEDAL (Account Age) - "Experience"
	// -------------------------------------------------------------------------
	years := int(time.Since(data.AccountCreated).Hours() / 24 / 365)
	
	if years >= 20 { // SSS: Seasoned Veteran
		ribbons = append(ribbons, Ribbon{Name: "Seasoned Veteran", Rank: "SSS", Stripes: []Stripe{
			{RibbonPurple, 2}, {RibbonGold, 2}, {RibbonPurple, 2},
		}})
	} else if years >= 15 { // SS: Grandmaster
		ribbons = append(ribbons, Ribbon{Name: "Grandmaster", Rank: "SS", Stripes: []Stripe{
			{RibbonRed, 3}, {RibbonGold, 1}, {RibbonRed, 3},
		}})
	} else if years >= 10 { // S: Master Dev
		ribbons = append(ribbons, Ribbon{Name: "Master Dev", Rank: "S", Stripes: []Stripe{
			{RibbonBlue, 3}, {RibbonGold, 1}, {RibbonBlue, 3},
		}})
	} else if years >= 7 { // AAA: Expert Dev
		ribbons = append(ribbons, Ribbon{Name: "Expert Dev", Rank: "AAA", Stripes: []Stripe{
			{RibbonGreen, 3}, {RibbonYellow, 1}, {RibbonGreen, 3},
		}})
	} else if years >= 5 { // AA: Experienced Dev
		ribbons = append(ribbons, Ribbon{Name: "Experienced Dev", Rank: "AA", Stripes: []Stripe{
			{RibbonGreen, 3}, {RibbonWhite, 1}, {RibbonGreen, 3},
		}})
	} else if years >= 3 { // A: Intermediate Dev
		ribbons = append(ribbons, Ribbon{Name: "Intermediate Dev", Rank: "A", Stripes: []Stripe{
			{RibbonGreen, 4}, {RibbonWhite, 1}, {RibbonGreen, 1},
		}})
	} else { // Rookie
		ribbons = append(ribbons, Ribbon{Name: "Junior Dev", Rank: "B", Stripes: []Stripe{
			{RibbonGrey, 4}, {RibbonWhite, 2}, {RibbonGrey, 4},
		}})
	}

	// -------------------------------------------------------------------------
	// 2. STARS MEDAL (Popularity)
	// -------------------------------------------------------------------------
	if data.RawStars >= 2000 { // SSS
		ribbons = append(ribbons, Ribbon{Name: "Super Stargazer", Rank: "SSS", Stripes: []Stripe{
			{RibbonGold, 4}, {RibbonRed, 1}, {RibbonGold, 4},
		}})
	} else if data.RawStars >= 700 { // SS
		ribbons = append(ribbons, Ribbon{Name: "High Stargazer", Rank: "SS", Stripes: []Stripe{
			{RibbonGold, 3}, {RibbonBlue, 2}, {RibbonGold, 3},
		}})
	} else if data.RawStars >= 200 { // S
		ribbons = append(ribbons, Ribbon{Name: "Stargazer", Rank: "S", Stripes: []Stripe{
			{RibbonYellow, 3}, {RibbonBlue, 2}, {RibbonYellow, 3},
		}})
	} else if data.RawStars >= 100 { // AAA
		ribbons = append(ribbons, Ribbon{Name: "Super Star", Rank: "AAA", Stripes: []Stripe{
			{RibbonYellow, 4}, {RibbonWhite, 1}, {RibbonYellow, 4},
		}})
	} else if data.RawStars >= 50 { // AA
		ribbons = append(ribbons, Ribbon{Name: "High Star", Rank: "AA", Stripes: []Stripe{
			{RibbonBlue, 4}, {RibbonYellow, 1}, {RibbonBlue, 4},
		}})
	} else if data.RawStars >= 30 { // A
		ribbons = append(ribbons, Ribbon{Name: "Star", Rank: "A", Stripes: []Stripe{
			{RibbonBlue, 4}, {RibbonWhite, 1}, {RibbonBlue, 4},
		}})
	}

	// -------------------------------------------------------------------------
	// 3. COMMITS MEDAL (Effort)
	// -------------------------------------------------------------------------
	if data.RawCommits >= 4000 { // SSS
		ribbons = append(ribbons, Ribbon{Name: "God Committer", Rank: "SSS", Stripes: []Stripe{
			{RibbonRed, 2}, {RibbonBlack, 1}, {RibbonRed, 2}, {RibbonBlack, 1}, {RibbonRed, 2},
		}})
	} else if data.RawCommits >= 2000 { // SS
		ribbons = append(ribbons, Ribbon{Name: "Deep Committer", Rank: "SS", Stripes: []Stripe{
			{RibbonRed, 3}, {RibbonBlack, 2}, {RibbonRed, 3},
		}})
	} else if data.RawCommits >= 1000 { // S
		ribbons = append(ribbons, Ribbon{Name: "Super Committer", Rank: "S", Stripes: []Stripe{
			{RibbonRed, 3}, {RibbonWhite, 2}, {RibbonRed, 3},
		}})
	} else if data.RawCommits >= 500 { // AAA
		ribbons = append(ribbons, Ribbon{Name: "Ultra Committer", Rank: "AAA", Stripes: []Stripe{
			{RibbonOrange, 3}, {RibbonWhite, 2}, {RibbonOrange, 3},
		}})
	} else if data.RawCommits >= 200 { // AA
		ribbons = append(ribbons, Ribbon{Name: "Hyper Committer", Rank: "AA", Stripes: []Stripe{
			{RibbonOrange, 4}, {RibbonWhite, 1}, {RibbonOrange, 4},
		}})
	} else if data.RawCommits >= 100 { // A
		ribbons = append(ribbons, Ribbon{Name: "High Committer", Rank: "A", Stripes: []Stripe{
			{RibbonOrange, 4}, {RibbonYellow, 1}, {RibbonOrange, 4},
		}})
	}

	// -------------------------------------------------------------------------
	// 4. FOLLOWERS MEDAL (Influence)
	// -------------------------------------------------------------------------
	if data.RawFollowers >= 1000 { // SSS
		ribbons = append(ribbons, Ribbon{Name: "Super Celebrity", Rank: "SSS", Stripes: []Stripe{
			{RibbonPurple, 1}, {RibbonGold, 1}, {RibbonPurple, 1}, {RibbonGold, 1}, {RibbonPurple, 1},
		}})
	} else if data.RawFollowers >= 400 { // SS
		ribbons = append(ribbons, Ribbon{Name: "Ultra Celebrity", Rank: "SS", Stripes: []Stripe{
			{RibbonPurple, 2}, {RibbonWhite, 1}, {RibbonPurple, 2},
		}})
	} else if data.RawFollowers >= 200 { // S
		ribbons = append(ribbons, Ribbon{Name: "Hyper Celebrity", Rank: "S", Stripes: []Stripe{
			{RibbonBlue, 2}, {RibbonPurple, 2}, {RibbonBlue, 2},
		}})
	} else if data.RawFollowers >= 100 { // AAA
		ribbons = append(ribbons, Ribbon{Name: "Famous User", Rank: "AAA", Stripes: []Stripe{
			{RibbonCyan, 3}, {RibbonWhite, 2}, {RibbonCyan, 3},
		}})
	} else if data.RawFollowers >= 50 { // AA
		ribbons = append(ribbons, Ribbon{Name: "Active User", Rank: "AA", Stripes: []Stripe{
			{RibbonCyan, 4}, {RibbonWhite, 1}, {RibbonCyan, 4},
		}})
	}

	// -------------------------------------------------------------------------
	// 5. REPOSITORIES MEDAL (Creation)
	// -------------------------------------------------------------------------
	// Note: We need RawRepos count in data. Assuming we add it or just use TopRepos length as proxy (bad proxy).
	// Let's assume you add `RawRepos int` to DashboardData in fetcher.go, sourced from qStats.User.Repositories.TotalCount
	// For now, I will comment this out or use a dummy check if you haven't added RawRepos yet.
	/*
	if data.RawRepos >= 50 {
		ribbons = append(ribbons, Ribbon{Name: "God Repo Creator", Rank: "SSS", Stripes: []Stripe{
			{RibbonBlack, 3}, {RibbonGold, 2}, {RibbonBlack, 3},
		}})
	}
	*/

	// -------------------------------------------------------------------------
	// 6. SPECIAL RIBBONS
	// -------------------------------------------------------------------------
	
	// Polyglot (4+ Languages)
	if len(data.Languages) >= 4 {
		ribbons = append(ribbons, Ribbon{Name: "Polyglot", Rank: "S", Stripes: []Stripe{
			{RibbonRed, 1}, {RibbonGreen, 1}, {RibbonBlue, 1}, {RibbonYellow, 1},
		}})
	}

	// OG User (Joined 2008-2010)
	if data.AccountCreated.Year() <= 2010 {
		ribbons = append(ribbons, Ribbon{Name: "Ancient User", Rank: "SSS", Stripes: []Stripe{
			{RibbonBrown, 2}, {RibbonGold, 1}, {RibbonBrown, 2},
		}})
	}

	// Cap at 9 ribbons (3x3 grid)
	if len(ribbons) > 9 {
		ribbons = ribbons[:9]
	}

	return ribbons
}