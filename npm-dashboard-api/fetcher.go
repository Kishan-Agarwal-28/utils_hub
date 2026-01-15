package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// --- Structs ---
type NpmDashboardData struct {
	Username       string
	AvatarURL      string
	
	// Stats
	TotalPackages    int
	TotalDownloads   int
	FirstPackageDate time.Time // For "Years Active"
	
	// Lists
	TopPackages    []PackageItem
	RecentReleases []ActivityItem
	
	// Graphs
	DownloadHistory []int 
	PublishActivity []int // Heatmap data
}

type PackageItem struct { Name, Version, Downloads string; RawDownloads int }
type ActivityItem struct { Action, Repo, Type string }

// --- API Response Structures ---
type npmSearchResp struct {
	Total   int `json:"total"`
	Objects []struct {
		Package struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Date    string `json:"date"`
			Publisher struct { Username string } `json:"publisher"`
		} `json:"package"`
	} `json:"objects"`
}

type npmRangeResp struct {
	Downloads []struct {
		Downloads int    `json:"downloads"`
		Day       string `json:"day"`
	} `json:"downloads"`
}

func FetchNPMData(username string) (NpmDashboardData, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	data := NpmDashboardData{Username: username}

	// 1. Search Packages (Max 250)
	// We do NOT filter by publisher, so Bot releases are INCLUDED.
	url := fmt.Sprintf("https://registry.npmjs.org/-/v1/search?text=maintainer:%s&size=250", username)
	resp, err := client.Get(url)
	if err != nil { return data, err }
	defer resp.Body.Close()

	var searchRes npmSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&searchRes); err != nil { return data, err }

	data.TotalPackages = searchRes.Total
	if len(searchRes.Objects) > 0 {
		data.AvatarURL = "https://github.com/" + username + ".png"
	}

	// 2. Prepare for Bulk Download Fetch
	var allPackages []PackageItem
	var packageNames []string
	
	oldestDate := time.Now()
	foundDate := false

	// Heatmap buckets (last 12 months)
	monthBuckets := make([]int, 12)
	now := time.Now()

	for _, obj := range searchRes.Objects {
		// Track Dates for Heatmap & Service Ribbon
		if t, err := time.Parse(time.RFC3339, obj.Package.Date); err == nil {
			if t.Before(oldestDate) { oldestDate = t; foundDate = true }
			
			// Calculate Heatmap (0 = this month, 11 = year ago)
			monthsAgo := int(now.Sub(t).Hours() / 24 / 30)
			if monthsAgo >= 0 && monthsAgo < 12 {
				monthBuckets[11-monthsAgo]++ 
			}
		}

		pkg := PackageItem{Name: obj.Package.Name, Version: "v" + obj.Package.Version}
		allPackages = append(allPackages, pkg)
		packageNames = append(packageNames, obj.Package.Name)
	}
	if foundDate { data.FirstPackageDate = oldestDate }
	data.PublishActivity = monthBuckets

	// 3. BULK DOWNLOADS FETCH (Crucial for Ranking)
	// We chunk requests because URLs can't be infinite length
	chunkSize := 30 
	for i := 0; i < len(packageNames); i += chunkSize {
		end := i + chunkSize
		if end > len(packageNames) { end = len(packageNames) }
		
		subset := packageNames[i:end]
		bulkUrl := fmt.Sprintf("https://api.npmjs.org/downloads/point/last-year/%s", strings.Join(subset, ","))
		
		if resp, err := client.Get(bulkUrl); err == nil {
			var bulkResult map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&bulkResult)
			resp.Body.Close()

			// Map downloads back to allPackages
			for idx := range allPackages {
				name := allPackages[idx].Name
				
				// Logic to extract download count from dynamic JSON response
				if val, ok := bulkResult[name]; ok {
					if vMap, ok := val.(map[string]interface{}); ok {
						if d, ok := vMap["downloads"].(float64); ok {
							allPackages[idx].RawDownloads = int(d)
							data.TotalDownloads += int(d)
						}
					}
				}
				// Handle single-package response format
				if len(subset) == 1 && name == subset[0] {
					if d, ok := bulkResult["downloads"].(float64); ok {
						allPackages[idx].RawDownloads = int(d)
						data.TotalDownloads += int(d)
					}
				}
			}
		}
	}

	// 4. SORT BY POPULARITY (Fixes the "Not Ranked Top" issue)
	sort.Slice(allPackages, func(i, j int) bool {
		return allPackages[i].RawDownloads > allPackages[j].RawDownloads
	})

	// 5. Fill Top Packages List
	for i, p := range allPackages {
		if i >= 3 { break }
		p.Downloads = formatNumber(p.RawDownloads)
		data.TopPackages = append(data.TopPackages, p)
	}

	// 6. REAL GRAPH DATA (Range)
	// Fetch history for the #1 package to create the main graph
	if len(data.TopPackages) > 0 {
		topPkg := data.TopPackages[0].Name
		rangeUrl := fmt.Sprintf("https://api.npmjs.org/downloads/range/last-year/%s", topPkg)
		
		if resp, err := client.Get(rangeUrl); err == nil {
			var rangeRes npmRangeResp
			if err := json.NewDecoder(resp.Body).Decode(&rangeRes); err == nil {
				// Compress 365 days into ~40 points for the SVG
				step := len(rangeRes.Downloads) / 40
				if step < 1 { step = 1 }
				for i := 0; i < len(rangeRes.Downloads); i += step {
					data.DownloadHistory = append(data.DownloadHistory, rangeRes.Downloads[i].Downloads)
				}
			}
			resp.Body.Close()
		}
	}
	
	// Fallback Graph
	if len(data.DownloadHistory) == 0 {
		for i := 0; i < 12; i++ { data.DownloadHistory = append(data.DownloadHistory, 100) }
	}

	// 7. Recent Releases List (Sort by date of metadata)
	// We rely on the search order usually being relevance, but let's re-sort searchRes.Objects by Date
	sort.Slice(searchRes.Objects, func(i, j int) bool {
		return searchRes.Objects[i].Package.Date > searchRes.Objects[j].Package.Date
	})

	for i, obj := range searchRes.Objects {
		if i < 4 {
			data.RecentReleases = append(data.RecentReleases, ActivityItem{
				Action: "Published",
				Repo: obj.Package.Name + " v" + obj.Package.Version,
				Type: "push",
			})
		}
	}

	return data, nil
}

func formatNumber(n int) string {
	if n >= 1000000 { return fmt.Sprintf("%.1fM", float64(n)/1000000.0) }
	if n >= 1000 { return fmt.Sprintf("%.1fk", float64(n)/1000.0) }
	return fmt.Sprintf("%d", n)
}