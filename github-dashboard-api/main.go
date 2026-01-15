package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Println("Error loading .env file")
    }
	if os.Getenv("GITHUB_TOKEN") == "" {
        log.Fatal("❌ CRITICAL: GITHUB_TOKEN is missing from environment!")
    }
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Compress(5))
	r.Use(httprate.Limit(
		100,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	}))

	r.Get("/", documentationHandler)
	r.Get("/api/github-stats", fetcherHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}
	log.Printf("🚀 Server running on port %s", port)
	http.ListenAndServe(":"+port, r)
}

func fetcherHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	timezone := r.URL.Query().Get("timezone")
	format := r.URL.Query().Get("format") 
	if format == "" {
		format = "svg"
	}
	if username == ""  || timezone == "" {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}
	data, err := FetchGitHubData(username, os.Getenv("GITHUB_TOKEN"),resolveTimezone(timezone))
	if err != nil {
		fmt.Println(err)
		return
	}
	var buf bytes.Buffer
		renderer := NewRenderer(&buf)
		renderer.Render(data)
		if format == "png" {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=7200")
			
			if err := renderPNG(w, buf.Bytes()); err != nil {
				log.Printf("PNG Encoding error: %v", err)
				http.Error(w, "Failed to generate PNG", 500)
			}
		} else {
			// Default: Stream the SVG directly
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Cache-Control", "public, max-age=7200")
			w.Write(buf.Bytes())
		}
}

func resolveTimezone(input string) *time.Location {
	if input == "" {
		return time.UTC
	}

	commonZones := map[string]string{
		"IST": "Asia/Kolkata", "UTC": "UTC", "GMT": "Etc/GMT",
		"EST": "America/New_York", "EDT": "America/New_York",
		"CST": "America/Chicago", "CDT": "America/Chicago",
		"MST": "America/Denver", "MDT": "America/Denver",
		"PST": "America/Los_Angeles", "PDT": "America/Los_Angeles",
		"CET": "Europe/Paris", "CEST": "Europe/Paris",
		"BST": "Europe/London", "JST": "Asia/Tokyo",
		"AEST": "Australia/Sydney",
	}

	// Try Abbreviation
	if iana, ok := commonZones[strings.ToUpper(input)]; ok {
		if loc, err := time.LoadLocation(iana); err == nil {
			return loc
		}
	}

	// Try Standard Name (e.g. "Europe/Berlin")
	if loc, err := time.LoadLocation(input); err == nil {
		return loc
	}

	return time.UTC
}
func renderPNG(w http.ResponseWriter, svgData []byte) error {
    // Call the external 'rsvg-convert' tool
    // It reads from Stdin and writes to Stdout
    cmd := exec.Command("rsvg-convert")
    cmd.Stdin = bytes.NewReader(svgData)
    cmd.Stdout = w // Stream output directly to the HTTP response

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("rsvg-convert failed: %w", err)
    }
    return nil
}